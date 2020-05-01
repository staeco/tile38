#!/bin/bash

python_server() {
  python -m SimpleHTTPServer 8081
}

python_server & tile38-server -d /data
