# Getting Started

This project contains the Nightwing Workflow Engine. It uses the [Dagger](https://dagger.io) framework to implement the batCAVE CI/CD pipelines as code that will run on both GitLab and GitHub environments.

## Github Access

You will need access to the https://github.com/nightwing-demo/workflow-engine repository. If you don't already have access you can contact the following Nightwing team members:

- [Bacchus (BJ) Jackson](https://cmsgov.slack.com/team/U02THSF1F3N) ([email](bacchus.jackson@cms.hhs.gov))
- [Danielle (Danni) Smith](https://cmsgov.slack.com/team/U02LU9ECMSM) ([email](danielle.smith@cms.hhs.gov))

to get access.

## Required Tools

The following are required tools for running the Nightwing Workflow Engine

### Go

The Nightwing Workflow Engine is written in [Go](https://go.dev).

To install on a Mac, install using Homebrew:

```
brew install go
```

Optional: if you would like Go built tools to be available locally on the command line, add the following to your `~/.zshrc` or `~/.zprofile` file:

```
# Go
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

#### Recommended Resources

If you are new to Go, or would like a refresher, here are some recommended resources:

- [Go Documentation](https://go.dev/doc/effective_go)
- [101 Go Mistakes and How to Avoid Them](https://www.manning.com/books/100-go-mistakes-and-how-to-avoid-them) - A free an online summarized version can be found [here](https://github.com/teivah/100-go-mistakes)

### Dagger

The Nightwing Workflow Engine uses the [Dagger](https://dagger.io) framework to implement the batCAVE CI/CD pipelines as code.

Prerequisites - To use Dagger, you will need to have the following installed:

- [Docker](https://docs.docker.com/engine/install/)
- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git)

To install Dagger on a Mac using Homebrew:

```
brew install dagger/tap/dagger
```

Then from your existing Go module, run the following commands:

```
go get dagger.io/dagger@latest
go mod init <module name>
go mod tidy
```

Finally, run Dagger:

```
dagger run go run <module name>
```

<b>Example</b>: to run the Nightwing Workflow Engine:

```
git clone https://github.com/nightwing-demo/workflow-engine.git
cd workflow-engine
go get dagger.io/dagger@latest
go mod init main.go
go mod tidy
dagger run go run main.go
```

## Optional Tools

The following are optional tools that may be installed to enhance the developer experience.

### mdbook

[mdbook](https://github.com/rust-lang/mdBook) is written in Rust and requires Rust to be installed as a pre-requisite.

To install Rust on a Mac or other Unix-like OS:

```
curl --proto '=https' --tlsv1.2 -sSf https://sh.rustup.rs | sh
```

If you've installed rustup in the past, you can update your installation by running:

```
rustup update
```

Once you have installed Rust, the following command can be used to build and install mdbook:

```
cargo install mdbook
```

Once mdbook is installed, you can serve it by going to the directory containing the mdbook markdown files and running:

```
mdbook serve
```

### just

[just](https://github.com/casey/just) is "just" a command runner. It is a handy way to save and run project-specific commands.

To install just on a Mac:

You can use the following command on Linux, MacOS, or Windows to download the latest release, just replace `<destination directory>` with the directory where you'd like to put just:

```
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to <destination directory>
```

For example, to install `just` to `~/bin`:

```
# create ~/bin
mkdir -p ~/bin

# download and extract just to ~/bin/just
curl --proto '=https' --tlsv1.2 -sSf https://just.systems/install.sh | bash -s -- --to ~/bin

# add `~/bin` to the paths that your shell searches for executables
# this line should be added to your shell's initialization file,
# e.g. `~/.bashrc` or `~/.zshrc`
export PATH="$PATH:$HOME/bin"

# just should now be executable
just --help
```
