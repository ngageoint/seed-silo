#!/usr/bin/env bash

set -e

# Ensure script directory is CWD
pushd "${0%/*}" > /dev/null

VERSION=$1
if [[ "${VERSION}x" == "x" ]]
then
    echo Missing version parameter - setting to snapshot
    VERSION=snapshot
fi

REGISTRY=$1
if [[ "${REGISTRY}x" == "x" ]]
then
    echo Missing registry parameter - setting to docker hub
    REGISTRY=docker.io
fi

ORG=$1
if [[ "${ORG}x" == "x" ]]
then
    echo Missing org parameter - setting to geoint
    ORG=geoint
fi

UNAME=$(uname -s)

case "${UNAME}" in
    Linux*)     SUDO=sudo;;
esac

build-silo.sh

${SUDO} docker build --build-arg IMAGE=centos:centos7 . -t $REGISTRY/$ORG/silo:$VERSION
${SUDO} docker push $REGISTRY/$ORG/silo:$VERSION

popd >/dev/null