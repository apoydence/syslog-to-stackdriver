#!/bin/bash

set -eu

pwd=$PWD
PROJECT_DIR="$(cd "$(dirname "$0")/.."; pwd)"

pushd ./cmd/syslog-to-stackdriver
    gcloud app deploy --quiet
popd
