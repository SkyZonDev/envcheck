package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Error couvre tout ce que l'utilisateur peut corriger lui-même : fichier
// absent, YAML mal formé, champ inconnu, schéma invalide. Le CLI mappe ce
// type sur le code de sortie 2 (voir internal/core/exitcode.go). Toute
// autre erreur remontée par Load est considérée interne (code 3).
type Error struct {
	Path string
	Msg  string
}

func (e *Error) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s: %s", e.Path, e.Msg)
	}
	return e.Msg
}

func newError(path, format string, args ...any) *Error {
	return &Error{Path: path, Msg: fmt.Sprintf(format, args...)}
}

// Load lit path, décode le YAML en refusant les champs inconnus
// (KnownFields(true) — voir section 3 de la spec), puis valide le schéma.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, newError(path, "fichier introuvable")
		}
		return nil, newError(path, "lecture impossible : %v", err)
	}

	var file File
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(&file); err != nil {
		return nil, newError(path, "YAML invalide : %v", err)
	}

	if err := Validate(&file); err != nil {
		var verr *Error
		if errors.As(err, &verr) {
			verr.Path = path
		}
		return nil, err
	}

	return &file, nil
}
