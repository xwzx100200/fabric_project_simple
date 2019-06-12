#!/usr/bin/env bash


#DOCKER_CMD="docker"

FIXTURE_PROJECT_NAME="fabsdkgo"
DOCKER_CMD="${DOCKER_CMD:-docker}"
FIXTURE_PROJECT_NAME="${FIXTURE_PROJECT_NAME:-fabsdkgo}"

DOCKER_REMOVE_ARGS="-f"


CONTAINERS=$($DOCKER_CMD ps -a | grep "${FIXTURE_PROJECT_NAME}-peer.\.org.\.example\.com-" | awk '{print $1}')
IMAGES=$($DOCKER_CMD images | grep "${FIXTURE_PROJECT_NAME}-peer.\.org.\.example\.com-" | awk '{print $1}')

if [ ! -z "$CONTAINERS" ]; then
    if [ "$DOCKER_REMOVE_FORCE" = "true" ]; then
        echo "Stopping chaincode containers created from fixtures ..."
        $DOCKER_CMD stop $CONTAINERS
    fi

    echo "Removing chaincode containers created from fixtures ..."
    $DOCKER_CMD rm $DOCKER_REMOVE_ARGS $CONTAINERS
fi

if [ ! -z "$IMAGES" ]; then
    echo "Removing chaincode images created from fixtures ..."
    $DOCKER_CMD rmi $DOCKER_REMOVE_ARGS $IMAGES
fi

./mys.sh