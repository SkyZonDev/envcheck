package core_test

import (
	"context"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/core"
)

func requiredCheck(id, envName string) config.Check {
	yes := true
	return config.Check{ID: id, Type: config.TypeEnv, Required: &yes, EnvName: envName}
}

func optionalCheck(id, envName string) config.Check {
	no := false
	return config.Check{ID: id, Type: config.TypeEnv, Required: &no, EnvName: envName}
}

func TestRun_AllPass(t *testing.T) {
	file := &config.File{Version: 1, Checks: []config.Check{
		requiredCheck("a", "A"),
		requiredCheck("b", "B"),
	}}
	rt := check.Runtime{Env: map[string]string{"A": "1", "B": "2"}}

	report := core.Run(context.Background(), file, "envcheck.yml", "0.1.0", check.DefaultRegistry(), rt)

	if !report.Passed {
		t.Fatalf("Passed = false, attendu true : %+v", report.Summary)
	}
	if report.Summary.Pass != 2 {
		t.Fatalf("Summary.Pass = %d, attendu 2", report.Summary.Pass)
	}
}

func TestRun_RequiredFailureBlocks(t *testing.T) {
	file := &config.File{Version: 1, Checks: []config.Check{
		requiredCheck("a", "A"),
		requiredCheck("missing", "MISSING"),
	}}
	rt := check.Runtime{Env: map[string]string{"A": "1"}}

	report := core.Run(context.Background(), file, "envcheck.yml", "0.1.0", check.DefaultRegistry(), rt)

	if report.Passed {
		t.Fatal("Passed = true, attendu false (une règle requise échoue)")
	}
	if report.Summary.Fail != 1 {
		t.Fatalf("Summary.Fail = %d, attendu 1", report.Summary.Fail)
	}
}

func TestRun_OptionalFailureDoesNotBlock(t *testing.T) {
	file := &config.File{Version: 1, Checks: []config.Check{
		requiredCheck("a", "A"),
		optionalCheck("nice-to-have", "MISSING"),
	}}
	rt := check.Runtime{Env: map[string]string{"A": "1"}}

	report := core.Run(context.Background(), file, "envcheck.yml", "0.1.0", check.DefaultRegistry(), rt)

	if !report.Passed {
		t.Fatalf("Passed = false, attendu true (l'échec est optionnel donc warn) : %+v", report.Summary)
	}
	if report.Summary.Warn != 1 {
		t.Fatalf("Summary.Warn = %d, attendu 1", report.Summary.Warn)
	}
}

func TestRun_UnknownTypeDoesNotStopOtherChecks(t *testing.T) {
	yes := true
	file := &config.File{Version: 1, Checks: []config.Check{
		{ID: "future", Type: "plugin", Required: &yes},
		requiredCheck("a", "A"),
	}}
	rt := check.Runtime{Env: map[string]string{"A": "1"}}

	report := core.Run(context.Background(), file, "envcheck.yml", "0.1.0", check.DefaultRegistry(), rt)

	if len(report.Checks) != 2 {
		t.Fatalf("len(Checks) = %d, attendu 2 (type inconnu ne doit pas bloquer env)", len(report.Checks))
	}
	if report.Checks[0].Status != check.Error {
		t.Fatalf("type inconnu requis = error, got %+v", report.Checks[0])
	}
	if report.Checks[1].Status != check.Pass {
		t.Fatalf("le second contrôle (env, valide) devrait quand même s'exécuter : %+v", report.Checks[1])
	}
	if report.Passed {
		t.Fatal("Passed = true, attendu false (type requis inconnu = error)")
	}
}

func TestRun_PreservesFileOrder(t *testing.T) {
	file := &config.File{Version: 1, Checks: []config.Check{
		requiredCheck("z", "Z"),
		requiredCheck("a", "A"),
		requiredCheck("m", "M"),
	}}
	rt := check.Runtime{Env: map[string]string{"Z": "1", "A": "1", "M": "1"}}

	report := core.Run(context.Background(), file, "envcheck.yml", "0.1.0", check.DefaultRegistry(), rt)

	gotOrder := []string{report.Checks[0].ID, report.Checks[1].ID, report.Checks[2].ID}
	wantOrder := []string{"z", "a", "m"}
	for i := range wantOrder {
		if gotOrder[i] != wantOrder[i] {
			t.Fatalf("ordre = %v, attendu %v", gotOrder, wantOrder)
		}
	}
}
