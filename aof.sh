#!/bin/bash -e

sleep 10 # wait for server to start

./tile38-cli readonly yes

# shrink the aof every 5 minutes
# while true; do ./tile38-cli AOFSHRINK; sleep 300; done
