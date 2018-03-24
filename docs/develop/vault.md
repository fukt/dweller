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

Add some secret for development purposes:

    vault write secret/postgres username=johndoe password=123123

Check it:

    vault read -field password secret/postgres

Now you can [run Dweller](deployment.md) and use the deployed Vault to fetch secrets from it. For
example, the following VaultSecretClaim resource should work with the Vault secret
created above:

    cat <<EOF | kubectl create -f -
    apiVersion: dweller.io/v1alpha1
    kind: VaultSecretClaim
    metadata:
      name: test
      labels:
        app: test
    spec:
      secret:
        metadata:
          labels:
            app: test
        data:
        - key: POSTGRES_PASSWORD
          vaultPath: secret/postgres
          vaultField: password
    EOF

After that you can see Dweller created a kubernetes secret from the claim:

    kubectl get secret -l app=test
