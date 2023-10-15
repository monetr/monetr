# Build

Building monetr locally using cmake.

## General Requirements

- `Node >= v18.0.0`
- `Go >= 1.19.0`
- `Git`
- `GNUMake`

monetr will also create some binaries within the CMake binary directory, these binaries may be created by `gem`, `npm`
or `go install`. These will not be installed on the host system itself, and are automatically removed when the source
directory is cleaned.

### Linux

If you want to do a release build of monetr on linux, you will also need the following packages:

- `ruby-full`
- `libssl-dev`
- `pkg-config`

These are required in order for monetr to setup [licensed](https://github.com/github/licensed). This tool is setup in
the CMake binary directory and is not installed on the host, but is installed via gem. The tool is used to generate
monetr's third party notice file that is embedded in it at build time. This file contains a list of all of the licenses
of all of monetr's Go and JS dependencies.

## Building monetr

To build monetr you can simply run the following make command. This will run the CMake configuration and build steps
necessary to produce a binary at `$PWD/build/monetr`.

```shell title="Shell"
make monetr
```

### Release Build

In order to produce a release build of monetr the following flag must be added.

```shell title="Shell"
make monetr CMAKE_OPTIONS=-DCMAKE_BUILD_TYPE=Release
```

??? note

  In CI/CD, all builds are performed in release mode.


