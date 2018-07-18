#!/bin/bash -e

docker build -f Dockerfile.Custom -t gcr.io/stae-product/tile38:aof-2 .

docker push gcr.io/stae-product/tile38:aof-2
