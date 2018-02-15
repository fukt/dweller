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

package v1alpha1

import (
	v1alpha1 "github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
	scheme "github.com/fukt/dweller/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// VaultSecretClaimsGetter has a method to return a VaultSecretClaimInterface.
// A group's client should implement this interface.
type VaultSecretClaimsGetter interface {
	VaultSecretClaims(namespace string) VaultSecretClaimInterface
}

// VaultSecretClaimInterface has methods to work with VaultSecretClaim resources.
type VaultSecretClaimInterface interface {
	Create(*v1alpha1.VaultSecretClaim) (*v1alpha1.VaultSecretClaim, error)
	Update(*v1alpha1.VaultSecretClaim) (*v1alpha1.VaultSecretClaim, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.VaultSecretClaim, error)
	List(opts v1.ListOptions) (*v1alpha1.VaultSecretClaimList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VaultSecretClaim, err error)
	VaultSecretClaimExpansion
}

// vaultSecretClaims implements VaultSecretClaimInterface
type vaultSecretClaims struct {
	client rest.Interface
	ns     string
}

// newVaultSecretClaims returns a VaultSecretClaims
func newVaultSecretClaims(c *DwellerV1alpha1Client, namespace string) *vaultSecretClaims {
	return &vaultSecretClaims{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the vaultSecretClaim, and returns the corresponding vaultSecretClaim object, and an error if there is any.
func (c *vaultSecretClaims) Get(name string, options v1.GetOptions) (result *v1alpha1.VaultSecretClaim, err error) {
	result = &v1alpha1.VaultSecretClaim{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("vaultsecretclaims").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of VaultSecretClaims that match those selectors.
func (c *vaultSecretClaims) List(opts v1.ListOptions) (result *v1alpha1.VaultSecretClaimList, err error) {
	result = &v1alpha1.VaultSecretClaimList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("vaultsecretclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested vaultSecretClaims.
func (c *vaultSecretClaims) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("vaultsecretclaims").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a vaultSecretClaim and creates it.  Returns the server's representation of the vaultSecretClaim, and an error, if there is any.
func (c *vaultSecretClaims) Create(vaultSecretClaim *v1alpha1.VaultSecretClaim) (result *v1alpha1.VaultSecretClaim, err error) {
	result = &v1alpha1.VaultSecretClaim{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("vaultsecretclaims").
		Body(vaultSecretClaim).
		Do().
		Into(result)
	return
}

// Update takes the representation of a vaultSecretClaim and updates it. Returns the server's representation of the vaultSecretClaim, and an error, if there is any.
func (c *vaultSecretClaims) Update(vaultSecretClaim *v1alpha1.VaultSecretClaim) (result *v1alpha1.VaultSecretClaim, err error) {
	result = &v1alpha1.VaultSecretClaim{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("vaultsecretclaims").
		Name(vaultSecretClaim.Name).
		Body(vaultSecretClaim).
		Do().
		Into(result)
	return
}

// Delete takes name of the vaultSecretClaim and deletes it. Returns an error if one occurs.
func (c *vaultSecretClaims) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("vaultsecretclaims").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *vaultSecretClaims) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("vaultsecretclaims").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched vaultSecretClaim.
func (c *vaultSecretClaims) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.VaultSecretClaim, err error) {
	result = &v1alpha1.VaultSecretClaim{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("vaultsecretclaims").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
