#!/bin/bash

VERSION=$(git describe --tags --abbrev=0)
export TAG=aof-${BUILD_REV}

docker build -f Dockerfile.Readonly -t gcr.io/stae-product/tile38:${TAG} .
docker push gcr.io/stae-product/tile38:${TAG}
