#!/bin/sh

set -e
set -o pipefail

# This script creates a new branch and tag in the repository.

# parse arguments optional -d indicates dry-run
# version is a positional argument
DRYRUN=false
REMOTE=origin
PUSH=true

# parse arguments
# -v version
# -r remote
# -d dry-run
# -n no push

while getopts "r:v:d" flag; do
case "$flag" in
    d) DRYRUN=true;;
    v) VERSION=$OPTARG;;
    r) REMOTE=$OPTARG;;
    n) PUSH=false;;
esac
done

regex="^[0-9]+\.[0-9]+\.[0-9]+$"

if [[ $VERSION =~ $regex ]]; then
    echo "Valid release version: $VERSION"
    exit 0
else
    echo "Invalid release version: $version. Correct format: <major>.<minor>.<patch>"
    exit 1
fi

#parse version into major, minor, patch
major=$(echo $VERSION | cut -d. -f1)
minor=$(echo $VERSION | cut -d. -f2)
patch=$(echo $VERSION | cut -d. -f3)


BRANCH_NAME=release-$major.$minor
TAG_NAME=$VERSION
if [ "$DRYRUN" = true ]; then
    echo "Running in dry-run mode"
    BRANCH_NAME=release-${major}.${minor}-dryrun
    TAG_NAME=$VERSION-dryrun
fi

# check if relase branch already exists on origin, otherwise create it
echo "Checking if branch $BRANCH_NAME already exists on $REMOTE"
if git ls-remote --exit-code --heads $REMOTE $BRANCH_NAME; then
    echo "Branch $BRANCH_NAME already exists"
else
    echo "Creating branch $BRANCH_NAME"
    git checkout -b $BRANCH_NAME
    if [ "$PUSH" = true ]; then
        echo "Pushing branch $BRANCH_NAME to $REMOTE"
        git push $REMOTE $BRANCH_NAME
    fi
fi

# check if tag already exists on origin, otherwise create it
echo "Checking if tag $TAG_NAME already exists on $REMOTE"
if git ls-remote --exit-code --tags $REMOTE $TAG_NAME; then
    echo "Tag $TAG_NAME already exists"
else
    echo "Creating tag $TAG_NAME"
    git tag $TAG_NAME
    if [ "$PUSH" = true ]; then
        echo "Pushing tag $TAG_NAME to $REMOTE"
        git push $REMOTE $TAG_NAME
    fi
fi
