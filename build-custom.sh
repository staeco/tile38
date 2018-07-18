#!/bin/bash -e

export BUILD_REV=$(git rev-parse HEAD)

docker build -f Dockerfile.Custom -t gcr.io/stae-product/tile38:aof-${BUILD_REV} .

docker push gcr.io/stae-product/tile38:aof-${BUILD_REV}
