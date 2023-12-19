# Troubleshooting

## Mac M1 Docker Container Execution Failure

If you are running on a Mac M1, and are getting an error similar to:

```
ERR execution failure error="input:1: container.from.withEnvVariable.withExec.stdout process \"echo sample output from debug container\" did not complete successfully: exit code: 1\n\nStdout:\n\nStderr:\n"
```

You may need to install [colima](https://github.com/abiosoft/colima).

To install colima on a Mac using Homebrew:

```
brew install colima
```

Start colima:

```
colima start --arch x86_64
```
Then go ahead and run the workflow engine.

## Registry Authentication Issues

If you getting an error connecting to the GitHub container registry [ghcr.io](https://ghcr.io) similar to:

```
ERR execution failure error="input:1: container.from unexpected status from HEAD request to https://ghcr.io/v2/nightwing-demo/omnibus/manifests/v1.0.0: 403 Forbidden\n
```

You will need to login to the GitHub Container Registry as follows.

### Login to GitHub Container Registry

To login to the GitHub Container Registry, you will need to first create a GitHub Personal Access Token ([PAT](https://docs.github.com/en/enterprise-server@3.9/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens)) and use the token to login to the GitHub Container Registry using the following command:

```
docker login ghcr.io
```
Then go ahead and run the workflow engine in the same terminal window.