# Baronial

[![Build Status](https://mstrobel.visualstudio.com/Envelopes/_apis/build/status/Baronial-CI?branchName=master)](https://mstrobel.visualstudio.com/Envelopes/_build/latest?definitionId=7?branchName=master)

Manage your personal finances with all of the power of a scriptable command-line tool, using Baronial!

## Install

### Build from Source

> NOTE: To build from source, you'll need Go 1.12 or greater, perl, and Git. See [CONTRIBUTING.md](./CONTRIBUTING.md) for
more information on setting up your machine to build Baronial. 

_Unix-Based Machines:_

If you're using Linux or a Mac, take advantage of the Makefile that's included in this project. 

``` bash
git clone https://github.com/marstr/baronial.git
cd baronial
make install
```

_Windows Machines:_

It's still easy to build from source on a Windows machine.

``` Batch
git clone https://github.com/marstr/baronial.git
cd baronial
make.bat install
```