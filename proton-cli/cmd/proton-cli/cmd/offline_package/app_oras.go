package offline_package

import (
	"context"

	"oras.land/oras-go/v2/registry/remote/auth"
)

func authClient(username, password string) *auth.Client {
	return &auth.Client{
		Credential: func(_ context.Context, _ string) (auth.Credential, error) {
			return auth.Credential{
				Username: username,
				Password: password,
			}, nil
		},
		Cache: auth.NewCache(),
	}
}
