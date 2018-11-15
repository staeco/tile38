#!/bin/bash -e

export VERSION=1.14.1
export TAG=circleci-${VERSION}
export LDFLAGS="-X github.com/tidwall/tile38/core.Version=${VERSION}"

docker build -f Dockerfile.Custom -t staeco/tile38:${TAG} --build-arg="LDFLAGS=${LDFLAGS}" .

docker push staeco/tile38:${TAG}
