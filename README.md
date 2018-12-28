# Baronial

Manage your personal finances with all of the power of a scriptable command-line tool, using Baronial!

## Install

### Build from Source

> NOTE: To build from source, you'll need the Go 1.11 or greater and Git. See [CONTRIBUTING.md](./CONTRIBUTING.md) for
more information on setting up your machine to build Baronial. 

_Unix-Based Machines:_

If you're using Linux or a Mac, take advantage of the Makefile that's included in this project. 

``` bash
git clone https://github.com/marstr/baronial.git
cd baronial
make install
```

_Windows Machines:_

It's still easy to build from source on a Windows machine. However, it will be tremendously helpful if you ever need to
file a bug if you inject a little information about which version you installed as you build it.

``` PowerShell
git clone https://github.com/marstr/baronial.git
cd baronial
go install -ldflags "-X github.com/marstr/baronial/cmd.revision=$(git rev-parse HEAD)"
```