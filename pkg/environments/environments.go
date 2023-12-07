package environments

import "dagger.io/dagger"

// Environment wraps the dagger container for convinence
type Envrionment struct {
	image     string
	container *dagger.Container
}

// NewEnvironment creates a new environment in a container to run commands
func NewEnvironment(containerImage string, client *dagger.Client) *Envrionment {
	c := client.Container().From(containerImage)
	return &Envrionment{image: containerImage, container: c}
}
