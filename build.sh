#!/usr/bin/env bash
echo "Building osd in docker"
make docker
docker stop osd-dev
docker rm osd-dev
docker run --name osd-dev -itd osd bash
echo "Copying osd binary to host"
docker cp osd-dev:/bin/osd .
