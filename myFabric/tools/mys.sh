#!/usr/bin/env bash

export ORG1CA1_FABRIC_CA_SERVER_CA_KEYFILE=/etc/hyperledger/fabric-ca-server-config/171b7e701fce182e00bc04b2c58748613fa9a428a5e48b1c39af10978dc6f51e_sk
export ORG2CA1_FABRIC_CA_SERVER_CA_KEYFILE=/etc/hyperledger/fabric-ca-server-config/c796051e93406d0974240971cdf4ce871e6b15c42e681f889f14134801adee0c_sk

export TEST_CHANGED_ONLY=""
export FABRIC_SDKGO_CODELEVEL_VER="v1.4"
export FABRIC_SDKGO_CODELEVEL_TAG="stable"
export FABRIC_DOCKER_REGISTRY=""
export GO_TESTFLAGS=" -failfast"


cd ../fixtures/dockerenv
docker-compose -f docker-compose-std.yaml -f docker-compose.yaml up --remove-orphans --force-recreate --abort-on-container-exit