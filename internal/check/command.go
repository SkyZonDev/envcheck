package check

import (
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/Masterminds/semver/v3"

	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/platform"
)

// defaultVersionPattern capture un triplet SemVer dans la sortie d'un
// binaire. Un versionPattern YAML le remplace si le format n'est pas
// standard (ex: `go version go1.23.0`).
const defaultVersionPattern = `([0-9]+\.[0-9]+\.[0-9]+)`

// CommandChecker vérifie qu'un exécutable est dans le PATH et, le cas
// échéant, que sa version satisfait une contrainte SemVer. L'exécutable
// et ses arguments sont toujours transmis séparément à os/exec (pas de
// shell). Un timeout de 3 s est appliqué par le Runner.
type CommandChecker struct{}

func NewCommandChecker() *CommandChecker { return &CommandChecker{} }

func (c *CommandChecker) Type() string { return config.TypeCommand }

func (c *CommandChecker) Run(ctx context.Context, def config.Check, rt Runtime) (result Result) {
	start := rt.now()
	result = Result{
		ID:       def.ID,
		Type:     config.TypeCommand,
		Required: def.IsRequired(),
		Hint:     def.Hint,
		Expected: def.Version,
	}
	defer func() {
		result.DurationMS = rt.now().Sub(start).Milliseconds()
	}()

	runner := rt.runner()
	if _, err := runner.LookPath(def.Command); err != nil {
		result.Status = statusForFailure(def)
		result.Summary = fmt.Sprintf("La commande `%s` est introuvable dans le PATH.", def.Command)
		return result
	}

	if def.Version == "" {
		result.Status = Pass
		result.Summary = fmt.Sprintf("La commande `%s` est disponible.", def.Command)
		return result
	}

	args := def.VersionArgs
	if len(args) == 0 {
		args = []string{"--version"}
	}

	run := runner.Run(ctx, def.Command, args)
	if run.TimedOut {
		result.Status = statusForFailure(def)
		result.Summary = fmt.Sprintf("La commande `%s` n'a pas répondu en moins de 3 s.", def.Command)
		return result
	}

	output := platform.StripANSI(string(run.Stdout) + "\n" + string(run.Stderr))
	extracted := extractVersion(output, versionPattern(def))
	if extracted == "" {
		result.Status = statusForFailure(def)
		result.Summary = fmt.Sprintf("Impossible d'extraire la version de `%s`.", def.Command)
		return result
	}

	ver, err := semver.NewVersion(extracted)
	if err != nil {
		result.Status = statusForFailure(def)
		result.Summary = fmt.Sprintf("Version extraite invalide : %s.", extracted)
		result.Actual = extracted
		return result
	}

	constraint, err := semver.NewConstraint(def.Version)
	if err != nil {
		// Normalement rejeté à la validation YAML (code 2). Défense en profondeur.
		result.Status = Error
		result.Summary = fmt.Sprintf("contrainte de version invalide : %s", def.Version)
		return result
	}

	result.Actual = extracted
	if !constraint.Check(ver) {
		result.Status = statusForFailure(def)
		result.Summary = fmt.Sprintf("Détectée : %s ; attendue : %s.", extracted, def.Version)
		return result
	}

	result.Status = Pass
	result.Summary = fmt.Sprintf("%s %s satisfait %s", displayName(def.Command), extracted, def.Version)
	return result
}

func versionPattern(def config.Check) string {
	if def.VersionPattern != "" {
		return def.VersionPattern
	}
	return defaultVersionPattern
}

func extractVersion(output, pattern string) string {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return ""
	}
	m := re.FindStringSubmatch(output)
	if len(m) == 0 {
		return ""
	}
	if len(m) >= 2 && m[1] != "" {
		return m[1]
	}
	return m[0]
}

func displayName(cmd string) string {
	base := filepath.Base(cmd)
	if base == "" || base == "." {
		return cmd
	}
	r, size := utf8.DecodeRuneInString(base)
	if r == utf8.RuneError {
		return base
	}
	return string(unicode.ToUpper(r)) + base[size:]
}
