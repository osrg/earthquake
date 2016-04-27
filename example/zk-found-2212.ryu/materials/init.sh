#!/bin/bash

set -e # exit on an error
. ${NMZ_MATERIALS_DIR}/lib.sh

CHECK_PREREQUISITES
FETCH_ZK
BUILD_DOCKER_IMAGE

exit 0
