# Usage

TODO: delete this doc

## Intro

Let's imagine we are modeling **manufactory**, and currently, we are creating an
application named **workbench**. We need **fabric** for a work, and it is stored
in a **warehouse**, which is implemented by PostgreSQL. Obviously we don't want anyone
to just take our fabric, so to access the warehouse we use a key, in the form of 
PostgreSQL credentials. 

Let's imagine we are modeling **manufactory**. The workload is split among
many **workbenches**. We need **fabric** for a work, and it is stored
in a **warehouse**. Obviously, we don't want anyone to just take our fabric, so 
to access the warehouse we use a key. We store keys in a centralized **safe** 
because it is a robust and reliable solution for our purposes.

The problem is, we have many workbenches with many workers, and each one of them
needs an access to different keys in the safe. Of course, we can give them codes
and teach them the hard process of opening the safe, but do we really want them
to make such efforts every time?

To solve this problem we hire a **key keeper** - the person who controls the safe
and manages access to it. It verifies workers identities and gives them 
corresponding keys. 

Now, let's translate this story to to our software:

- manufactory is a Kubernetes namespace
- workbench is some microservice
- warehouse is PostgreSQL
- key is PostgreSQL credentials
- safe is Hashicorp Vault
- key keeper is Dweller

## Usage

*This example assumes Kubernetes running and Vault integration set up.*

First of all, we put credentials to Vault:

    vault write secret/postgres \
        username=johndoe \
        password=ilovecupcakes

We also attach a policy to it to control access:

    cat <<EOF | vault policy write warehouse-accessor -
    path "secret/postgres" { policy = "read" }
    EOF

Next, we are going to deploy our application **workbench** to Kubernetes into 
**manufactory** namespace. In this example we omit application Kubernetes 
resources like deployments or services, but focus on that we must create a 
`ServiceAccount` for our application:

    cat <<EOF | kubectl create -f -
    apiVersion: v1
    kind: ServiceAccount
    metadata:
      name: strong-worker
      labels:
        app: workbench
    EOF

That's one step closer! Now, as we know our `ServiceAccount` name and a namespace
where the application resides, we need to bind things together by creating 
a role in Vault. This will allow **worker** to delegate authorization, 
so **Dweller** can go to Vault and bring secrets:

    vault write auth/kubernetes/role/worker \
        bound_service_account_names=strong-worker \
        bound_service_account_namespaces=manufactory \
        policies=warehouse-accessor

Lastly, to claim credentials we put into Vault, we deploy `VaultSecretClaim`
resource:

    cat <<EOF | kubectl create -f -
    apiVersion: dweller.io/v1alpha1
    kind: VaultSecretClaim
    metadata:
      name: warehouse-key
      labels:
        app: workbench
    spec:
      serviceAccountName: strong-worker
      vaultRole: worker
      secret:
        metadata:
          labels:
            app: workbench
        data:
        - key: POSTGRES_PASSWORD
          vaultPath: secret/postgres
          vaultField: password
    EOF

Done!

You now have a secret named **warehouse-key** in **manufactory** 
namespace with PostgreSQL credentials taken from Vault and brought by fellow **Dweller**:

    kubectl get secret -o yaml warehouse-key

