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

BASEIMAGE=$2
if [[ "${BASEIMAGE}x" == "x" ]]
then
    echo Missing base iamge parameter - setting to centos:centos7
    BASEIMAGE=centos:centos7
fi

REGISTRY=$3
if [[ "${REGISTRY}x" == "x" ]]
then
    echo Missing registry parameter - setting to docker hub
    REGISTRY=docker.io
fi

ORG=$4
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

${SUDO} docker build --build-arg IMAGE=$BASEIMAGE . -t $REGISTRY/$ORG/silo:$VERSION
${SUDO} docker push $REGISTRY/$ORG/silo:$VERSION

popd >/dev/null