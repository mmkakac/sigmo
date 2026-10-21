//go:build minimal

package auth

// oidcState is empty in minimal builds: OIDC and Cloudflare Access login state
// is not compiled, which also drops the go-oidc and oauth2 dependencies.
type oidcState struct{}

func newOIDCState() oidcState { return oidcState{} }
