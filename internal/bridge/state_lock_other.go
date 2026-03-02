//go:build !windows

package bridge

// isLockError on non-Windows platforms always returns false — file locks
// after Chrome exit are a Windows-specific issue.
func isLockError(_ error) bool {
	return false
}
