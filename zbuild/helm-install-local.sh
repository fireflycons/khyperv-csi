#!/usr/bin/env bash

if [[ -z "${HV_APIKEY}" ]] || [[ -z "${HV_URL}" ]] ; then
    echo "You must export the following"
    echo "export HV_APIKEY=<api key from service installer>"
    echo "export HV_URL=<url from service installer>"
    exit 1
fi

if [[ -z "$VERSION" ]]; then
    echo "VERSION is not set. Are you running this via make?"
    exit 1
fi

NS=hyperv-csi-plugin
RELEASE_NAME=hyperv
CA_ARG=""

if [[ -n "$HV_CA" ]]; then
    CA_ARG="--set-file controller.caCert=$HV_CA"
fi

echo
echo "image.tag=$VERSION"
echo "controller.apiKey=$HV_APIKEY"
echo "controller.serviceUrl=$HV_URL"
if [[ -n "$HV_CA" ]]; then
    echo "controller.caCert=$HV_CA"
fi
echo
helm upgrade $RELEASE_NAME --install --create-namespace --namespace "$NS" $CA_ARG \
    --set image.tag=$VERSION \
    --set controller.apiKey=$HV_APIKEY \
    --set controller.serviceUrl=$HV_URL \
    --set controller.loglevel=5 \
    --set image.pullPolicy=Always \
    ./chart

