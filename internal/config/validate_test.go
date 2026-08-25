package config_test

import (
	"strings"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/config"
)

func validCommand(id string) config.Check {
	return config.Check{ID: id, Type: config.TypeCommand, Command: "git"}
}

func TestValidate_InvalidVersionPattern(t *testing.T) {
	file := &config.File{
		Version: 1,
		Checks: []config.Check{{
			ID:             "git",
			Type:           config.TypeCommand,
			Command:        "git",
			VersionPattern: "(",
		}},
	}
	err := config.Validate(file)
	if err == nil {
		t.Fatal("Validate() aurait dû rejeter un versionPattern invalide")
	}
	if !strings.Contains(err.Error(), "versionPattern") {
		t.Fatalf("erreur = %q, attendu une mention de versionPattern", err)
	}
}

func TestValidate_InvalidVersionConstraint(t *testing.T) {
	file := &config.File{
		Version: 1,
		Checks: []config.Check{{
			ID:      "git",
			Type:    config.TypeCommand,
			Command: "git",
			Version: "pas-semver",
		}},
	}
	err := config.Validate(file)
	if err == nil {
		t.Fatal("Validate() aurait dû rejeter une contrainte SemVer invalide")
	}
}

func TestValidate_AcceptsSemVerRange(t *testing.T) {
	file := &config.File{
		Version: 1,
		Checks: []config.Check{{
			ID:             "node",
			Type:           config.TypeCommand,
			Command:        "node",
			Version:        ">=20.0.0 <23.0.0",
			VersionArgs:    []string{"--version"},
			VersionPattern: `([0-9]+\.[0-9]+\.[0-9]+)`,
		}},
	}
	if err := config.Validate(file); err != nil {
		t.Fatalf("Validate() erreur inattendue: %v", err)
	}
}

func TestValidate_CommandRequiresCommandField(t *testing.T) {
	file := &config.File{
		Version: 1,
		Checks:  []config.Check{{ID: "git", Type: config.TypeCommand}},
	}
	err := config.Validate(file)
	if err == nil {
		t.Fatal("Validate() aurait dû exiger le champ command")
	}
}

func TestValidate_PathRequiresPathField(t *testing.T) {
	file := &config.File{
		Version: 1,
		Checks:  []config.Check{{ID: "migrations", Type: config.TypePath}},
	}
	err := config.Validate(file)
	if err == nil {
		t.Fatal("Validate() aurait dû exiger le champ path")
	}
}

func TestValidate_InvalidKind(t *testing.T) {
	file := &config.File{
		Version: 1,
		Checks:  []config.Check{{ID: "x", Type: config.TypePath, Path: "./x", Kind: "symlink"}},
	}
	err := config.Validate(file)
	if err == nil {
		t.Fatal("Validate() aurait dû rejeter kind=symlink")
	}
}

func TestValidate_MinimalCommand(t *testing.T) {
	file := &config.File{Version: 1, Checks: []config.Check{validCommand("git")}}
	if err := config.Validate(file); err != nil {
		t.Fatalf("un contrôle command sans version est valide: %v", err)
	}
}
