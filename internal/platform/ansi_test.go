package platform_test

import (
	"testing"

	"github.com/SkyZonDev/envcheck/internal/platform"
)

func TestStripANSI_RemovesColorAroundVersion(t *testing.T) {
	raw := "\x1b[32m2.45.1\x1b[0m"
	got := platform.StripANSI(raw)
	if got != "2.45.1" {
		t.Fatalf("StripANSI = %q, attendu %q", got, "2.45.1")
	}
}

func TestStripANSI_LeavesPlainText(t *testing.T) {
	const in = "git version 2.45.1"
	if got := platform.StripANSI(in); got != in {
		t.Fatalf("StripANSI a modifié du texte sans ANSI: %q", got)
	}
}
