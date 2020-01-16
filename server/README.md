# Unshort.link Server

This is the backend service running [unshort.link](https://unshort.link). You can build & run it yourself for even better privacy

## Building

For using up unshort.link on your own server you need a [working golang installation](https://golang.org/doc/install)

### 1) Generating assets

The assets (html, css, js,...) are directly build into the binary for more portability and an easier usage. You need to 
generate that code by entering `go generate ./...` in the main folder of the project.

### 2) Building
   
Building the project works with `go build` in the main folder of the project. (Please keep in mind that you need to generate
the assets first)

## Or using Make

You can easily build the server by using `make build`

## Or using Docker

### 1) Building

Use `docker build -t unshort .`

Or

Use the command `make dockerize` to build the docker image using make

### 2) Running

Now run the docker container using `docker container run --rm --name unlink -p 80:8080 unshort`

You can any port instead of 80 in the example command above.

For example, you can run the docker container on the host machine's port 8085 using `docker container run -p 8085:8080 unshort`   

## Setup

The building process provides you with an all-inclusive binary. Just enter `./unshort.link` in your console and you should
be up and running

### Available configuration flags

- `--url`: Set the url of the server you are running on (this is only required for the frontend) (Default: `http://localhost:8080`)
- `--port`: Port to start the server on (Default: `8080`)
- `--local`: Use the assets (frontend & blacklist) directly from the filesystem instead of the internal binary storage. This helps during the development of the frontend as you do not have to do `go generate ./...` after every change. This should not be used in production. (Default: `false`)
- `--blacklist-sources`:  Comma separated list of blacklist urls to periodically sync. The blacklist should be a list of newline separated domains (Default https://hosts.ubuntu101.co.za/domains.list)
- `--sync`: Blacklist synchronization interval. The format is a number and a unit. For example `30m` or `1.5h`, available units are `"s", "m", "h"`. Mixed values are also possible: `XhXmXs` (Default: one hour)

## Development
### How to run the unit tests
Just run `go test ./...` from the command line.

Or use `make test` command

### How to clean generated files and db
Just run `make clean` from the command line.