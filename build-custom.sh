#!/bin/bash -e

docker build -f Dockerfile.Custom -t staeco/tile38:alpine .

docker push staeco/tile38:alpine
