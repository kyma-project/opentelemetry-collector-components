#!/usr/bin/env bash

# standard bash error handling
set -o nounset  # treat unset variables as an error and exit immediately.
set -o errexit  # exit immediately when a command fails.
set -E          # must be set if you want the ERR trap
set -o pipefail # prevents errors in a pipeline from being masked

# This script has the following arguments:
#                       -  image tag - mandatory
#
# ./check_artifacts_existence.sh 2.1.0


export IMAGE_TAG=$1

PROTOCOL=docker://
IMAGE_NAMES=(
  kyma-otel-collector
)

for image in "${IMAGE_NAMES[@]}"; do
  if [ $(skopeo list-tags ${PROTOCOL}europe-docker.pkg.dev/kyma-project/prod/${image} | jq '.Tags|any(. == env.IMAGE_TAG)') == "true" ]; then
    echo "::WARNING ::${image} binary image for tag ${IMAGE_TAG} already exists"
  else
    echo "No previous ${image} binary image found for tag ${IMAGE_TAG}"
  fi
done
