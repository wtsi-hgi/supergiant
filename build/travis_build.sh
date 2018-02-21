#!/bin/bash


## Global Vars
TAG=${TRAVIS_BRANCH:-unstable}
ARCHLIST="linux-amd64 linux-arm64"

## Functions
build_and_push() {
  REPO=$1
  ARCH=$2
  BUILDTAG=$3
  BUILDDIR=$4

  echo "Building and pushing Repository: supergiant/$REPO-$ARCH, TAG: $BUILDTAG"
  cp dist/supergiant-$BUILDDIR-$ARCH build/docker/$BUILDDIR/$ARCH/
  docker build -t supergiant/$REPO-$ARCH:$TAG build/docker/$BUILDDIR/$ARCH/
  docker push supergiant/$REPO-$ARCH
}

build_manifest() {
  REPO=$1
  ARCH=$2
  BUILDTAG=$3
  REPOTAG=$4

  docker manifest create supergiant/$REPO:$BUILDTAG --amend \
  supergiant/$REPO-$ARCH:$REPOTAG

  docker manifest annotate supergiant/$REPO:$BUILDTAG \
  supergiant/$REPO-$ARCH:$REPOTAG --os linux --arch $(echo $ARCH | sed 's/linux-//g')
}

### MAIN
## Build Release
echo "Tag Name: ${TAG}"
if [[ "$TAG" =~ ^v[0-9]. ]]; then
  echo "release"

  docker login -u $DOCKER_USER -p $DOCKER_PASS

###############################
  ## UI Docker Build

  COMP="ui"
  for ARCH in $ARCHLIST; do
    build_and_push "supergiant-${COMP}" $ARCH $TAG $COMP

    build_manifest "supergiant-${COMP}" $ARCH $TAG $TAG
    build_manifest "supergiant-${COMP}" $ARCH 'latest' $TAG
  done

  docker manifest push --purge supergiant/supergiant-$COMP:latest
  docker manifest push --purge supergiant/supergiant-$COMP:$TAG


###############################

  # ## API Docker Build

  COMP="api"
  for ARCH in $ARCHLIST; do
    build_and_push "supergiant-${COMP}" $ARCH $TAG $COMP

    build_manifest "supergiant-${COMP}" $ARCH $TAG $TAG
    build_manifest "supergiant-${COMP}" $ARCH 'latest' $TAG
  done

  docker manifest push --purge supergiant/supergiant-$COMP:latest
  docker manifest push --purge supergiant/supergiant-$COMP:$TAG

###############################

#Packer currently disabled.
##  ./packer build build/build_release.json

elif [[ "$TAG" == "master" ]]; then
    echo "master *Only Test*"
else
  echo "branch unstable"
  docker login -u $DOCKER_USER -p $DOCKER_PASS

  ## UI Docker Build
  COMP="ui"
  for ARCH in $ARCHLIST; do
    build_and_push "supergiant-${COMP}" $ARCH $TAG $COMP
  done

  ## API Docker Build
  COMP="api"
  for ARCH in $ARCHLIST; do
    build_and_push "supergiant-${COMP}" $ARCH $TAG $COMP

    build_manifest "supergiant-${COMP}" $ARCH $TAG $TAG
    build_manifest "supergiant-${COMP}" $ARCH 'latest' $TAG
  done

  ## Maybe a flag to build a test AMI? Just takes a long time...
  # ./packer build build/build_branch.json
fi
