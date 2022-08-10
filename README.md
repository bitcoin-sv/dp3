# DPP Proxy

[![Release](https://img.shields.io/github/release-pre/bitcoin-sv/dpp-proxy.svg?logo=github&style=flat&v=1)](https://github.com/bitcoin-sv/dpp-proxy/releases)
[![Build Status](https://img.shields.io/github/workflow/status/bitcoin-sv/dpp-proxy/Go?logo=github&v=3)](https://github.com/bitcoin-sv/dpp-proxy/actions)
[![Report](https://goreportcard.com/badge/github.com/bitcoin-sv/dpp-proxy?style=flat&v=1)](https://goreportcard.com/report/github.com/bitcoin-sv/dpp-proxy)
[![Go](https://img.shields.io/github/go-mod/go-version/bitcoin-sv/dpp-proxy?v=1)](https://golang.org/)

The Direct Payment Protocol (DPP - previously BIP270) Proxy, or dpp-proxy, is a basic reference implementation of a Direct Payment Protocol Proxy Server. This proxy service is used in order for wallets that are not directly accessible on the public internet to be able to be contacted through this proxy server.

This is written in go and integrates with a wallet running the Payment Protocol PayD Interface.

## Exploring Endpoints

To explore the endpoints and functionality, navigate to the [Swagger page](https://bitcoin-sv.github.io/dpp-proxy/) where the endpoints and their models are described in detail. You can also access the Swagger endpoint on the server, just run the proxy server using `go run cmd/rest-server/main.go` and hit the [Swagger endpoint](http://localhost:8443/swagger/index.html).

## Configuring dpp-proxy

The server has a series of environment variables that allow you to configure the behaviours and integrations of the server.
Values can also be passed at build time to provide information such as build information, region, version etc.

### Server

| Key                    | Description                                                        | Default        |
| ---------------------- | ------------------------------------------------------------------ | -------------- |
| SERVER_PORT            | Port which this server should use                                  | :8445          |
| SERVER_HOST            | Host name under which this server is found                         | dpp-proxy      |
| SERVER_SWAGGER_ENABLED | If set to true we will expose an endpoint hosting the Swagger docs | true           |
| SERVER_SWAGGER_HOST    | Sets the base url for swagger ui calls                             | localhost:8445 |

### Environment / Deployment Info

| Key                 | Description                                                                | Default          |
| ------------------- | -------------------------------------------------------------------------- | ---------------- |
| ENV_ENVIRONMENT     | What enviornment we are running in, for example 'production'               | dev              |
| ENV_REGION          | Region we are running in, for example 'eu-west-1'                          | local            |
| ENV_COMMIT          | Commit hash for the current build                                          | test             |
| ENV_VERSION         | Semver tag for the current build, for example v1.0.0                       | v0.0.0           |
| ENV_BUILDDATE       | Date the code was build                                                    | Current UTC time |
| ENV_BITCOIN_NETWORK | What bitcoin network we are connecting to (mainnet, testnet, stn, regtest) | regtest          |

### Logging

| Key       | Description                                                           | Default |
| --------- | --------------------------------------------------------------------- | ------- |
| LOG_LEVEL | Level of logging we want within the server (debug, error, warn, info) | info    |

### PayD Wallet

| Key         | Description                                              | Default |
| ----------- | -------------------------------------------------------- | ------- |
| PAYD_HOST   | Host for the wallet we are connecting to                 | payd    |
| PAYD_PORT   | Port the PayD wallet is listening on                     | :8443   |
| PAYD_SECURE | If true the dpp-proxy server will validate the wallet TLS certs | false   |
| PAYD_NOOP   | If true we will use a dummy data store in place of payd  | true    |

## Working with dpp-proxy

There are a set of makefile commands listed under the [Makefile](Makefile) which give some useful shortcuts when working
with the repo.

Some of the more common commands are listed below:

`make pre-commit` - ensures dependencies are up to date and runs linter and unit tests.

`make build-image` - builds a local docker image, useful when testing dpp-proxy in docker.

`make run-compose` - runs dpp-proxy in compose, a reference PayD wallet will be added to compose soon NOTE the above command will need ran first.

### Rebuild on code change

You can also add an optional `docker-compose.dev.yml` file (this is not committed) where you can safely overwrite values or add other services without impacting the main compose file.

If you add this file, you can run it with `make run-compose-dev`.

The file I use has a watcher which means it auto rebuilds the image on code change and ensures compose is always up to date, this full file is shown below:

```yaml
version: "3.7"

services:
  dpp-proxy:
    image: theflyingcodr/go-watcher:1.15.8
    environment:
      GO111MODULE: "on"
      GOFLAGS: "-mod=vendor"
      DB_DSN: "file:data/wallet.db?cache=shared&_foreign_keys=true;"
      DB_SCHEMA_PATH: "data/sqlite/migrations"
    command: watcher -run github.com/bitcoinsv/dpp-proxy/cmd/rest-server/ -watch github.com/bitcoinsv/dpp-proxy
    working_dir: /go/src/github.com/bitcoinsv/dpp-proxy
    volumes:
      - ~/git/bitcoinsv/dpp-proxy-proxy:/go/src/github.com/bitcoinsv/dpp-proxy
```

## CI / CD

We use github actions to test and build the code.

If a new release is required, after your PR is approved and code added to master, simply add a new semver tag and a GitHub action will build and publish your code as well as create a GitHub release.
