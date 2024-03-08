package shell

import "io"

type orasCmd struct {
	InitCmd func() *Executable
}

// Version outputs the version of the Grype CLI
//
// shell: `oras version`
func (o *orasCmd) Version() *Command {
	exe := o.InitCmd().WithArgs("version")
	return NewCommand(exe)
}

// PushBundle use the push command to push a gatecheck bundle
// shell: `oras push --disable-path-validation --artifact-type ${BUNDLE_ARTIFACT_TYPE} ${SAST_ARTIFACT_IMAGE} ${GATECHECK_BUNDLE} | tee log.txt`
func (o *orasCmd) PushBundle(artifactImage string, bundleFilename string) *Command {
	exe := o.InitCmd().WithArgs(
		"push",
		"--disable-path-validation",
		"--artifact-type",
		"application/vnd.gatecheckdev.gatecheck.bundle.tar+gzip",
		artifactImage,
		bundleFilename,
	)

	return NewCommand(exe)
}

func OrasCommand(stdin io.Reader, stdout io.Writer, stderr io.Writer) *orasCmd {
	return &orasCmd{
		InitCmd: func() *Executable {
			return NewExecutable("oras").WithIO(stdin, stdout, stderr)
		},
	}
}
