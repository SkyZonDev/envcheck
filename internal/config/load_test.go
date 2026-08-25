package config_test

import (
	"errors"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/config"
)

func TestLoad_ValidFile(t *testing.T) {
	file, err := config.Load("../../testdata/configs/valid.yml")
	if err != nil {
		t.Fatalf("Load() erreur inattendue: %v", err)
	}
	if file.Version != 1 {
		t.Fatalf("Version = %d, attendu 1", file.Version)
	}
	if len(file.Checks) != 4 {
		t.Fatalf("len(Checks) = %d, attendu 4", len(file.Checks))
	}

	// La règle docker-daemon est explicitement required: true.
	docker := file.Checks[1]
	if docker.ID != "docker-daemon" || !docker.IsRequired() {
		t.Fatalf("docker-daemon devrait être requis, got %+v", docker)
	}

	// git n'a pas de champ required : le défaut (true) doit s'appliquer.
	git := file.Checks[0]
	if !git.IsRequired() {
		t.Fatalf("git devrait être requis par défaut, got %+v", git)
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := config.Load("../../testdata/configs/does-not-exist.yml")
	assertConfigError(t, err)
}

func TestLoad_UnknownField(t *testing.T) {
	_, err := config.Load("../../testdata/configs/unknown-field.yml")
	assertConfigError(t, err)
}

func TestLoad_DuplicateID(t *testing.T) {
	_, err := config.Load("../../testdata/configs/duplicate-id.yml")
	assertConfigError(t, err)
}

func TestLoad_UnsupportedSchemaVersion(t *testing.T) {
	_, err := config.Load("../../testdata/configs/bad-schema-version.yml")
	assertConfigError(t, err)
}

// assertConfigError vérifie que l'erreur est bien un *config.Error : c'est
// ce type que le CLI mappera sur le code de sortie 2 (fichier absent, YAML
// invalide, schéma invalide) plutôt que 3 (erreur interne inattendue).
func assertConfigError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("Load() aurait dû échouer")
	}
	var cfgErr *config.Error
	if !errors.As(err, &cfgErr) {
		t.Fatalf("erreur = %T, attendu *config.Error", err)
	}
}
