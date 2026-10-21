//go:build !minimal

package settings

// authModeSupported reports whether an authentication mode is compiled into
// this build. In the default build every mode is available.
func authModeSupported(mode string) bool {
	return true
}
