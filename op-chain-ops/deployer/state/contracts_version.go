package state

import (
	"fmt"
	"regexp"
)

var (
	gitTagRegex    = regexp.MustCompile(`^op-contracts/v\d+\.\d+\.\d+$`)
	gitCommitRegex = regexp.MustCompile(`^[0-9a-f]{40}$`)
)

type ContractsVersion string

func (c ContractsVersion) Check() error {
	isGitTag := gitTagRegex.MatchString(string(c))
	isGitCommit := gitCommitRegex.MatchString(string(c))
	isLocal := c == "local"
	isLatest := c == "latest"

	if !isGitTag && !isGitCommit && !isLocal && !isLatest {
		return fmt.Errorf("contracts version must be a git tag, git commit, or the literal 'local'")
	}

	return nil
}
