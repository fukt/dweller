/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package fake

import (
	v1alpha1 "github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeVaultSecretClaims implements VaultSecretClaimInterface
type FakeVaultSecretClaims struct {
	Fake *FakeDwellerV1alpha1
	ns   string
}

var vaultsecretclaimsResource = schema.GroupVersionResource{Group: "dweller.io", Version: "v1alpha1", Resource: "vaultsecretclaims"}

var vaultsecretclaimsKind = schema.GroupVersionKind{Group: "dweller.io", Version: "v1alpha1", Kind: "VaultSecretClaim"}

// Get takes name of the vaultSecretClaim, and returns the corresponding vaultSecretClaim object, and an error if there is any.
func (c *FakeVaultSecretClaims) Get(name string, options v1.GetOptions) (result *v1alpha1.VaultSecretClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(vaultsecretclaimsResource, c.ns, name), &v1alpha1.VaultSecretClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSecretClaim), err
}

// List takes label and field selectors, and returns the list of VaultSecretClaims that match those selectors.
func (c *FakeVaultSecretClaims) List(opts v1.ListOptions) (result *v1alpha1.VaultSecretClaimList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(vaultsecretclaimsResource, vaultsecretclaimsKind, c.ns, opts), &v1alpha1.VaultSecretClaimList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.VaultSecretClaimList{}
	for _, item := range obj.(*v1alpha1.VaultSecretClaimList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested vaultSecretClaims.
func (c *FakeVaultSecretClaims) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(vaultsecretclaimsResource, c.ns, opts))

}

// Create takes the representation of a vaultSecretClaim and creates it.  Returns the server's representation of the vaultSecretClaim, and an error, if there is any.
func (c *FakeVaultSecretClaims) Create(vaultSecretClaim *v1alpha1.VaultSecretClaim) (result *v1alpha1.VaultSecretClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(vaultsecretclaimsResource, c.ns, vaultSecretClaim), &v1alpha1.VaultSecretClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSecretClaim), err
}

// Update takes the representation of a vaultSecretClaim and updates it. Returns the server's representation of the vaultSecretClaim, and an error, if there is any.
func (c *FakeVaultSecretClaims) Update(vaultSecretClaim *v1alpha1.VaultSecretClaim) (result *v1alpha1.VaultSecretClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(vaultsecretclaimsResource, c.ns, vaultSecretClaim), &v1alpha1.VaultSecretClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSecretClaim), err
}

// Delete takes name of the vaultSecretClaim and deletes it. Returns an error if one occurs.
func (c *FakeVaultSecretClaims) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(vaultsecretclaimsResource, c.ns, name), &v1alpha1.VaultSecretClaim{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeVaultSecretClaims) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(vaultsecretclaimsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.VaultSecretClaimList{})
	return err
}

// Patch applies the patch and returns the patched vaultSecretClaim.
func (c *FakeVaultSecretClaims) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VaultSecretClaim, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(vaultsecretclaimsResource, c.ns, name, data, subresources...), &v1alpha1.VaultSecretClaim{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.VaultSecretClaim), err
}
