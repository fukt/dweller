package secret

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
)

// Credentials represent auth params for secret provider backend.
type Credentials struct {
	Token string
	Role  string
}

// Assembler can assemble kubernetes secret based on VaultSecretClaim.
type Assembler interface {
	// Assemble assembles kubernetes secret based on VaultSecretClaim.
	// Credentials will be used to authenticate to the secret provider backend.
	Assemble(vsc *v1alpha1.VaultSecretClaim, creds *Credentials) (corev1.Secret, error)
}
