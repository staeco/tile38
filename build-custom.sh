#!/bin/bash -e

export VERSION=1.14.1
export TAG=aof-${BUILD_REV}
export LDFLAGS="-X github.com/tidwall/tile38/core.Version=${VERSION}"

docker build -f Dockerfile.Custom -t gcr.io/stae-product/tile38:${TAG} --build-arg="LDFLAGS=${LDFLAGS}" .

docker push gcr.io/stae-product/tile38:aof-${BUILD_REV}
