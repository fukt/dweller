# Deployment

## Prerequisites

#### Vault

Dweller requires Vault to work with secrets. See [these instructions](vault.md)
on how to deploy Vault for development purposes.

#### Custom Resource Definition

To be able to deploy VaultSecretClaim resource to kubernetes, you need to create
custom resource definition first:

    kubectl apply -f deployment/custom-resource-definition.yaml

## Run dweller locally (out of kubernetes cluster)

Build the binary:

    make

Configure dweller using environment variables. The convenient way is just to
use [.env.example](../../.env.example) file replacing variables within it:

    cp .env.example .env
    vim .env

Run the binary:

    ./bin/dweller

