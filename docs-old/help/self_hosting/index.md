# Self-Hosted monetr

monetr can be self hosted by running the distributed binaries directly (these can be found on the GitHub Releases page),
or by running the container that is distributed via [Docker Hub](https://hub.docker.com/r/monetr/monetr) or [GitHub
Container Registry](https://github.com/monetr/monetr/pkgs/container/monetr).

## Docker Compose

monetr includes a docker compose yaml file that is an okay starting point for running monetr yourself. If you have
monetr cloned locally, from the project root you can run the following commands to get something simple started up.

```shell title="Shell"
docker compose up
```

This will build the current version of monetr you have cloned locally and start a PostgreSQL and Redis container
alongside it. You'll notice that it will output some warnings about missing Plaid credential variables. At the time of
writing this, monetr only supports Plaid for budgeting; though work is being done to support manual budgeting which will
not require a link at all. If you want to give Plaid a try, please follow the directions in [Third-Party-Credentials -
Plaid](../../developing/credentials.md#Plaid). Once you have credentials, you can pass those as environment variables to
the docker compose up command.

