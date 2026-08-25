package config

import (
	"fmt"
	"regexp"

	"github.com/Masterminds/semver/v3"
)

// idPattern impose des identifiants stables et sûrs à utiliser comme clés
// JSON dans le rapport (section 3.1 de la spec).
var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

var validTypes = map[string]bool{
	TypeCommand: true,
	TypeDocker:  true,
	TypeEnv:     true,
	TypePath:    true,
}

var validKinds = map[string]bool{
	"": true, "file": true, "directory": true, "any": true,
}

// Validate applique les règles de schéma qui ne sont pas déjà garanties par
// le décodage strict (KnownFields(true) rejette déjà les clés inconnues) :
// version supportée, présence d'au moins une règle, unicité et format des
// id, type reconnu, champs obligatoires par type.
func Validate(file *File) error {
	if file.Version != 1 {
		return newError("", "version %d non supportée (seule la valeur 1 est acceptée)", file.Version)
	}
	if len(file.Checks) == 0 {
		return newError("", "aucun contrôle déclaré dans checks")
	}

	seen := make(map[string]bool, len(file.Checks))
	for i, c := range file.Checks {
		field := fmt.Sprintf("checks[%d]", i)

		if c.ID == "" {
			return newError("", "%s: id manquant", field)
		}
		if !idPattern.MatchString(c.ID) {
			return newError("", "%s (id=%q): l'id doit respecter [a-z0-9][a-z0-9-]{0,62}", field, c.ID)
		}
		if seen[c.ID] {
			return newError("", "id dupliqué : %q", c.ID)
		}
		seen[c.ID] = true

		if !validTypes[c.Type] {
			return newError("", "%s (id=%q): type %q inconnu (attendu : command, docker, env, path)", field, c.ID, c.Type)
		}

		if err := validateTypeFields(field, c); err != nil {
			return err
		}
	}
	return nil
}

func validateTypeFields(field string, c Check) error {
	switch c.Type {
	case TypeCommand:
		if c.Command == "" {
			return newError("", "%s (id=%q): le type command requiert le champ command", field, c.ID)
		}
		if c.Version != "" {
			if _, err := semver.NewConstraint(c.Version); err != nil {
				return newError("", "%s (id=%q): contrainte de version invalide", field, c.ID)
			}
		}
		if c.VersionPattern != "" {
			if _, err := regexp.Compile(c.VersionPattern); err != nil {
				return newError("", "%s.versionPattern est invalide", field)
			}
		}
	case TypeEnv:
		if c.EnvName == "" {
			return newError("", "%s (id=%q): le type env requiert le champ envName", field, c.ID)
		}
	case TypePath:
		if c.Path == "" {
			return newError("", "%s (id=%q): le type path requiert le champ path", field, c.ID)
		}
		if !validKinds[c.Kind] {
			return newError("", "%s (id=%q): kind doit être file, directory ou any", field, c.ID)
		}
	case TypeDocker:
		// Aucun champ spécifique à ce type.
	}
	return nil
}
