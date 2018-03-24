// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type VaultSecretClaim struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec VaultSecretClaimSpec `json:"spec"`
}

// VaultSecretClaimSpec is a specification for vault secret claim.
type VaultSecretClaimSpec struct {
	Secret SecretTemplate `json:"secret"`
}

// SecretTemplate is a template for kubernetes secret created by vault secret
// claim.
type SecretTemplate struct {
	Metadata metav1.ObjectMeta `json:"metadata,omitempty"`
	Data     []DataItem        `json:"data"`
}

// DataItem describes kubernetes secret data key with value requesting from the
// vault.
type DataItem struct {
	Key        string `json:"key"`
	VaultPath  string `json:"vaultPath"`
	VaultField string `json:"vaultField"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VaultSecretClaimList is a list of VaultSecretClaim's.
type VaultSecretClaimList struct {
	metav1.TypeMeta `json:",inline"`
	// Standard list metadata.
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	// Items is the list of Deployments.
	Items []VaultSecretClaim `json:"items"`
}
