# Contributing

We'd love to accept your contributions to this project! There are just a few guidelines you need to follow.

## Feature Requests

Feature Requests should be opened up as [Issues](/issues/new/) on this repository!

## Issues

[Issues](/issues/new/) are always welcome!

## Pull Requests

**NOTE: We recommend you start by opening a new issue describing the bug or feature you're intending to fix. Even if you think it's relatively minor, it's helpful to know what people are working on.**

We are always welcome to new PRs! You can follow the below guide for learning how you can contribute to the project!

## Getting Started

### Prerequisites

* [Review the commit guide we follow](https://chris.beams.io/posts/git-commit/#seven-rules) - ensure your commits follow our standards
* [Docker](https://docs.docker.com/install/) - building block for local development
* [Docker Compose](https://docs.docker.com/compose/install/) - start up local development
* [Golang](https://golang.org/dl/) - for source code and [dependency management](https://github.com/golang/go/wiki/Modules)
* _optional but recommended_ [Make](https://www.gnu.org/software/make/) - start up local development
* _optional but recommended_ [go-vela/server](https://github.com/go-vela/server) - start up local development

### Setup

* [Fork](/fork) this repository

* Clone this repository to your workstation:

```bash
# Clone the project
git clone git@github.com:go-vela/worker.git $HOME/go-vela/worker
```

* Navigate to the repository code:

```bash
# Change into the project directory
cd $HOME/go-vela/worker
```

* Point the original code at your fork:

```bash
# Add a remote branch pointing to your fork
git remote add fork https://github.com/your_fork/worker
```

### Running Locally

**If you haven't already, please see the [Vela server documentation](https://github.com/go-vela/server/blob/master/.github/DOCS.md) to create the services necessary for executing builds locally.**

* Navigate to the repository code:

```bash
# Change into the project directory
cd $HOME/go-vela/worker
```

* Build the repository code:

```bash
# Build the code with `make`
make build

# Build the code with `go`
GOOS=linux CGO_ENABLED=0 go build -o release/vela-worker github.com/go-vela/worker/cmd/server
```

* Run the repository code:

```bash
# Run the code with `make`
make up

# Run the code with `docker-compose`
docker-compose -f docker-compose.yml up -d --build
```

* For rebuilding the repository code:

```bash
# Rebuild the code with `make`
make rebuild

# Rebuild the code with `docker-compose`
docker-compose -f docker-compose.yml build
```

### Development

**Please see our [local development documentation](DOCS.md) for more information.**

* Navigate to the repository code:

```bash
# Change into the project directory
cd $HOME/go-vela/worker
```

* Write your code and [test locally](#running-locally)
  - Please be sure to [follow our commit rules](https://chris.beams.io/posts/git-commit/#seven-rules)

* Write tests for your changes and ensure they pass:

```bash
# Test the code with `go`
go test ./...
```

* Ensure your code meets the project standards:

```bash
# Clean the code with `make`
make clean

# Clean the code with `go`
go mod tidy
go fmt ./...
go vet ./...
```

* Push to your fork:

```bash
# Push your code up to your fork
git push fork master
```

* Open a pull request. Thank you for your contribution!
