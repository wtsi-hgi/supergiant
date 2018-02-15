#!/bin/bash -x

# Travis deployment script. After test success actions go here.

TAG=${TRAVIS_BRANCH:-unstable}


echo "Tag Name: ${TAG}"
if [[ "$TAG" =~ ^v[0-9]. ]]; then
  echo "release"

  docker login -u $DOCKER_USER -p $DOCKER_PASS
###############################
  ## UI Docker Build
  REPO=supergiant/supergiant-ui
  cp dist/supergiant-ui-linux-amd64 build/docker/ui/linux-amd64/
  cp dist/supergiant-ui-linux-arm64 build/docker/ui/linux-arm64/
  docker build -t $REPO:$TAG-linux-x64 build/docker/ui/linux-amd64/
  docker build -t $REPO:$TAG-linux-arm64 build/docker/ui/linux-arm64/
  docker push $REPO

  ## Multi Arch Release
  docker manifest create $REPO:$TAG \
  $REPO:$TAG-linux-x64 \
  $REPO:$TAG-linux-arm64 \

  docker manifest annotate $REPO:$TAG \
  $REPO:$TAG-linux-x64 --os linux --arch amd64

  docker manifest annotate $REPO:$TAG \
  $REPO:$TAG-linux-arm64 --os linux --arch arm64

  docker manifest push $REPO:$TAG

  ## Multi Arch Latest
  docker manifest create $REPO:latest \
  $REPO:$TAG-linux-x64 \
  $REPO:$TAG-linux-arm64 \

  docker manifest annotate $REPO:latest \
  $REPO:$TAG-linux-x64 --os linux --arch amd64

  docker manifest annotate $REPO:latest \
  $REPO:$TAG-linux-arm64 --os linux --arch arm64

  docker manifest push $REPO:latest

###############################

  ## API Docker Build
  REPO=supergiant/supergiant-api
  cp dist/supergiant-server-linux-amd64 build/docker/api/linux-amd64/
  cp dist/supergiant-server-linux-arm64 build/docker/api/linux-arm64/
  docker build -t $REPO:$TAG-linux-x64 build/docker/api/linux-amd64/
  docker build -t $REPO:$TAG-linux-arm64 build/docker/api/linux-arm64/
  docker push $REPO

  ## Multi Arch Release
  docker manifest create $REPO:$TAG \
  $REPO:$TAG-linux-x64 \
  $REPO:$TAG-linux-arm64 \

  docker manifest annotate $REPO:$TAG \
  $REPO:$TAG-linux-x64 --os linux --arch amd64

  docker manifest annotate $REPO:$TAG \
  $REPO:$TAG-linux-arm64 --os linux --arch arm64

  ## Multi Arch Latest
  docker manifest create $REPO:latest \
  $REPO:$TAG-linux-x64 \
  $REPO:$TAG-linux-arm64 \

  docker manifest annotate $REPO:latest \
  $REPO:$TAG-linux-x64 --os linux --arch amd64

  docker manifest annotate $REPO:latest \
  $REPO:$TAG-linux-arm64 --os linux --arch arm64

   docker manifest push $REPO:$TAG
   docker manifest push $REPO:latest

###############################

#Packer currently disabled.
##  ./packer build build/build_release.json

elif [[ "$TAG" == "master" ]]; then
    echo "master *Only Test*"
else
  echo "branch unstable"
  docker login -u $DOCKER_USER -p $DOCKER_PASS

  ## UI Docker Build
  REPO=supergiant/supergiant-ui
  cp dist/supergiant-ui-linux-amd64 build/docker/ui/linux-amd64/
  cp dist/supergiant-ui-linux-arm64 build/docker/ui/linux-arm64/
  docker build -t $REPO:$TAG-linux-x64 build/docker/ui/linux-amd64/
  docker build -t $REPO:$TAG-linux-arm64 build/docker/ui/linux-arm64/
  docker push $REPO

  ## Multi Arch Release
  docker manifest create $REPO:$TAG \
  $REPO:$TAG-linux-x64 \
  $REPO:$TAG-linux-arm64 \

  docker manifest annotate $REPO:$TAG \
  $REPO:$TAG-linux-x64 --os linux --arch amd64

  docker manifest annotate $REPO:$TAG \
  $REPO:$TAG-linux-arm64 --os linux --arch arm64

  docker manifest push $REPO:$TAG

  ## API Docker Build
  REPO=supergiant/supergiant-api
  cp dist/supergiant-server-linux-amd64 build/docker/api/linux-amd64/
  cp dist/supergiant-server-linux-arm64 build/docker/api/linux-arm64/
  docker build -t $REPO:$TAG-linux-x64 build/docker/api/linux-amd64/
  docker build -t $REPO:$TAG-linux-arm64 build/docker/api/linux-arm64/
  docker push $REPO

  ## Multi Arch Release
  docker manifest create $REPO:$TAG \
  $REPO:$TAG-linux-x64 \
  $REPO:$TAG-linux-arm64 \

  docker manifest annotate $REPO:$TAG \
  $REPO:$TAG-linux-x64 --os linux --arch amd64

  docker manifest annotate $REPO:$TAG \
  $REPO:$TAG-linux-arm64 --os linux --arch arm64

  docker manifest push $REPO:$TAG

  ## Maybe a flag to build a test AMI? Just takes a long time...
  # ./packer build build/build_branch.json
fi
