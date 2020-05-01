#!/bin/bash

export GOOS=linux
export GOARCH=amd64

#!/bin/bash

# Hardcode some values to the core package.
if [ -d ".git" ]; then
	VERSION=$(git describe --tags --abbrev=0)
	GITSHA=$(git rev-parse --short HEAD)
	LDFLAGS="$LDFLAGS -X github.com/tidwall/tile38/core.Version=${VERSION}"
	LDFLAGS="$LDFLAGS -X github.com/tidwall/tile38/core.GitSHA=${GITSHA}"
fi
LDFLAGS="$LDFLAGS -X github.com/tidwall/tile38/core.BuildTime=$(date +%FT%T%z)"

export TAG=readonly-${VERSION}

# Generate the core package
core/gen.sh

# Set final Go environment options
LDFLAGS="$LDFLAGS -extldflags '-static'"
export CGO_ENABLED=0

# Build and store objects into original directory.
go build -ldflags "$LDFLAGS" -o tile38-server cmd/tile38-server/*.go
go build -ldflags "$LDFLAGS" -o tile38-cli cmd/tile38-cli/*.go

docker build -f Dockerfile.Custom -t staeco/tile38:${TAG} .
docker push staeco/tile38:${TAG}
