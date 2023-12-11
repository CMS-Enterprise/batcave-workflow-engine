package environments

import "dagger.io/dagger"

type Environment interface {
	Container() *dagger.Container
}

// Omnibus is a container that contains a host of CI/CD security tools
type Omnibus struct {
	container *dagger.Container
}

// NewOmnibus loads the appropriate image for omnibus
func NewOmnibus(c *dagger.Client, baseImage string) *Omnibus {
	container := c.Container().From(baseImage)
	return &Omnibus{container: container}
}

// Container returns the internal container for the environment
func (o *Omnibus) Container() *dagger.Container {
	return o.container
}

// Alpine is a bare environment, used for debugging
type Alpine struct {
	container *dagger.Container
}

// NewAlpine return an initialized environment
func NewAlpine(c *dagger.Client, baseImage string) *Alpine {
	container := c.Container().From(baseImage)
	return &Alpine{container: container}
}

// Container returns the internal container for the environment
func (a *Alpine) Container() *dagger.Container {
	return a.container
}
