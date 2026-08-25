package check

import (
	"context"
	"errors"
	"os/exec"
	"strings"

	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/platform"
)

const dockerBin = "docker"

// DockerChecker vérifie que le client Docker est dans le PATH et que le
// démon répond à `docker info` dans le délai du Runner. Le socle de test
// n'appelle jamais un démon réel : uniquement un Runner injecté (section 7
// de la spec).
type DockerChecker struct{}

func NewDockerChecker() *DockerChecker { return &DockerChecker{} }

func (c *DockerChecker) Type() string { return config.TypeDocker }

func (c *DockerChecker) Run(ctx context.Context, def config.Check, rt Runtime) (result Result) {
	start := rt.now()
	result = Result{
		ID:       def.ID,
		Type:     config.TypeDocker,
		Required: def.IsRequired(),
		Hint:     def.Hint,
	}
	defer func() {
		result.DurationMS = rt.now().Sub(start).Milliseconds()
	}()

	runner := rt.runner()
	if _, err := runner.LookPath(dockerBin); err != nil {
		result.Status = statusForFailure(def)
		result.Summary = "Le client Docker est introuvable dans le PATH."
		return result
	}

	run := runner.Run(ctx, dockerBin, []string{"info"})
	if run.TimedOut {
		result.Status = statusForFailure(def)
		result.Summary = "La commande Docker n'a pas répondu en moins de 3 s."
		return result
	}
	if isCommandNotFound(run.Err) {
		result.Status = statusForFailure(def)
		result.Summary = "Le client Docker est introuvable dans le PATH."
		return result
	}
	if run.Err != nil || run.ExitCode != 0 {
		result.Status = statusForFailure(def)
		if dockerPermissionDenied(run) {
			result.Summary = "Le client Docker est disponible, mais les permissions sont insuffisantes."
		} else {
			result.Summary = "Le client Docker est disponible, mais le démon est inaccessible."
		}
		return result
	}

	result.Status = Pass
	result.Summary = "Le démon Docker est accessible."
	return result
}

func isCommandNotFound(err error) bool {
	return err != nil && errors.Is(err, exec.ErrNotFound)
}

// dockerPermissionDenied inspecte stdout/stderr uniquement en interne : rien
// de la sortie brute n'est recopié dans le Result.
func dockerPermissionDenied(run platform.RunResult) bool {
	combined := strings.ToLower(platform.StripANSI(string(run.Stdout) + "\n" + string(run.Stderr)))
	for _, needle := range []string{
		"permission denied",
		"access is denied",
		"operation not permitted",
	} {
		if strings.Contains(combined, needle) {
			return true
		}
	}
	return false
}
