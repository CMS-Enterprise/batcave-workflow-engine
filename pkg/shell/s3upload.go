package shell

import (
	"io"
)

type s3uploadCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the S3Upload CLI
// // shell: `s3upload version`
func (s *s3uploadCmd) Version() *Command {
	return NewCommand(s.InitCmd().WithArgs("--version"))
}

// ScanFile runs a S3Upload cmd with provided vars
//
// shell: `s3upload -f ${DOCKERFILE_NAME} -b ${AWS_BUCKET} -k ${S3_KEY}/${DOCKERFILE_NAME}`
func (s *s3uploadCmd) Scan(sourceFile string, bucketName string, s3Key string, destFile string) *Command {
	exe := s.InitCmd().WithArgs(
		"-f",
		sourceFile,
		"-b",
		bucketName,
		"-k",
		s3Key+"/"+destFile,
	)

	return NewCommand(exe)
}

// S3Upload Command with custom stdout and stderr
func S3uploadCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *s3uploadCmd {
	return &s3uploadCmd{
		InitCmd: func() *Executable {
			return NewExecutable("s3upload").WithIO(stdin, stdout, stderr)
		},
	}
}
