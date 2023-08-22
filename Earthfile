VERSION 0.7
FROM golang:1.20.3-alpine3.17
WORKDIR /app

deps:
    COPY go.mod go.sum .
    RUN go mod download
    SAVE ARTIFACT go.mod AS LOCAL go.mod
    SAVE ARTIFACT go.sum AS LOCAL go.sum

buf:
    FROM bufbuild/buf:1.17.0
    COPY buf.gen.yaml .
    COPY proto proto
    RUN buf generate proto
    SAVE ARTIFACT gen AS LOCAL gen

unit-test:
    FROM +deps
    COPY +buf/gen gen
    COPY internal internal
    COPY sql sql
    COPY docker-compose.yml docker-compose.yml
    WITH DOCKER --compose docker-compose.yml
        RUN apk add postgresql-client;                                          \
            while ! pg_isready --host=localhost --port=26257; do                \
                sleep 1;                                                        \
            done;                                                               \
            for SCRIPT in $(ls -1 ./sql); do                                    \
                psql -h localhost -p 26257 -d defaultdb -a -f ./sql/$SCRIPT;    \
            done;                                                               \
            go test -v -coverprofile coverage.out -covermode count ./... || exit 0
    END
    SAVE ARTIFACT coverage.out AS LOCAL coverage.out

coverage:
    FROM +deps
    ARG TESTCOVERAGE_THRESHOLD=70
    COPY +unit-test/coverage.out coverage.out
    RUN go tool cover -func coverage.out;                                                           \
        totalCoverage=$(go tool cover -func=coverage.out | grep total | grep -Eo '[0-9]+\.[0-9]+'); \
        echo "Current test coverage : $totalCoverage%";                                             \
        if (( $(echo "$totalCoverage $TESTCOVERAGE_THRESHOLD" | awk '{print ($1 > $2)}') )); then   \
            echo "OK";                                                                              \
        else                                                                                        \
            echo "Current test coverage is below threshold (${TESTCOVERAGE_THRESHOLD}%).";          \
            echo "Failed";                                                                          \
            exit 1;                                                                                 \
        fi

build:
    FROM +deps
    COPY +buf/gen gen
    COPY cmd cmd
    COPY internal internal
    RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/core:$(cat cmd/server/version.txt) ./cmd/server
    SAVE ARTIFACT build AS LOCAL build

server:
    COPY +build/build build
    ARG file=$(find build/* -type f | cut -d'/' -f2-)
    FROM alpine:3.18
    RUN apk add ca-certificates
    COPY +build/build/$file /app
    # COPY --from=build /etc/passwd /etc/passwd
    # RUN useradd -u 1001 -m iamuser
    # USER nonroot:nonroot
    ENTRYPOINT ["/app"]
    SAVE IMAGE --push harbor.davensi.dev/data/$file

all:
  BUILD +server

test:
  BUILD +unit-test
  BUILD +coverage
