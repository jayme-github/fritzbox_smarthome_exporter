#!/bin/bash
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker buildx build --push --platform linux/amd64,linux/arm64,linux/arm,linux/386 -t "jaymedh/fritzbox_smarthome_exporter:${TRAVIS_TAG}" .
