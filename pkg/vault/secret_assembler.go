package vault

import (
	"fmt"

	vault "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
)

// SecretAssembler assembles kubernetes secrets using Vault as a secret provider.
type SecretAssembler struct {
	vault *vault.Client
}

// NewSecretAssembler returns new Vault secret assembler.
func NewSecretAssembler(vault *vault.Client) *SecretAssembler {
	return &SecretAssembler{vault: vault}
}

// Assemble assembles a kubernetes secret from the vault secret claim fetching
// secret values from Vault.
func (asm *SecretAssembler) Assemble(vsc *v1alpha1.VaultSecretClaim) (corev1.Secret, error) {
	meta := asm.assembleMeta(vsc)

	secret := corev1.Secret{
		ObjectMeta: meta,
		Type:       corev1.SecretTypeOpaque,
		StringData: make(map[string]string),
	}

	if err := asm.fetchVaultSecrets(vsc.Spec.Secret.Data, &secret); err != nil {
		return secret, err
	}

	return secret, nil
}

func (asm *SecretAssembler) assembleMeta(vsc *v1alpha1.VaultSecretClaim) metav1.ObjectMeta {
	meta := vsc.Spec.Secret.Metadata

	// Force name to be generated after parent name.
	meta.Name = ""
	meta.GenerateName = vsc.ObjectMeta.Name + "-"

	// Set ownership to benefit from garbage collection.
	// See: https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
	ownerRef := metav1.NewControllerRef(vsc, v1alpha1.SchemeGroupVersionKind)
	meta.OwnerReferences = []metav1.OwnerReference{*ownerRef}

	return meta
}

func (asm *SecretAssembler) fetchVaultSecrets(items []v1alpha1.DataItem, secret *corev1.Secret) error {
	for _, item := range items {
		vaultSecret, err := asm.vault.Logical().Read(item.VaultPath)
		if err != nil {
			return err
		}

		var value string
		fieldValue := vaultSecret.Data[item.VaultField]
		switch fv := fieldValue.(type) {
		case string:
			value = fv
		default:
			return fmt.Errorf("unknown type: %T", fieldValue)
		}

		secret.StringData[item.Key] = value
	}

	return nil
}
