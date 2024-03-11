package pipelines

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"os"
	"path"
	"workflow-engine/pkg/shell"
)

const (
	DefaultDirMode  = 0o755
	DefaultFileMode = 0o644
)

// OpenOverwrite is a file mode that will:
// 1. Create the file if it doesn't exist
// 2. Overwrite the content if it does
// 3. Write Only
const OpenOverwrite = os.O_CREATE | os.O_WRONLY | os.O_TRUNC

// MakeDirectoryP will create the directory, creating paths if neccessary with sensible defaults
//
// If the Directory already exists, it will return successfully as nil
func MakeDirectoryP(directoryName string) error {
	slog.Debug("make directory", "path", directoryName)
	return os.MkdirAll(directoryName, DefaultDirMode)
}

// OpenOrCreateFile will create the file or overwrite an existing file
func OpenOrCreateFile(filename string) (*os.File, error) {
	slog.Debug("create or open and overwrite existing file", "path", filename)
	return os.OpenFile(filename, OpenOverwrite, DefaultFileMode)
}

// Common Shell Commands to Functions

// InitGatecheckBundle will encode the config file to JSON and create a new bundle or add it to an existing one
//
// The stderr will be suppressed unless there is an non-zero exit code
func InitGatecheckBundle(config *Config, stderr io.Writer, dryRunEnabled bool) error {
	tempConfigFilename := path.Join(os.TempDir(), "wfe-config.json")

	tempFile, err := OpenOrCreateFile(tempConfigFilename)
	if err != nil {
		slog.Error("cannot create temp config file", "error", err)
		return err
	}

	defer func() {
		_ = tempFile.Close()
		_ = os.Remove(tempConfigFilename)
	}()

	if err := json.NewEncoder(tempFile).Encode(config); err != nil {
		slog.Error("cannot encode temp config file", "error", err)
		return err
	}

	bundleFilename := path.Join(config.ArtifactDir, config.GatecheckBundleFilename)
	opts := []shell.OptionFunc{
		shell.WithDryRun(dryRunEnabled),
		shell.WithBundleFile(bundleFilename, tempConfigFilename),
		shell.WithErrorOnly(stderr),
	}
	if _, err = os.Stat(bundleFilename); err != nil {
		// The bundle file does not exist
		if errors.Is(err, os.ErrNotExist) {
			exitCode := shell.GatecheckBundleCreate(opts...)
			return exitCode
		}
		return err
	}

	exitCode := shell.GatecheckBundleAdd(opts...)

	return exitCode
}
