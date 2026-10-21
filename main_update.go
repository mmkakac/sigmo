//go:build !minimal

package main

import (
	appupdate "github.com/damonto/sigmo/internal/app/update"
)

// recoverPendingUpdate restores an interrupted update before the process
// starts, so the freshly installed binary takes over on the next launch.
func recoverPendingUpdate(executable string) (bool, error) {
	return appupdate.RecoverPending(executable)
}
