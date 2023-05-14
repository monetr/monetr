# Local Development

This guide walks you through setting up a local development environment for monetr on macOS or Linux. If you are using
Windows, it is still possible to run the development environment locally. However, it is not documented at this time.

## Prerequisites

At the time of writing this, monetr requires Plaid credentials for development. Among other credentials, documentation
on how to retrieve them can be found here: [Developing > Credentials](credentials.md)

## Clone the repository

To get started, clone the monetr repository from [GitHub](https://github.com/monetr/monetr) or from your fork.

```shell title="Shell"
git clone https://github.com/monetr/monetr.git
cd monetr
```

The rest of the commands in this guide are issued from this directory.

## Dependencies

monetr's local development tries to not require anything to be installed on the developer's actual computer. Some tools
must be installed though in order to prevent other tools being required. For the most part this is covered by Docker or
another container engine. In the future we plan to fully support Podman as a local container engine as well as Docker,
but at the moment Docker is recommended.

### Mac Specific

macOS can ship with a version of `make` that is outdated. It is recommended that you use `brew` or any other preferred
method to install the most recent version of GNUMake on your Mac. This will not break anything that is already using
make, but will make sure that your version is compatible with the monetr Makefiles.

For example; you should see something like this for your make version.

```shell title="Shell"
make -v
# GNU Make 4.3
# Built for x86_64-apple-darwin20.1.0
# Copyright (C) 1988-2020 Free Software Foundation, Inc.
# License GPLv3+: GNU GPL version 3 or later <http://gnu.org/licenses/gpl.html>
# This is free software: you are free to change and redistribute it.
# There is NO WARRANTY, to the extent permitted by law.
```

### Required

The following tools are required in order to work on monetr locally.

- Docker
- Node (with `yarn` installed)
- GNUMake

These tools can be installed by running the following command.

```shell title="Shell"
brew bundle --verbose
```

But they can also be installed manually through your own preferred methods. As long as `docker`, `make` and `yarn` are
all available in your `PATH` variable, the rest of this guide should work just fine.

## Configuration & Credentials

At the moment monetr requires at least Plaid credentials in order to run properly, even for development. You can read
more about obtaining these credentials here: [Credentials](credentials.md)

The makefile will look for these development credentials and some configuration options in the following path:

```shell title="monetr development env file"
$HOME/.monetr/development.env
```

You can create this file using the following commands:

```shell title="Creating with the helper"
make settings
```

This will create the directory for you and seed the development environment file with a template for you to start with.
It will also open the file created in your `$EDITOR`.

You can also create the file manually like this:

```shell title="Manually creating the development env file"
mkdir $HOME/.monetr
touch $HOME/.monetr/development.env
vim $HOME/.monetr/development.env
```

Once you've opened this file you'll need to provide the Plaid Client ID as `PLAID_CLIENT_ID` and Plaid Client Secret as
`PLAID_CLIENT_SECRET` here.

## Starting It Up

With the above requirements installed. You should be able to spin up the local development environment that runs inside
of Docker compose.

This command will also load any of the environment variables specified in the development env file (mentioned above)
into the `monetr` container where the API is running.

```shell title="Shell"
make develop
```

This will install node modules in the project's directory, as well as start up the containers needed for monetr to run
locally. This command will exit automatically once all the containers are healthy. If you want to follow along while it
is starting up you can use the `make logs` command in another terminal inside the project directory.

## Working

Congratulations, you should now have monetr running locally on your computer. The `develop` task will print out some
basic information for you. But if you missed it, you can access the development version of monetr at `http://monetr.local`.

If you are working on documentation then that can be accessed at `http://monetr.local/documentation`.

Almost all of monetr's code is setup to hot-reload as you make changes. The documentation, Go code and React UI will all
automatically reload as changes are made. Changes to the Go code will not invoke a browser refresh of any sort though,
so to observe a new behavior in the API you will need to refresh or make the API call again.

??? note

    If you want to disable hot reloading of the Go code, you can include `DISABLE_GO_RELOAD=true` in your env variables
    when you run `make develop`.

### Debugging

The monetr container running the API has [delve](https://github.com/go-delve/delve) included. If you prefer to work
using a step-debugger you can connect your editor to it. You will need to reconnect your editor each time it reloads,
but it is very easy to make your changes and then hit _debug_ and let your breakpoints be hit.

Delve is available via port `2345` on `localhost`. I'm not sure what the configuration will be for every editor to
connect to it; but this is a screenshot of IntellJ IDEA's configuration for remote debugging.

![IntellJ IDEA Configuration](assets/IntellJ_IDEA_Go_Debugging.png)

### Running Tests

monetr requires a PostgreSQL instance to be available for Go tests to be run. At the moment there isn't a shorthand
script to provision this instance. But an easy way to do so is this:

```shell title="Shell"
docker run -e POSTGRES_HOST_AUTH_METHOD=trust --name postgres --rm -d -p 5432:5432 postgres:14
```

This will start a PostgreSQL instance in Docker (or remove an existing one) and make it available on `locahost:5432` as
well as not require authentication. This makes it easy for tests to target it.

If tests are run via `make` then nothing more needs to be done. However, if you want to run tests directly from your
editor or other tools you will need to run the database migrations.

```shell title="Shell"
make migrate
```

Will run all schema migrations on the PostgreSQL database on your localhost container.

---

Tests can be run using the `go test` CLI, or all tests can be run using:

```shell title="Shell"
make test
```

### Running Storybook

monetr now provides a storybook setup for working on UI components outside of running the entire application locally. To
start storybook you can run the following command.

```shell title="Shell"
make storybook
```

This will kick off the storybook server and build process. Once it is ready it will open the storybook in your default
browser. You can then make changes to the components in the stories and see the changes in real time without needing to
run the entire application stack locally.

## Cleaning Up

Once you have finished your work and you want to take the local development environment down you have a few options.

### Stopping

Stopping allows you to keep the data you have in your development environment for later and come back to it. This can be
helpful if you are working on a larger feature or bug. But overall it is recommended to completely wipe and re-provision
your development environment each time you start working on something. This keeps the development environment tidy, but
does have the downside of not introducing edge cases that might otherwise be harder to find. You can pause your
development environment by running:

```shell title="Shell"
make stop
```

This will stop all of the Docker containers that were started, but will not remove any of their data. When you are ready
you can start the containers again with the following command.

```shell title="Shell"
make start
```

### Shutting Down Development Environment

If you want to completely shut everything down (1) then you can run the following command:

??? danger
    This will delete all of your local development data, including any Plaid links, expenses, goals, etc...

```shell title="Shell"
make shutdown
```

This removes the Plaid links that are active, takes down the Docker compose containers, removes their volumes.

### Completely Clean up

If you want to completely start fresh you can run the following make task. This will shut down the local development
environment if it is running, but it will also delete any files created or generated during development. This deletes
your `node_modules` folder, any submodules, and generated UI code.

```shell title="Shell"
make clean
```

This should leave the project directory in a state similar to when it was initially cloned.
