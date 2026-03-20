#!/bin/bash

echo "> Install external-secrets"

helm repo add external-secrets https://charts.external-secrets.io

helm install external-secrets \
   external-secrets/external-secrets \
    -n external-secrets \
    --create-namespace

echo "> external-secrets installed"
