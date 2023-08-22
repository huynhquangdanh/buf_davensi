## Core

### Dev

Before developing gRPC APIs in Go, we recommend installing the following VS Code extensions:
- [Go](https://marketplace.visualstudio.com/items?itemName=golang.Go)
- [Protobuf 3](https://marketplace.visualstudio.com/items?itemName=zxh404.vscode-proto3)
- [Buf](https://marketplace.visualstudio.com/items?itemName=bufbuild.vscode-buf)

We also recommend installing the following VS Code extensions:
- [EditorConfig](https://marketplace.visualstudio.com/items?itemName=EditorConfig.EditorConfig)
- [Earthfile Syntax Highlighting](https://marketplace.visualstudio.com/items?itemName=earthly.earthfile-syntax-highlighting)
- [GitHub Actions](https://marketplace.visualstudio.com/items?itemName=GitHub.vscode-github-actions)

#### Golangci-lint

Install [golangci-lint](https://golangci-lint.run/usage/install/#local-installation), the default Go linter configured for this repository:
```sh
brew install golangci-lint
```

Run:
```sh
golangci-lint run
```

#### Pre-commit

Install [pre-commit](https://pre-commit.com/):
```sh
brew install pre-commit
```

Set up the git hook scripts:
```sh
pre-commit install
```

Run against all the files:
```sh
pre-commit run --all-files
```

#### Earthly

Install [Earthly](https://earthly.dev/):
```sh
brew install earthly && earthly bootstrap
```

### Working with Protocol Buffers

Update Buf dependencies:
```sh
buf mod update proto
```

Lint your proto files:
```sh
buf lint proto
```

Generate Go packages from proto files:
```sh
buf generate proto
```

Push your modules to the Buf Schema Registry:
```sh
buf registry login
buf push proto
```

### Server

To run the server locally, you need access to a CockroachDB cluster. You can either setup your own local CockroachDB cluster or access our dev cluster using the Teleport local database proxy:
```sh
tsh login --proxy=teleport.davensi.dev --user=[USERNAME]
tsh db login cockroachdb-non-prod --db-user=roach --db-name=demo
tsh proxy db --tunnel --port 26257 cockroachdb-non-prod
```

Run your server locally:
```sh
go run cmd/server/*.go
```

Run your server locally in a Docker container:
```sh
# run docker with DB in localhost network
docker run -p 8080:8080 --env-file [ENVFILE] --network=host [IMAGE]

# The content of .env file is the same key with config.yaml. Such as:
DEBUG=true
APP_NAME=core
COCKROACHDB_HOST=localhost
COCKROACHDB_PORT=26257
COCKROACHDB_USERNAME=roach
COCKROACHDB_DATABASE=reference
COCKROACHDB_TLS_ENABLE=false
COCKROACHDB_TLS_SKIP_VERIFY=true
COCKROACHDB_MAX_CONN=100
APP_ADDRESS_PORT=:8080
```

### Client

You can invoke your APIs  with [Buf Curl](https://buf.build/docs/curl/usage), [gRPCurl](https://github.com/fullstorydev/grpcurl), Curl, HTTPie, etc.

For example:
```sh
buf curl --schema proto/uoms/v1/uoms_service.proto http://localhost:8080/uoms.v1.UoMsService/GetUoM --data '{"id": "8ac04e6c-c307-4da6-b893-d7e320ee92c2"}'
buf curl --schema proto/uoms/v1/uoms_service.proto http://localhost:8080/uoms.v1.UoMsService/GetUoMList --data '{"type": 1}'
```

Or using [HTTPie](https://github.com/httpie/httpie):
```sh
http http://localhost:8080/uoms.v1.UoMsService/CreateUoM type:=70 symbol=TEST icon=test managed_decimals=10 displayed_decimals=2
http http://localhost:8080/uoms.v1.UoMsService/UpdateUoM id="80ef1f8a-adc3-44e3-aa2b-cf56d31d258f" type:=68
http http://localhost:8080/uoms.v1.UoMsService/GetUoM id="8ac04e6c-c307-4da6-b893-d7e320ee92c2"
http http://localhost:8080/uoms.v1.UoMsService/GetUoMList type:=68
http http://localhost:8080/users.v1.UsersService/GetUser user_id=b77f82ce-07af-44e6-a49f-40b0199b5d12
```


### Build Docker images

Install [Earthly](https://earthly.dev/) if you haven't already, then run:
```sh
earthly --push +all
```

Although this is usually done as part of the CI step:
```sh
act --container-architecture linux/amd64 -s HARBOR_USERNAME -s HARBOR_PASSWORD
```

### To do:
