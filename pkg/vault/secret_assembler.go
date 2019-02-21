package vault

import (
	"fmt"

	vault "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/fukt/dweller/pkg/apis/dweller/v1alpha1"
	"github.com/fukt/dweller/pkg/secret"
)

// SecretAssembler assembles kubernetes secrets using Vault as a secret provider.
type SecretAssembler struct {
	vault *vault.Client
}

// Assemble assembles a kubernetes secret from the vault secret claim fetching
// secret values from Vault.
func (asm *SecretAssembler) Assemble(vsc *v1alpha1.VaultSecretClaim, creds *secret.Credentials) (corev1.Secret, error) {
	meta := asm.assembleMeta(vsc)

	secret := corev1.Secret{
		ObjectMeta: meta,
		Type:       corev1.SecretTypeOpaque,
		StringData: make(map[string]string),
	}

	if err := asm.fetchVaultSecrets(vsc.Spec.Secret.Data, &secret, creds); err != nil {
		return secret, err
	}

	return secret, nil
}

func (asm *SecretAssembler) assembleMeta(vsc *v1alpha1.VaultSecretClaim) metav1.ObjectMeta {
	meta := metav1.ObjectMeta{}

	// Force name to be equal to vsc name.
	meta.Name = vsc.Name
	// Force the same namespace.
	meta.Namespace = vsc.Namespace

	// Only labels and annotations are copied from spec.secret.metadata.
	meta.Labels = vsc.Spec.Secret.Metadata.Labels
	meta.Annotations = vsc.Spec.Secret.Metadata.Annotations

	// Set ownership to benefit from garbage collection.
	// See: https://kubernetes.io/docs/concepts/workloads/controllers/garbage-collection/
	ownerRef := metav1.NewControllerRef(vsc, v1alpha1.SchemeGroupVersionKind)
	meta.OwnerReferences = []metav1.OwnerReference{*ownerRef}

	return meta
}

func (asm *SecretAssembler) fetchVaultSecrets(items []v1alpha1.DataItem, secret *corev1.Secret, creds *secret.Credentials) error {
	vc, err := asm.login(creds)
	if err != nil {
		return err
	}

	for _, item := range items {
		vaultSecret, err := vc.Logical().Read(item.VaultPath)
		if err != nil {
			return err
		}

		if vaultSecret == nil {
			return fmt.Errorf("no secret found by path %q", item.VaultPath)
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

func (asm *SecretAssembler) login(creds *secret.Credentials) (*vault.Client, error) {
	if creds == nil {
		// Just return the client as-is.
		return asm.vault.Clone()
	}

	res, err := asm.vault.Logical().
		Write("auth/kubernetes/login", map[string]interface{}{
			"role": creds.Role,
			"jwt":  creds.Token,
		})
	if err != nil {
		return nil, err
	}
	if res.Auth == nil {
		return nil, fmt.Errorf("no authentication information attached to vault res")
	}

	vc, err := asm.vault.Clone()
	if err != nil {
		return nil, err
	}
	vc.SetToken(res.Auth.ClientToken)
	return vc, nil
}

// NewSecretAssembler returns new Vault secret assembler.
func NewSecretAssembler(vault *vault.Client) *SecretAssembler {
	return &SecretAssembler{vault: vault}
}
