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

func RunGatecheckList(dst io.Writer, stdIn io.Reader, stdErr io.Writer, filetype string, dryRunEnabled bool) error {
	return shell.GatecheckCommand(stdIn, dst, stdErr).List(filetype).WithDryRun(dryRunEnabled).Run()
}

func RunGatecheckListAll(dst io.Writer, stdIn io.Reader, stdErr io.Writer, filetype string, dryRunEnabled bool) error {
	return shell.GatecheckCommand(stdIn, dst, stdErr).ListAll(filetype).WithDryRun(dryRunEnabled).Run()
}

func RunGatecheckBundleAdd(bundleFilename string, stdErr io.Writer, dryRunEnabled bool, filenames ...string) error {
	cmd := shell.GatecheckCommand(nil, nil, stdErr)
	for _, filename := range filenames {
		err := cmd.BundleAdd(bundleFilename, filename).WithDryRun(dryRunEnabled).Run()
		if err != nil {
			return err
		}
	}
	return nil
}

// InitGatecheckBundle will encode the config file to JSON and create a new bundle or add it to an existing one
func InitGatecheckBundle(config *Config, stdErr io.Writer, dryRunEnabled bool) error {
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
	gatecheck := shell.GatecheckCommand(nil, nil, stdErr)

	bundleFilename := path.Join(config.ArtifactsDir, config.GatecheckBundleFilename)
	if _, err = os.Stat(bundleFilename); err != nil {
		// The bundle file does not exist
		if errors.Is(err, os.ErrNotExist) {
			return gatecheck.BundleCreate(bundleFilename, tempConfigFilename).WithDryRun(dryRunEnabled).Run()
		}
		return err
	}

	return gatecheck.BundleAdd(bundleFilename, tempConfigFilename).WithDryRun(dryRunEnabled).Run()
}
