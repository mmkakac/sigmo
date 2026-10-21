//go:build minimal

package settings

// authModeSupported reports whether an authentication mode is compiled into
// this build. Minimal builds omit the OIDC and Cloudflare Access login paths
// to keep the binary smaller, so those modes cannot be selected or validated.
func authModeSupported(mode string) bool {
	return mode != "oidc" && mode != "cloudflare_access"
}
