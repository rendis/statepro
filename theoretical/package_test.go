package theoretical_test

import "testing"

// Keeps `go test -cover` / Gremlins coverage gathering healthy for this
// model-only package (no runtime logic to unit-test here).
func TestPackageCompiles(t *testing.T) {
	t.Helper()
}
