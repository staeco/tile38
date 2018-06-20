#!/bin/bash
while ! nc -z tile38-write 9851; do sleep 3; done
./tile38-cli FOLLOW tile38-write 9851
./tile38-cli READONLY yes

./aof.sh
