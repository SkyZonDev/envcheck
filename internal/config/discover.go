package config

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotFound est renvoyé quand aucun fichier candidat n'existe dans le
// répertoire courant. Le CLI le traduit en code de sortie 2.
var ErrNotFound = errors.New("aucun fichier envcheck.yml ou .envcheck.yml trouvé dans le répertoire courant")

// candidateNames est ordonné : envcheck.yml est préféré à .envcheck.yml.
var candidateNames = []string{"envcheck.yml", ".envcheck.yml"}

// Discover cherche un fichier de configuration dans dir EXCLUSIVEMENT :
// on ne remonte jamais vers les répertoires parents, pour qu'un dépôt ne
// puisse pas hériter silencieusement de la configuration d'un autre projet.
func Discover(dir string) (string, error) {
	for _, name := range candidateNames {
		candidate := filepath.Join(dir, name)
		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", ErrNotFound
}
