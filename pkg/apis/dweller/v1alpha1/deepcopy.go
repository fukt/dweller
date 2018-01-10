package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretClaim) DeepCopyInto(out *SecretClaim) {
	*out = *in
	return
}

// DeepCopy is a deepcopy function, copying the receiver, creating a new RawExtension.
func (in *SecretClaim) DeepCopy() *SecretClaim {
	if in == nil {
		return nil
	}
	out := new(SecretClaim)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is a deepcopy function, copying the receiver, creating a new Object.
func (in *SecretClaim) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

func (obj *SecretClaim) GetObjectKind() schema.ObjectKind { return obj }

func (obj *SecretClaim) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}

func (obj *SecretClaim) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}

// DeepCopyInto is a deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SecretClaimList) DeepCopyInto(out *SecretClaimList) {
	*out = *in
	return
}

// DeepCopy is a deepcopy function, copying the receiver, creating a new RawExtension.
func (in *SecretClaimList) DeepCopy() *SecretClaimList {
	if in == nil {
		return nil
	}
	out := new(SecretClaimList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is a deepcopy function, copying the receiver, creating a new Object.
func (in *SecretClaimList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

func (obj *SecretClaimList) GetObjectKind() schema.ObjectKind { return obj }

func (obj *SecretClaimList) SetGroupVersionKind(gvk schema.GroupVersionKind) {
	obj.APIVersion, obj.Kind = gvk.ToAPIVersionAndKind()
}

func (obj *SecretClaimList) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(obj.APIVersion, obj.Kind)
}
