# dweller
Dweller is a k8s controller which allows easy access to Vault secrets

## deepcopy-gen is creepy-shitty hell

```
go get k8s.io/gengo/examples/deepcopy-gen
go get k8s.io/apimachinery
``

Besides, deepcopy-gen requires these packages to exist:

```
go get golang_org/x/crypto
go get golang_org/x/net
```

And, finally, run magic:
`deepcopy-gen --logtostderr --v=4 -i $(go list ./pkg/apis/dweller/v1alpha1 | paste -sd' ' - | sed 's/ /,/g') -O zz_generated`