#!/bin/bash

docker build -t docker.horner.codes/artifacts/artifacts:latest .
docker push docker.horner.codes/artifacts/artifacts:latest
