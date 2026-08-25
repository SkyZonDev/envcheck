package check_test

import (
	"context"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/config"
)

func envDef(id, name string, required, allowEmpty bool) config.Check {
	return config.Check{
		ID:         id,
		Type:       config.TypeEnv,
		Required:   &required,
		EnvName:    name,
		AllowEmpty: allowEmpty,
		Hint:       "corrige ça",
	}
}

func TestEnvChecker_Present(t *testing.T) {
	c := check.NewEnvChecker()
	rt := check.Runtime{Env: map[string]string{"DATABASE_URL": "postgres://localhost"}}

	got := c.Run(context.Background(), envDef("database-url", "DATABASE_URL", true, false), rt)

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass", got.Status)
	}
	if got.Actual != "" {
		t.Fatalf("Actual = %q, la valeur d'une variable ne doit JAMAIS apparaître dans le résultat", got.Actual)
	}
}

func TestEnvChecker_MissingIsFailWhenRequired(t *testing.T) {
	c := check.NewEnvChecker()
	rt := check.Runtime{Env: map[string]string{}}

	got := c.Run(context.Background(), envDef("database-url", "DATABASE_URL", true, false), rt)

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
}

func TestEnvChecker_MissingIsWarnWhenOptional(t *testing.T) {
	c := check.NewEnvChecker()
	rt := check.Runtime{Env: map[string]string{}}

	got := c.Run(context.Background(), envDef("sentry-dsn", "SENTRY_DSN", false, false), rt)

	if got.Status != check.Warn {
		t.Fatalf("Status = %v, attendu warn pour une règle optionnelle", got.Status)
	}
}

func TestEnvChecker_EmptyFailsUnlessAllowed(t *testing.T) {
	c := check.NewEnvChecker()
	rt := check.Runtime{Env: map[string]string{"DATABASE_URL": ""}}

	notAllowed := c.Run(context.Background(), envDef("database-url", "DATABASE_URL", true, false), rt)
	if notAllowed.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail (chaîne vide, allowEmpty=false)", notAllowed.Status)
	}

	allowed := c.Run(context.Background(), envDef("database-url", "DATABASE_URL", true, true), rt)
	if allowed.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (allowEmpty=true)", allowed.Status)
	}
}

func TestEnvChecker_NeverLeaksValueInSummaryOrDetail(t *testing.T) {
	c := check.NewEnvChecker()
	secret := "super-secret-token-xyz"
	rt := check.Runtime{Env: map[string]string{"API_TOKEN": secret}}

	got := c.Run(context.Background(), envDef("api-token", "API_TOKEN", true, false), rt)

	for _, field := range []string{got.Summary, got.Detail, got.Actual, got.Expected} {
		if field != "" && field == secret {
			t.Fatalf("la valeur secrète a fuité dans un champ du résultat: %+v", got)
		}
	}
}
