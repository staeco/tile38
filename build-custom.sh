#!/bin/bash -e

docker build -f Dockerfile.Custom -t staeco/tile38:circleci .

docker push staeco/tile38:circleci
