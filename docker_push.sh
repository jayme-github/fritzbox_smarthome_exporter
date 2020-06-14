#!/bin/bash
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker push "jaymedh/fritzbox_smarthome_exporter:${TRAVIS_TAG}"
