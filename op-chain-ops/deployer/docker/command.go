package docker

import (
	"archive/tar"
	"context"
	"errors"
	"fmt"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/ethereum/go-ethereum/log"
	"io"
	"os"
	"path/filepath"
)

var (
	ErrNonzeroExit  = errors.New("nonzero exit code")
	ErrFileNotFound = errors.New("file not found in tar")
)

const LinuxAMD64Platform = "linux/amd64"

type Command struct {
	image         string
	platform      string
	containerName string
	cmd           []string
	env           []string
	stdout        io.Writer
	stderr        io.Writer
	mounts        []mount.Mount
	dkr           *client.Client
	lgr           log.Logger
	id            string
	logsClosed    chan struct{}
}

type CommandOpt func(d *Command)

func WithCmd(cmd ...string) CommandOpt {
	return func(d *Command) {
		d.cmd = cmd
	}
}

func WithStdout(w io.Writer) CommandOpt {
	return func(d *Command) {
		d.stdout = w
	}
}

func WithStderr(w io.Writer) CommandOpt {
	return func(d *Command) {
		d.stderr = w
	}
}

func WithEnvVars(vars map[string]string) CommandOpt {
	return func(d *Command) {
		for k, v := range vars {
			d.env = append(d.env, k+"="+v)
		}
	}
}

func WithMount(src, target string) CommandOpt {
	return func(d *Command) {
		d.mounts = append(d.mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: src,
			Target: target,
		})
	}
}

func WithContainerName(name string) CommandOpt {
	return func(d *Command) {
		d.containerName = name
	}
}

func WithImagePlatform(platform string) CommandOpt {
	return func(d *Command) {
		d.platform = platform
	}
}

func NewCommand(lgr log.Logger, image string, opts ...CommandOpt) (*Command, error) {
	dkr, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	d := &Command{
		image:      image,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
		dkr:        dkr,
		lgr:        lgr,
		logsClosed: make(chan struct{}),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d, nil
}

func (d *Command) Run(ctx context.Context) error {
	if err := d.ensureImage(ctx); err != nil {
		return err
	}

	resp, err := d.dkr.ContainerCreate(ctx, &container.Config{
		Image: d.image,
		Env:   d.env,
		Cmd:   d.cmd,
	}, &container.HostConfig{
		Mounts: d.mounts,
	}, nil, nil, d.containerName)
	if err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}

	d.id = resp.ID

	if err := d.dkr.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}

	go d.streamLogs(ctx)

	d.lgr.Info("container started", "id", d.id)

	return d.awaitContainerExit(ctx)
}

func (d *Command) ReadFile(ctx context.Context, path string) ([]byte, error) {
	stream, _, err := d.dkr.CopyFromContainer(ctx, d.id, path)
	if err != nil {
		return nil, fmt.Errorf("failed to copy copy file from container: %w", err)
	}
	defer stream.Close()

	tr := tar.NewReader(stream)
	filename := filepath.Base(path)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil, ErrFileNotFound
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read tar header: %w", err)
		}
		if header.Name == filename {
			return io.ReadAll(tr)
		}
	}
}

func (d *Command) ensureImage(ctx context.Context) error {
	exists, err := d.imageExists(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if image exists: %w", err)
	}

	if !exists {
		if err := d.pullImage(ctx); err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
	}

	return nil
}

func (d *Command) imageExists(ctx context.Context) (bool, error) {
	images, err := d.dkr.ImageList(ctx, image.ListOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to list images: %w", err)
	}

	for _, img := range images {
		for _, tag := range img.RepoTags {
			if tag == d.image {
				d.lgr.Debug("image found in cache", "image", d.image)
				return true, nil
			}
		}
	}

	return false, nil
}

func (d *Command) pullImage(ctx context.Context) error {
	d.lgr.Info("pulling image", "image", d.image)
	reader, err := d.dkr.ImagePull(ctx, d.image, image.PullOptions{
		Platform: d.platform,
	})
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer reader.Close()
	if _, err := io.Copy(os.Stderr, reader); err != nil {
		return fmt.Errorf("failed to copy image pull output: %w", err)
	}
	return nil
}

func (d *Command) streamLogs(ctx context.Context) {
	logs, err := d.dkr.ContainerLogs(ctx, d.id, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	defer close(d.logsClosed)
	if err != nil {
		d.lgr.Error("failed to stream logs", "err", err)
		return
	}
	defer logs.Close()

	_, _ = stdcopy.StdCopy(d.stdout, d.stderr, logs)
}

func (d *Command) awaitContainerExit(ctx context.Context) error {
	statusCh, errCh := d.dkr.ContainerWait(ctx, d.id, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if errors.Is(err, context.Canceled) {
			return err
		}

		<-d.logsClosed

		return fmt.Errorf("error in container: %w", err)
	case <-ctx.Done():
		d.lgr.Info("context cancelled, stopping container", "id", d.id)
		timeout := 0
		if err := d.dkr.ContainerStop(ctx, d.id, container.StopOptions{
			Timeout: &timeout,
		}); err != nil {
			d.lgr.Error("error stopping container", "id", d.id, "err", err)
		}

		<-d.logsClosed

		return ctx.Err()
	case status := <-statusCh:
		if status.StatusCode != 0 {
			return ErrNonzeroExit
		}

		<-d.logsClosed

		return nil
	}
}
