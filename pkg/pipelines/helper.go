package pipelines

import (
	"io"
	"log/slog"
	"os"
	"workflow-engine/pkg/shell"
)

const DefaultDirMode = 0o755
const DefaultFileMode = 0o644

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
