package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

// SchemeKind is kind of vault secret claim.
const SchemeKind = "VaultSecretClaim"

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{
	Group:   "dweller.io",
	Version: "v1alpha1",
}

// SchemeGroupVersionKind is group version kind of dweller vault secret claim.
var SchemeGroupVersionKind = schema.GroupVersionKind{
	Group:   "dweller.io",
	Version: "v1alpha1",
	Kind:    SchemeKind,
}

// Kind takes an unqualified kind and returns back a Group qualified GroupKind
func Kind(kind string) schema.GroupKind {
	return SchemeGroupVersion.WithKind(kind).GroupKind()
}

// Resource takes an unqualified resource and returns a Group qualified GroupResource
func Resource(resource string) schema.GroupResource {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}

var (
	Scheme         = runtime.NewScheme()
	Codecs         = serializer.NewCodecFactory(Scheme)
	ParameterCodec = runtime.NewParameterCodec(Scheme)
	CodecFactory   = serializer.NewCodecFactory(Scheme)
	SchemeBuilder  = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme    = SchemeBuilder.AddToScheme
)

// Adds the list of known types to Scheme.
func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&VaultSecretClaim{},
		&VaultSecretClaimList{},
	)
	metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
	return nil
}

func init() {
	addKnownTypes(Scheme)
	metav1.AddToGroupVersion(Scheme, SchemeGroupVersion)
	AddToScheme(Scheme)
}
