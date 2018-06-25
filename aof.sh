#!/bin/bash -e

sleep 10 # wait for server to start

# shrink the aof every 5 minutes
while true; do ./tile38-cli AOFSHRINK; sleep 300; done
