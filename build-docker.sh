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

UNAME=$(uname -s)

case "${UNAME}" in
    Linux*)     SUDO=sudo;;
esac

./build-silo.sh

${SUDO} docker build . -t silo:$VERSION
${SUDO} docker tag silo:$VERSION docker.platform.cloud.coe.ic.gov/nga-r-dev/silo:$VERSION
${SUDO} docker push./ docker.platform.cloud.coe.ic.gov/nga-r-dev/silo:$VERSION

popd >/dev/null