package platform_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/platform"
)

func TestOSFS_StatFileAndDir(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(file, []byte("module x\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	fs := platform.OSFS{}

	fi, err := fs.Stat(file)
	if err != nil {
		t.Fatalf("Stat(file) = %v", err)
	}
	if fi.IsDir() {
		t.Fatal("go.mod devrait être un fichier")
	}

	di, err := fs.Stat(dir)
	if err != nil {
		t.Fatalf("Stat(dir) = %v", err)
	}
	if !di.IsDir() {
		t.Fatal("TempDir devrait être un dossier")
	}
}

func TestOSFS_Missing(t *testing.T) {
	fs := platform.OSFS{}
	_, err := fs.Stat(filepath.Join(t.TempDir(), "absent"))
	if err == nil {
		t.Fatal("Stat d'un chemin absent devrait échouer")
	}
}

func TestMapFS_CleanPathAndMissing(t *testing.T) {
	dir := filepath.Join("proj", "migrations")
	m := platform.MapFS{
		filepath.Clean(dir): {Dir: true},
	}

	fi, err := m.Stat(dir)
	if err != nil {
		t.Fatalf("Stat = %v", err)
	}
	if !fi.IsDir() {
		t.Fatal("migrations devrait être un dossier")
	}

	_, err = m.Stat("proj/missing")
	if err == nil {
		t.Fatal("chemin absent devrait renvoyer une erreur")
	}
}
