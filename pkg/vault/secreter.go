package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"
)

// Secreter can fetch secrets using authorized vault client.
type Secreter struct {
	vault *api.Client
	error error
}

// Secret returns secret by path and field.
func (s Secreter) Secret(path, field string) (string, error) {
	if s.error != nil {
		return "", s.error
	}

	vaultSecret, err := s.vault.Logical().Read(path)
	if err != nil {
		return "", err
	}

	if vaultSecret == nil {
		return "", fmt.Errorf("no secret found by path %q", path)
	}

	fieldValue, ok := vaultSecret.Data[field]
	if !ok {
		return "", fmt.Errorf("secret has no field %q", field)
	}

	switch fv := fieldValue.(type) {
	case string:
		return fv, nil
	default:
		return "", fmt.Errorf("unknown type: %T", fieldValue)
	}
}
