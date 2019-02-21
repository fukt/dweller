package vault

import (
	"fmt"

	"github.com/hashicorp/vault/api"

	"github.com/fukt/dweller/pkg/secret"
)

// Loginer can login into Vault.
type Loginer struct {
	vault *api.Client
}

// Login attempts to login to Vault using credentials and
// returns Secreter that can be used to retrieve secrets using
// that credentials.
func (l Loginer) Login(creds *secret.Credentials) Secreter {
	if creds == nil {
		// Just return the client as-is.
		vc, err := l.vault.Clone()
		return Secreter{vault: vc, error: err}
	}

	res, err := l.vault.Logical().
		Write("auth/kubernetes/login", map[string]interface{}{
			"role": creds.Role,
			"jwt":  creds.Token,
		})
	if err != nil {
		return Secreter{vault: nil, error: err}
	}

	if res.Auth == nil {
		err = fmt.Errorf("no authentication information attached to vault res")
		return Secreter{vault: nil, error: err}
	}

	vc, err := l.vault.Clone()
	vc.SetToken(res.Auth.ClientToken)
	return Secreter{vault: vc, error: err}
}
