package check

import (
	"context"
	"fmt"

	"github.com/SkyZonDev/envcheck/internal/config"
)

// EnvChecker vérifie la présence (et, sauf allowEmpty, la non-vacuité)
// d'une variable d'environnement. Il ne journalise et ne renvoie jamais sa
// valeur : seul le fait qu'elle soit présente/vide est observable.
type EnvChecker struct{}

func NewEnvChecker() *EnvChecker { return &EnvChecker{} }

func (c *EnvChecker) Type() string { return config.TypeEnv }

func (c *EnvChecker) Run(ctx context.Context, def config.Check, rt Runtime) Result {
	start := rt.now()

	result := Result{
		ID:       def.ID,
		Type:     config.TypeEnv,
		Required: def.IsRequired(),
		Hint:     def.Hint,
	}

	value, present := rt.Env[def.EnvName]
	switch {
	case !present:
		result.Status = statusForFailure(def)
		result.Summary = fmt.Sprintf("La variable %s est absente", def.EnvName)
	case value == "" && !def.AllowEmpty:
		result.Status = statusForFailure(def)
		result.Summary = fmt.Sprintf("La variable %s est vide", def.EnvName)
	default:
		result.Status = Pass
		result.Summary = fmt.Sprintf("La variable %s est définie", def.EnvName)
	}

	result.DurationMS = rt.now().Sub(start).Milliseconds()
	return result
}
