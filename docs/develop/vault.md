# Vault

Dweller requires Vault to work with secrets.

Here is the simple instruction how to run Vault for development
puproses (don't use this for production!).

Run Vault server:

    docker run --rm --cap-add=IPC_LOCK --name=dweller-vault \
        -p 8200:8200 \
        vault:latest

You will see the instructions from the container output containing
Unseal Key and Root Token.

Unseal Vault:

    export VAULT_ADDR='http://0.0.0.0:8200'
    vault operator unseal <Unseal Key>

Export Root Token to use Vault CLI:

    export VAULT_TOKEN=<Root Token>

Enable [Vault auth plugin for Kubernetes](https://github.com/hashicorp/vault-plugin-auth-kubernetes):

*Note that this example is given for minikube but can be adapted for other cluster*

Enable Vault auth plugin for Kubernetes:

    vault auth enable kubernetes

    vault write auth/kubernetes/config \
        kubernetes_host=https://192.168.99.100:8443 \
        kubernetes_ca_cert=@~/.minikube/ca.crt

Now your Vault is ready to work with **dweller**.

###############

Create a role, we use PostgreSQL user role as an example:

    vault write auth/kubernetes/role/postgres_user \
        bound_service_account_names=cooker \
        bound_service_account_namespaces=kitchen \
        policies=postgres-reader


Add some secret for development purposes:

    vault write secret/postgres \
        username=johndoe \
        password=123123
    
    cat <<EOF | vault policy write postgres-reader -
    path "secret/postgres" { policy = "read" }
    EOF

Check it:

    vault read -field password secret/postgres

Now you can [run Dweller](deployment.md) and use the deployed Vault to fetch secrets from it. For
example, the following VaultSecretClaim resource should work with the Vault secret
created above:

    cat <<EOF | kubectl create -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: cooker
    EOF

    cat <<EOF | kubectl create -f -
    apiVersion: dweller.io/v1alpha1
    kind: VaultSecretClaim
    metadata:
      name: cooker
      labels:
        app: cooker
    spec:
      serviceAccountName: cooker
      vaultRole: 
      secret:
        metadata:
          labels:
            app: cooker
        data:
        - key: POSTGRES_PASSWORD
          vaultPath: secret/postgres
          vaultField: password
    EOF

After that you can see Dweller created a kubernetes secret from the claim:

    kubectl get secret -l app=test

## Vault kubernetes integration

*Examples below are given for minikube but can be adapted for other cluster*

Enable Vault auth plugin for Kubernetes:

    vault auth enable kubernetes

    vault write auth/kubernetes/config \
        kubernetes_host=https://192.168.99.100:8443 \
        kubernetes_ca_cert=@~/.minikube/ca.crt

Create a role, we use PostgreSQL user role as an example:

    vault write auth/kubernetes/role/postgres_user \
        bound_service_account_names=cooker \
        bound_service_account_namespaces=kitchen \
        policies=postgres-reader
