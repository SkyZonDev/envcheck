package platform

import (
	"io/fs"
	"os"
	"path/filepath"
)

// FileInfo est le sous-ensemble d'os.FileInfo dont les Checkers ont besoin.
// On n'expose ni le nom, ni le mode, ni la taille : un contrôle path ne
// doit juger que l'existence et le type (fichier vs dossier).
type FileInfo interface {
	IsDir() bool
}

// FileSystem isole l'accès disque pour que les tests n'aient pas à
// dépendre de la machine hôte (section 5 de la spec).
type FileSystem interface {
	Stat(name string) (FileInfo, error)
}

type fileInfo struct {
	dir bool
}

func (fi fileInfo) IsDir() bool { return fi.dir }

// OSFS est l'adaptateur sur le système de fichiers réel.
type OSFS struct{}

func (OSFS) Stat(name string) (FileInfo, error) {
	fi, err := os.Stat(name)
	if err != nil {
		return nil, err
	}
	return fileInfo{dir: fi.IsDir()}, nil
}

// DirInfo implémente FileInfo pour les tests (MapFS).
type DirInfo struct {
	Dir bool
}

func (d DirInfo) IsDir() bool { return d.Dir }

// MapFS est un FileSystem en mémoire, indexé par chemin nettoyé
// (filepath.Clean). Une entrée absente se comporte comme os.ErrNotExist.
type MapFS map[string]DirInfo

func (m MapFS) Stat(name string) (FileInfo, error) {
	if e, ok := m[filepath.Clean(name)]; ok {
		return e, nil
	}
	return nil, fs.ErrNotExist
}
