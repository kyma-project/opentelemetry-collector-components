#!/usr/bin/env bash

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # must be set if you want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

# This script has the following arguments:
#                       -  image tag - mandatory
#
# ./check_release_tag.sh 2.1.0

# Regular expression to match major.minor.patch format
regex="^[0-9]+\.[0-9]+\.[0-9]+$"

# check if the input string is a valid release version
version="$1"

if [[ $version =~ $regex ]]; then
    echo "Valid release version: $version"
    exit 0
else
    echo "Invalid release version: $version. Correct format: <major>.<minor>.<patch>"
    exit 1
fi
