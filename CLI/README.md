# pipeci CLI application

This folder contains the source code for the CLI application layer written in Golang. pipeci for Go CI/CD.

# Installation

We are using go1.23.4 for development.

```zsh
❯ go version
go version go1.23.4 darwin/amd64
```

Install dependencies.

```
go mod tidy
```

## Structures

Below is a brief description of the CLI application source code

```
├── CLI
│   ├── main.go                 # main
│   ├── Makefile                # Build file
│   ├── cmd                     # CLI source code
│   ├── schema                  # Configuration Schema
│   └── scripts                 # Related scripts
│   ├── LICENSE
```

## Build and Run

`make all` should test, build, and run checks on the source code.

Please see `Makefile` for more details on each part.

Upon make success, an executable file `pipeci` should be installed in the CLI folder. There are two ways to test this.

- Copy `pipeci` to the project root folder.

- Add GOPATH to PATH variable.

Once you have done this step, `pipeci` should work outside of the CLI folder.

Example:

```
❯ which pipeci
pipeci

❯ cd ..

❯ which pipeci
/Users/nguyencanhminh/go/bin/pipeci

❯ pipeci -h
pipeci helps you execute your CI/CD pipelines on both local and remote environments.

Usage:
  pipeci [flags]

Flags:
  -c, --check             Validate the pipeline configuration file.
  -f, --filename string   Path to the pipeline configuration file. (default ".pipelines/pipeline.yaml")
  -h, --help              help for pipeci

```
