#!/bin/bash

set -eu

pwd=$PWD
PROJECT_DIR="$(cd "$(dirname "$0")/.."; pwd)"

app_name="syslog-to-stackdriver"
project_id=""
log_id=""
creds=""

function print_usage {
    echo "Usage: $0 [a:p:l:c:h]"
    echo " -a application name              - The given name (and route) for syslog-to-stackdriver."
    echo " -p project ID (REQUIRED)         - The Google Project ID."
    echo " -c Google Credentials (REQUIRED) - The path to the Google Application Credentials"
    echo " -l Log ID                        - The Log ID (defaults to syslog)."
    echo " -h help                          - Shows this usage."
    echo
    echo "More information available at https://github.com/apoydence/syslog-to-stackdriver"
}

function abs_path {
    case $1 in
        /*) echo $1 ;;
        *) echo $pwd/$1 ;;
    esac
}

function fail {
    echo $1
    exit 1
}

while getopts 'a:p:l:c:h' flag; do
  case "${flag}" in
    a) app_name="${OPTARG}" ;;
    p) project_id="${OPTARG}" ;;
    l) log_id="${OPTARG}" ;;
    c) creds="${OPTARG}" ;;
    h) print_usage ; exit 1 ;;
  esac
done

# Ensure we are starting from the project directory
cd $PROJECT_DIR

if [ -z "$app_name" ]; then
    echo "AppName is required via -a flag"
    print_usage
    exit 1
fi

if [ -z "$project_id" ]; then
    echo "Project ID is required via -p flag"
    print_usage
    exit 1
fi

if [ -z "$creds" ]; then
    echo "Google Application Credentials are required via -c flag"
    print_usage
    exit 1
fi

TEMP_DIR=$(mktemp -d)

# syslog-to-stackdriver binary
echo "building Syslog to Stackdriver binary..."
GOOS=linux go build -o $TEMP_DIR/syslog-to-stackdriver ./cmd/syslog-to-stackdriver &> /dev/null || fail "failed to build syslog-to-stackdriver"
cp $creds $TEMP_DIR
echo "done building Syslog to Stackdriver binary."

echo "pushing $app_name..."
cf push $app_name --no-start -p $TEMP_DIR -b binary_buildpack -c ./syslog-to-stackdriver &> /dev/null || fail "failed to push app $app_name"
echo "done pushing $app_name."

if [ -z ${CF_HOME+x} ]; then
    CF_HOME=$HOME
fi

# Configure
echo "configuring $app_name..."
cf set-env $app_name PROJECT_ID "$project_id" &> /dev/null || fail "failed to set PROJECT_ID"
cf set-env $app_name GOOGLE_APPLICATION_CREDENTIALS "$(basename $creds)" &> /dev/null || fail "failed to set GOOGLE_APPLICATION_CREDENTIALS"
cf set-env $app_name NOT_APP_ENGINE "true" &> /dev/null || fail "failed to set NOT_APP_ENGINE"

if [ ! -z "$log_id" ]; then
    cf set-env $app_name LOG_ID "$log_id" &> /dev/null || fail "failed to set LOG_ID"
fi
echo "done configuring $app_name."

echo "starting $app_name..."
cf start $app_name &> /dev/null || fail "failed to start $app_name"
echo "done starting $app_name."
