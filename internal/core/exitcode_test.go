package core

import "testing"

func TestExitCodesAreDistinct(t *testing.T) {
	codes := []int{ExitOK, ExitFail, ExitConfigError, ExitInternalError}
	seen := map[int]bool{}
	for _, c := range codes {
		if seen[c] {
			t.Fatalf("code de sortie dupliqué : %d", c)
		}
		seen[c] = true
	}
}
