#!/usr/bin/env bash
###
### maintenance — turns on/off the maintenance mode
###
### Usage:
###   NAMESPACE="registry-namespace"
###   maintenance on|off
###
### Environment variables:
###   NAMESPACE Kubernetes namespace where the registry is deployed.
###
### Options:
###   on|off    Turns maintenance mode on or off.
###   -h        Show this message.

# Shows help based on the comment at the top of the file.
#
# return    help message
help() {
    awk -F'### ' '/^###/ { print $2 }' "$0"
}

if [[ $# == 0 || "$1" == "-h" ]]; then
    help
    exit 1
fi

# Check input args.
if [[ "$1" != "on" && "$1" != "off" ]]; then
    help
    exit 1
fi

# Check env vars.
if [[ "$NAMESPACE" == "" ]]; then
    help
    exit 1
fi


WAIT_TIMEOUT="120s"
POD_SELECTOR="app=registry"


# Replaces env variable value to turn on maintenance mode.
#
# $1        manifest_path   path to registry deployment manifest
# $2        line_search     search string to find the first line containing the env var
# $3        search          search string to find the env variable and its value (may be multiline)
# $4        replace         string, which should replace the $search string
# return    manifest with the env variable replaced
replace_env_var() {
    local manifest_path="$1"
    local line_search="$2"
    local search="$3"
    local replace="$4"

    sed '/'"$line_search"'/{
        $!{ N
            s/'"$search"'/'"$replace"'/
            t sub-yes
            :sub-not
            P
            D
            :sub-yes
        }
    }' "$manifest_path"
}

# Waits for deployment.
#
# $1    namespace   kubernetes namespace to watch in
# $2    timeout     maximum time to wait before exiting with error
# $3    selector    expression to select pods to watch for
wait_for_upgrade() {
    local namespace="$1"
    local timeout="$2"
    local selector="$3"
    kubectl wait --namespace="$NAMESPACE" --for=condition=ready --timeout "$WAIT_TIMEOUT" pod -l "$selector"
}


# Turn maintenance on.
if [[ "$1" == "on" ]]; then
    manifest=$(
        replace_env_var \
            "k8s/registry-deployment.yaml" \
            'REGISTRY_STORAGE_MAINTENANCE_READONLY' \
            'REGISTRY_STORAGE_MAINTENANCE_READONLY\n          value: "{\\"enabled\\": false}"' \
            'REGISTRY_STORAGE_MAINTENANCE_READONLY\n          value: "{\\"enabled\\": true}"'
        )
    echo "$manifest" | kubectl apply -f -
    wait_for_upgrade "$NAMESPACE" "$WAIT_TIMEOUT" "$POD_SELECTOR"
fi

# Turn maintenance off.
if [[ "$1" == "off" ]]; then
    kubectl apply -f k8s/registry-deployment.yaml
    wait_for_upgrade "$NAMESPACE" "$WAIT_TIMEOUT" "$POD_SELECTOR"
fi
