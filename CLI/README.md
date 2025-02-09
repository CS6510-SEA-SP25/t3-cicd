# GoCC CLI application

This folder contains the source code for the CLI application layer written in Golang. GoCC for Go CI/CD.

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

Upon make success, an executable file `gocc` should be installed in the CLI folder. There are two ways to test this.

- Copy `gocc` to the project root folder.

- Add GOPATH to PATH variable.

Once you have done this step, `gocc` should work outside of the CLI folder.

Example:

```
❯ which gocc
gocc

❯ cd ..

❯ which gocc
/Users/nguyencanhminh/go/bin/gocc

❯ gocc -h
GoCC helps you execute your CI/CD pipelines on both local and remote environments.

Usage:
  gocc [flags]

Flags:
  -c, --check             Validate the pipeline configuration file.
  -f, --filename string   Path to the pipeline configuration file. (default ".pipelines/pipeline.yaml")
  -h, --help              help for gocc

```
