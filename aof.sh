#!/bin/bash -e

# shrink the aof every 5 minutes
while true; do ./tile38-cli AOFSHRINK; sleep 3600; done
