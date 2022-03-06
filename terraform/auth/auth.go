package auth

import (
	"fmt"
	"net/http"

	"github.com/nimbolus/terraform-backend/terraform"
	"github.com/nimbolus/terraform-backend/terraform/auth/basic"
	"github.com/nimbolus/terraform-backend/terraform/auth/jwt"
	"github.com/spf13/viper"
)

type Authenticator interface {
	GetName() string
	Authenticate(secret string, s *terraform.State) (bool, error)
}

func Authenticate(req *http.Request, s *terraform.State) (ok bool, err error) {
	backend, secret, ok := req.BasicAuth()
	if !ok {
		return false, fmt.Errorf("no basic auth header found")
	}

	var authenticator Authenticator
	switch backend {
	case "basic":
		authenticator = basic.NewBasicAuth()
	case "jwt":
		issuerURL := viper.GetString("auth_jwt_oidc_issuer_url")
		if addr := viper.GetString("vault_addr"); issuerURL == "" && addr != "" {
			issuerURL = fmt.Sprintf("%s/v1/identity/oidc", addr)
		} else {
			return false, fmt.Errorf("jwt auth is not enabled")
		}
		authenticator = jwt.NewJWTAuth(issuerURL)
	default:
		err = fmt.Errorf("backend is not implemented")
	}
	if err != nil {
		return false, fmt.Errorf("failed to initialize auth backend %s: %v", backend, err)
	}

	return authenticator.Authenticate(secret, s)
}
