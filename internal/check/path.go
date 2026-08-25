package check

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/platform"
)

// PathChecker vérifie qu'un fichier ou un dossier existe, relativement
// au CWD du Runtime (chemins relatifs du YAML).
type PathChecker struct{}

func NewPathChecker() *PathChecker { return &PathChecker{} }

func (c *PathChecker) Type() string { return config.TypePath }

func (c *PathChecker) Run(_ context.Context, def config.Check, rt Runtime) (result Result) {
	start := rt.now()
	result = Result{
		ID:       def.ID,
		Type:     config.TypePath,
		Required: def.IsRequired(),
		Hint:     def.Hint,
		Expected: kindLabel(def.Kind),
	}
	defer func() {
		result.DurationMS = rt.now().Sub(start).Milliseconds()
	}()

	resolved := resolvePath(rt.CWD, def.Path)
	info, err := rt.fs().Stat(resolved)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			result.Status = statusForFailure(def)
			result.Summary = fmt.Sprintf("Le chemin %s est introuvable", def.Path)
			return result
		}
		result.Status = statusForFailure(def)
		result.Summary = fmt.Sprintf("Le chemin %s n'est pas accessible", def.Path)
		return result
	}

	actual := "file"
	if info.IsDir() {
		actual = "directory"
	}
	result.Actual = actual

	if !kindMatches(def.Kind, info) {
		result.Status = statusForFailure(def)
		if info.IsDir() {
			result.Summary = fmt.Sprintf("Le chemin %s est un dossier, un fichier était attendu", def.Path)
		} else {
			result.Summary = fmt.Sprintf("Le chemin %s est un fichier, un dossier était attendu", def.Path)
		}
		return result
	}

	result.Status = Pass
	result.Summary = fmt.Sprintf("Le chemin %s existe", def.Path)
	return result
}

func resolvePath(cwd, p string) string {
	if filepath.IsAbs(p) {
		return filepath.Clean(p)
	}
	if cwd == "" {
		return filepath.Clean(p)
	}
	return filepath.Clean(filepath.Join(cwd, p))
}

func kindLabel(kind string) string {
	switch kind {
	case "file", "directory":
		return kind
	default:
		return "any"
	}
}

func kindMatches(kind string, info platform.FileInfo) bool {
	switch kind {
	case "file":
		return !info.IsDir()
	case "directory":
		return info.IsDir()
	default:
		// "" et "any" : l'existence suffit.
		return true
	}
}
