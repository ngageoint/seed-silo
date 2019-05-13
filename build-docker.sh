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

CENTOS_IMAGE=$2
if [[ "${CENTOS_IMAGE}x" == "x" ]]
then
    echo Missing centos image parameter - setting to centos:centos7
    CENTOS_IMAGE=centos:centos7
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

CERT_PATH=$5
if [[ "${CERT_PATH}x" == "x" ]]
then
    echo Missing cert path parameter, no certs will be added to image
fi

UNAME=$(uname -s)

case "${UNAME}" in
    Linux*)     SUDO=sudo;;
esac

./build-silo.sh

${SUDO} docker build --build-arg IMAGE=$CENTOS_IMAGE --build-arg CERT_PATH=$CERT_PATH . -t $REGISTRY/$ORG/silo:$VERSION
${SUDO} docker push $REGISTRY/$ORG/silo:$VERSION

popd >/dev/null
