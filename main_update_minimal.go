//go:build minimal

package main

// recoverPendingUpdate is a no-op in minimal builds: update functionality is
// not compiled, so there is no pending update to recover.
func recoverPendingUpdate(executable string) (bool, error) {
	return false, nil
}
