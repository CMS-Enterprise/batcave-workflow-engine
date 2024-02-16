# Image Build

## Command Parameters

### Build Directory

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--build-dir`        | `WFE_BUILD_DIR`           | `image.buildDir`             |

The directory from which to build the container (typically, but not always, the directory where the Dockerfile is located). This parameter is optional, expects a string value, and defaults to the current working directory.

### Dockerfile

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--dockerfile`       | `WFE_BUILD_DOCKERFILE`    | `image.buildDockerfile`      |

### Build Args

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--build-arg`        | `WFE_BUILD_ARGS`          | `image.buildArgs`            |

Defines [build arguments](https://docs.docker.com/build/guide/build-args/) that are passed to the actual container image build command. This parameter is optional, and expects a mapping of string keys to string values, the exact format of which depends on the medium by which it is specified.

#### CLI Flag

The `--build-arg` flag can be specified multiple times to specify different args. The key and value for each arg should be specified as a string in the format `key=value`.

#### Environment Variable

The `WFE_BUILD_ARGS` environment variable must contain all the build arguments in a JSON formatted object (i.e. `{"key":"value"}`).

#### Configuration File

Similar to how build args are specified as an environment variable, build args in config files must be specified as a JSON formatted object. The following is an example YAML config file:

```yaml
image:
  buildArgs: |-
    { "key": "value" }
```

Note that when specifying build args via the configuration file, special care must be taken to ensure that the case of the key is preserved. In the above example the value of `buildArgs` is a string, not a YAML object. When using a JSON config file this would need to be specified as follows:

```json
{
	"image": {
		"buildArgs": "{ \"key\": \"value\" }"
	}
}
```

This is because the workflow-engine configuration file loader does not preserve the case of keys, and build args in Dockerfiles are case sensitive.

### Tag

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--tag`              | `WFE_BUILD_TAG`           | `image.buildTag`             |

### Platform

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--platform`         | `WFE_BUILD_PLATFORM`      | `image.buildPlatform`        |

### Target

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--target`           | `WFE_BUILD_TARGET`        | `image.buildTarget`          |

For [multi-stage Dockerfiles](https://docs.docker.com/build/building/multi-stage/) this parameter specifies a named stage to build.

### Cache To

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--cache-to`         | `WFE_BUILD_CACHE_TO`      | `image.buildCacheTo`         |

### Cache From

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--cache-from`       | `WFE_BUILD_CACHE_FROM`    | `image.buildCacheFrom`       |

### Squash Layers

| CLI Flag             | Variable Name             | Config Field Name            |
|----------------------|---------------------------|------------------------------|
| `--squash-layers`    | `WFE_BUILD_SQUASH_LAYERS` | `image.buildSquashLayers`    |
