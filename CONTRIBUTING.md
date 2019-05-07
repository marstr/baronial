# Contributing to Baronial

This is a piece of Free Software, and is licensed to you under the [GNU Public License v3](./LICENSE). If this is your
first time working on or with Free Software, welcome! While there are more details, Free Software basically means that
you have the following rights:
1. You may modify this software to suit your needs. If you do modify it, please indicate that you have done so.
1. This software can be redistributed to whomever you would like, on the condition that you only charge for the medium
of transfer, or your work to redistribute it. You are not allowed to charge for the product itself, or any derived 
works. For example, if you give someone a USB drive with this software, you can charge them for the drive, but not this
product.

## Setting up your machine

Regardless of the operating system you're using, you'll need the following tools to contribute back to Baronial:
1. [The Go Programming Language, version 1.12 or higher](https://golang.org/dl), for compilation.
1. [Git](https://git-scm.org), for source control.
1. A text editor/IDE of your choice. Some popular options for working with Go include:
	- [VS Code by Microsoft](https://code.visualstudio.com)
	- [GoLand by JetBrains](https://www.jetbrains.com/go/)
	- [Vim](https://www.vim.org)
1. Perl 5, for executing platform independent build scripts. 
1. [golint](https://github.com/golang/lint), for style conformance.

Optionally, you may also want to install:
- [docker](https://www.docker.com/get-started), for testing and building Linux packages locally, even on Mac and Windows.

## Running the Tests

Tests are important! Make sure when you modify this program, that all of the tests execute without failure. This is
is always true, but will be strictly enforced when considering accepting Pull Requests. If you're making a contribution,
please make sure to add tests of your own.

_Unix-like Machines:_

``` bash
make lint
make test
``` 

_Windows Machines:_
``` PowerShell
go fmt ./...
go test ./...
go vet ./...
golint ./...
```

