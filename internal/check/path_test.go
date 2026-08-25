package check_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/platform"
)

func pathDef(id, p, kind string, required bool) config.Check {
	return config.Check{
		ID:       id,
		Type:     config.TypePath,
		Required: &required,
		Path:     p,
		Kind:     kind,
		Hint:     "créez le chemin",
	}
}

func TestPathChecker_Directory(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "migrations")
	if err := os.Mkdir(sub, 0o755); err != nil {
		t.Fatal(err)
	}

	got := check.NewPathChecker().Run(context.Background(), pathDef("migrations", "./migrations", "directory", true), check.Runtime{
		CWD: dir,
		FS:  platform.OSFS{},
	})

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (%s)", got.Status, got.Summary)
	}
	if got.Actual != "directory" {
		t.Fatalf("Actual = %q, attendu directory", got.Actual)
	}
}

func TestPathChecker_File(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := check.NewPathChecker().Run(context.Background(), pathDef("go-mod", "./go.mod", "file", true), check.Runtime{
		CWD: dir,
		FS:  platform.OSFS{},
	})

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (%s)", got.Status, got.Summary)
	}
	if got.Actual != "file" {
		t.Fatalf("Actual = %q, attendu file", got.Actual)
	}
}

func TestPathChecker_AnyKindAcceptsFileOrDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "README"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	c := check.NewPathChecker()
	rt := check.Runtime{CWD: dir, FS: platform.OSFS{}}

	for _, kind := range []string{"", "any"} {
		got := c.Run(context.Background(), pathDef("readme", "README", kind, true), rt)
		if got.Status != check.Pass {
			t.Fatalf("kind=%q Status = %v, attendu pass", kind, got.Status)
		}
	}
}

func TestPathChecker_MissingIsFailWhenRequired(t *testing.T) {
	dir := t.TempDir()
	got := check.NewPathChecker().Run(context.Background(), pathDef("migrations", "./migrations", "directory", true), check.Runtime{
		CWD: dir,
		FS:  platform.OSFS{},
	})
	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
}

func TestPathChecker_MissingIsWarnWhenOptional(t *testing.T) {
	dir := t.TempDir()
	got := check.NewPathChecker().Run(context.Background(), pathDef("optional-dir", "./missing", "directory", false), check.Runtime{
		CWD: dir,
		FS:  platform.OSFS{},
	})
	if got.Status != check.Warn {
		t.Fatalf("Status = %v, attendu warn pour une règle optionnelle", got.Status)
	}
}

func TestPathChecker_WrongKind(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "migrations"), []byte("not a dir"), 0o644); err != nil {
		t.Fatal(err)
	}

	got := check.NewPathChecker().Run(context.Background(), pathDef("migrations", "migrations", "directory", true), check.Runtime{
		CWD: dir,
		FS:  platform.OSFS{},
	})
	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail (fichier à la place d'un dossier)", got.Status)
	}
	if got.Actual != "file" {
		t.Fatalf("Actual = %q, attendu file (ce qui a été trouvé)", got.Actual)
	}
}

func TestPathChecker_RelativeToCWD(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "src")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "main.go"), []byte("package main\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// CWD = dir, path relatif src/main.go — pas un chemin déjà absolu.
	got := check.NewPathChecker().Run(context.Background(), pathDef("main", "src/main.go", "file", true), check.Runtime{
		CWD: dir,
		FS:  platform.OSFS{},
	})
	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (%s)", got.Status, got.Summary)
	}
}
