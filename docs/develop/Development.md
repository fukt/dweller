# Development

## Run dweller locally

Build the binary:

    make build

Run the binary specifying a path to your kubernetes cluster config file:

    KUBECONFIG=$HOME/.kube/minikube ./bin/dweller
