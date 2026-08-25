package template_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/template"
)

func TestDefaultConfig_IsValidYAML(t *testing.T) {
	path := filepath.Join(t.TempDir(), "envcheck.yml")
	if err := os.WriteFile(path, []byte(template.DefaultConfig), 0o644); err != nil {
		t.Fatal(err)
	}
	file, err := config.Load(path)
	if err != nil {
		t.Fatalf("le modèle init doit passer validation : %v", err)
	}
	if len(file.Checks) != 4 {
		t.Fatalf("len(Checks) = %d, attendu 4 (un par type MVP)", len(file.Checks))
	}
}

func TestExampleMatchesTemplate(t *testing.T) {
	got, err := os.ReadFile("../../examples/envcheck.yml")
	if err != nil {
		t.Fatal(err)
	}
	// Git sous Windows (core.autocrlf) peut livrer du CRLF ; le modèle Go
	// est toujours en LF, comme envcheck init.
	if strings.ReplaceAll(string(got), "\r\n", "\n") != template.DefaultConfig {
		t.Fatal("examples/envcheck.yml doit rester identique au modèle de envcheck init")
	}
}
