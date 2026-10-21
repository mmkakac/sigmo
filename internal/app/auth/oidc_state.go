//go:build !minimal

package auth

import (
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

// oidcState holds OIDC and Cloudflare Access login state. It is compiled only
// into non-minimal builds so minimal binaries do not link the OIDC and OAuth2
// dependencies.
type oidcState struct {
	logins    map[string]login
	providers map[string]*oidc.Provider
	keys      map[string]*oidc.RemoteKeySet
	client    *http.Client
}

func newOIDCState() oidcState {
	return oidcState{
		logins:    map[string]login{},
		providers: map[string]*oidc.Provider{},
		keys:      map[string]*oidc.RemoteKeySet{},
		client: &http.Client{Timeout: 15 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		}},
	}
}
