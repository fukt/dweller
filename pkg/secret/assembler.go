package secret

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
)

// Assembler can assemble kubernetes secret based on VaultSecretClaim.
type Assembler interface {
	// Assemble assembles kubernetes secret based on VaultSecretClaim.
	Assemble(vsc *v1alpha1.VaultSecretClaim) (corev1.Secret, error)
}
