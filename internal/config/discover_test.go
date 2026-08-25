package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/config"
)

func TestDiscover_PrefersEnvcheckYml(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, "envcheck.yml"), "version: 1")
	write(t, filepath.Join(dir, ".envcheck.yml"), "version: 1")

	got, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover() erreur inattendue: %v", err)
	}
	want := filepath.Join(dir, "envcheck.yml")
	if got != want {
		t.Fatalf("Discover() = %q, attendu %q", got, want)
	}
}

func TestDiscover_FallsBackToDotfile(t *testing.T) {
	dir := t.TempDir()
	write(t, filepath.Join(dir, ".envcheck.yml"), "version: 1")

	got, err := config.Discover(dir)
	if err != nil {
		t.Fatalf("Discover() erreur inattendue: %v", err)
	}
	want := filepath.Join(dir, ".envcheck.yml")
	if got != want {
		t.Fatalf("Discover() = %q, attendu %q", got, want)
	}
}

func TestDiscover_NotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := config.Discover(dir)
	if err != config.ErrNotFound {
		t.Fatalf("Discover() erreur = %v, attendu config.ErrNotFound", err)
	}
}

func TestDiscover_DoesNotSearchParentDirectories(t *testing.T) {
	parent := t.TempDir()
	write(t, filepath.Join(parent, "envcheck.yml"), "version: 1")

	child := filepath.Join(parent, "sous-dossier")
	if err := os.Mkdir(child, 0o755); err != nil {
		t.Fatalf("création du sous-dossier: %v", err)
	}

	_, err := config.Discover(child)
	if err != config.ErrNotFound {
		t.Fatalf("Discover() n'aurait pas dû trouver le fichier du parent, erreur = %v", err)
	}
}

func write(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("écriture de %s: %v", path, err)
	}
}
