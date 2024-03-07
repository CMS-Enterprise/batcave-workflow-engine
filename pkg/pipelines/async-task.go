package pipelines

import (
	"errors"
	"io"
	"log/slog"

	"github.com/lmittmann/tint"
)

// AsyncTask is a simple object for running tasks in the background.
//
// The owner of the task MUST close the stdErrPipeWriter or the Wait function
// will never stop waiting. Best method for this is to defer stdErrPipeWriter.Close()
// at the top of the function scope.
// The owner should use the internal logger in a go routine to prevent stderr interweaving
//
// When ready, the caller can use the Wait function to block until this tasks reader has closed.
//
// Any errors during should be directly set to exitError.
type AsyncTask struct {
	taskName         string
	stdErrPipeReader *io.PipeReader
	stdErrPipeWriter *io.PipeWriter
	logger           *slog.Logger
	exitError        error
}

func NewAsyncTask(name string) *AsyncTask {
	pr, pw := io.Pipe()
	return &AsyncTask{
		taskName:         name,
		logger:           slog.New(tint.NewHandler(pw, &tint.Options{Level: slog.LevelDebug})),
		stdErrPipeReader: pr,
		stdErrPipeWriter: pw,
	}
}

// Wait enables stderr streaming until task is complete
//
// The returned error is set by the task along with any possible read/write errors
func (t *AsyncTask) Wait(stderrWriter io.Writer) error {
	_, err := io.Copy(stderrWriter, t.stdErrPipeReader)
	return errors.Join(err, t.exitError)
}
