#!/usr/bin/env bash
###
### garbage-collector — runs built-in garbage collector to physically remove unlinked blobs
###
### Usage:
###   NAMESPACE="registry-namespace"
###   gargage-collector
###
### Environment variables:
###   NAMESPACE Kubernetes namespace where the registry is deployed.
###
### Options:
###   -h        Show this message.

# Shows help based on the comment at the top of the file.
#
# return    help message
help() {
    awk -F'### ' '/^###/ { print $2 }' "$0"
}

if [[ "$1" == "-h" ]]; then
    help
    exit 1
fi

# Check env vars.
if [[ "$NAMESPACE" == "" ]]; then
    help
    exit 1
fi


WAIT_TIMEOUT="1h"
JOB_NAME_PREFIX="registry-janitor-jenkins"

job_name="$JOB_NAME_PREFIX-$(date +%s)"
kubectl create job --namespace="$NAMESPACE" --from=cronjob/registry-janitor "$job_name"
kubectl wait --namespace="$NAMESPACE" --for=condition=complete --timeout "$WAIT_TIMEOUT" job -l "job-name=$job_name"
