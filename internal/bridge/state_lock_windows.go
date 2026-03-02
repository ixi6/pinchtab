//go:build windows

package bridge

import (
	"errors"
	"syscall"
)

// Windows error codes for file locking.
const (
	errSharingViolation syscall.Errno = 32 // ERROR_SHARING_VIOLATION
	errLockViolation    syscall.Errno = 33 // ERROR_LOCK_VIOLATION
	errAccessDenied     syscall.Errno = 5  // ERROR_ACCESS_DENIED
)

// isLockError reports whether err is a Windows file-lock error worth retrying.
func isLockError(err error) bool {
	var errno syscall.Errno
	if errors.As(err, &errno) {
		return errno == errSharingViolation || errno == errLockViolation || errno == errAccessDenied
	}
	return false
}
