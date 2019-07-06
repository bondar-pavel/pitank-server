#!/bin/bash

IMAGE_NAME=pitank-server
REGISTRY=pbondar
VERSION=0.1.0

docker rm -f ${IMAGE_NAME}
docker run -d -p 80:80 --name ${IMAGE_NAME} ${REGISTRY}/${IMAGE_NAME}:${VERSION}