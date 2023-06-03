#!/usr/bin/env bash

# build-image.sh builds the gomarkwiki/builder image.
#
# Usage:
#   build-image.sh [go_version]
#
# Example:
#   build-image.sh 1.20.4

# Exit on error.
set -eu

# Load lib.
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd "$SCRIPT_DIR"
source ./lib

export USAGE="build-image.sh [go_version]"

# Read arguments.
[ $# -eq 1 ] || printUsageAndExit
GO_VERSION="$1"
echo "Go version: ${GO_VERSION}"

# Trace.
set -x

# Do a signature check on the base image.
export DOCKER_CONTENT_TRUST=1

# Build
docker build --build-arg GO_VERSION=${GO_VERSION} --pull -t $IMAGE .

