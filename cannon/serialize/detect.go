package serialize

import (
	"os"
	"strings"

	"github.com/ethereum-optimism/optimism/op-service/jsonutil"
)

func Load[X any](inputPath string) (*X, error) {
	if IsBinaryFile(inputPath) {
		return LoadSerializedBinary[X](inputPath)
	}
	return jsonutil.LoadJSON[X](inputPath)
}

func Write[X Serializable](outputPath string, x X, perm os.FileMode) error {
	if IsBinaryFile(outputPath) {
		return WriteSerializedBinary(outputPath, x, perm)
	}
	return jsonutil.WriteJSON[X](outputPath, x, perm)
}

func IsBinaryFile(path string) bool {
	return strings.HasSuffix(path, ".bin") || strings.HasSuffix(path, ".bin.gz")
}
