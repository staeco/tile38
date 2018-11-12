#!/bin/bash -e

export TAG=circleci-1.14.0

docker build -f Dockerfile.Custom -t staeco/tile38:${TAG} .

docker push staeco/tile38:${TAG}
