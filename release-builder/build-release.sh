#!/usr/bin/env bash

# build-release.sh does a release build of gomarkwiki using the gomarkwiki/builder image.
#
# Usage:
#   build-release.sh [package_dir] [output_dir] [commit]
#
# Preconditions:
#   - Git repo is on the main branch and all files have been checked in.

# Exit on error or use of an undefined var.
set -eu

# Load lib.
SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
cd "$SCRIPT_DIR"
source ./lib

export USAGE="build-release.sh [package_dir] [output_dir] [commit]"

# Read arguments.
[ $# -eq 3 ] || printUsageAndExit
PACKAGE_DIR="$1"
[ -d "$PACKAGE_DIR" ] || errExit "Package directory $PACKAGE_DIR not found"
echo "Package directory: ${PACKAGE_DIR}"
OUTPUT_DIR="$2"
[ -d "$OUTPUT_DIR" ] || errExit "Output directory $OUTPUT_DIR not found"
echo "Output directory: ${OUTPUT_DIR}"
GOMARKWIKI_COMMIT="$3"

# Run
cd "${SCRIPT_DIR}/.."
docker run --rm \
    -v "$PACKAGE_DIR":/gomarkwiki \
    -v "$OUTPUT_DIR":/output \
    --cpuset-cpus=0-$(($(nproc) - 1)) \
    $IMAGE \
    go run ./release-builder/main.go ${GOMARKWIKI_COMMIT}
