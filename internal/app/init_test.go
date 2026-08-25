package app

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/SkyZonDev/envcheck/internal/core"
	"github.com/SkyZonDev/envcheck/internal/template"
)

func TestInit_DryRunPrintsTemplate(t *testing.T) {
	lastExitCode = core.ExitOK
	stdout, stderr := execInit(t, initOptions{dir: t.TempDir(), dryRun: true, format: "yaml"})

	if lastExitCode != core.ExitOK {
		t.Fatalf("code = %d, attendu 0", lastExitCode)
	}
	if stdout != template.DefaultConfig {
		t.Fatalf("stdout ne correspond pas au modèle")
	}
	if stderr != "" {
		t.Fatalf("stderr devrait être vide, got %q", stderr)
	}
}

func TestInit_DryRunRejectsJSON(t *testing.T) {
	lastExitCode = core.ExitOK
	_, stderr := execInit(t, initOptions{dir: t.TempDir(), dryRun: true, format: "json"})

	if lastExitCode != core.ExitConfigError {
		t.Fatalf("code = %d, attendu %d", lastExitCode, core.ExitConfigError)
	}
	if !strings.Contains(stderr, "yaml") {
		t.Fatalf("stderr = %q, attendu un rappel du format yaml", stderr)
	}
}

func TestInit_WritesFile(t *testing.T) {
	lastExitCode = core.ExitOK
	dir := t.TempDir()
	stdout, _ := execInit(t, initOptions{dir: dir})

	if lastExitCode != core.ExitOK {
		t.Fatalf("code = %d, attendu 0", lastExitCode)
	}
	got, err := os.ReadFile(filepath.Join(dir, "envcheck.yml"))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != template.DefaultConfig {
		t.Fatal("le fichier écrit ne correspond pas au modèle")
	}
	if !strings.Contains(stdout, "envcheck.yml") {
		t.Fatalf("stdout = %q, attendu le chemin écrit", stdout)
	}
}

func TestInit_RefusesExistingFile(t *testing.T) {
	lastExitCode = core.ExitOK
	dir := t.TempDir()
	path := filepath.Join(dir, "envcheck.yml")
	if err := os.WriteFile(path, []byte("deja là\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr := execInit(t, initOptions{dir: dir})
	if lastExitCode != core.ExitConfigError {
		t.Fatalf("code = %d, attendu %d", lastExitCode, core.ExitConfigError)
	}
	if !strings.Contains(stderr, "--force") {
		t.Fatalf("stderr = %q, attendu une mention de --force", stderr)
	}
	got, _ := os.ReadFile(path)
	if string(got) != "deja là\n" {
		t.Fatal("le fichier existant n'aurait pas dû être modifié")
	}
}

func TestInit_ForceOverwrites(t *testing.T) {
	lastExitCode = core.ExitOK
	dir := t.TempDir()
	path := filepath.Join(dir, "envcheck.yml")
	if err := os.WriteFile(path, []byte("ancien\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	execInit(t, initOptions{dir: dir, force: true})
	if lastExitCode != core.ExitOK {
		t.Fatalf("code = %d, attendu 0", lastExitCode)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != template.DefaultConfig {
		t.Fatal("--force aurait dû écraser le contenu")
	}
}

func TestInit_RefusesDotfile(t *testing.T) {
	lastExitCode = core.ExitOK
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, ".envcheck.yml"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	execInit(t, initOptions{dir: dir})
	if lastExitCode != core.ExitConfigError {
		t.Fatalf("code = %d, attendu %d", lastExitCode, core.ExitConfigError)
	}
	if _, err := os.Stat(filepath.Join(dir, "envcheck.yml")); !os.IsNotExist(err) {
		t.Fatal("aucun envcheck.yml n'aurait dû être créé à côté de .envcheck.yml")
	}
}

func execInit(t *testing.T, opts initOptions) (stdout, stderr string) {
	t.Helper()
	cmd := &cobra.Command{}
	var outBuf, errBuf bytes.Buffer
	cmd.SetOut(&outBuf)
	cmd.SetErr(&errBuf)
	if err := runInit(cmd, opts); err != nil {
		t.Fatalf("runInit() erreur inattendue: %v", err)
	}
	return outBuf.String(), errBuf.String()
}
