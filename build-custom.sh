#!/bin/bash -e

docker build -f Dockerfile.Custom -t staeco/tile38:aof-1 .

docker push staeco/tile38:aof-1
