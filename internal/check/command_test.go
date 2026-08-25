package check_test

import (
	"context"
	"os/exec"
	"strings"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/platform"
)

func cmdDef(id, command, version string, required bool) config.Check {
	return config.Check{
		ID:       id,
		Type:     config.TypeCommand,
		Required: &required,
		Command:  command,
		Version:  version,
		Hint:     "installez l'outil",
	}
}

func TestCommandChecker_MissingIsFailWhenRequired(t *testing.T) {
	stub := &platform.StubRunner{LookPathErr: exec.ErrNotFound}
	got := check.NewCommandChecker().Run(context.Background(), cmdDef("node", "node", ">=20.0.0", true), check.Runtime{Exec: stub})

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
	if !strings.Contains(got.Summary, "introuvable") {
		t.Fatalf("Summary = %q, attendu un message d'introuvable", got.Summary)
	}
	if stub.RunCalled {
		t.Fatal("un binaire introuvable ne doit pas être exécuté")
	}
}

func TestCommandChecker_MissingIsWarnWhenOptional(t *testing.T) {
	stub := &platform.StubRunner{LookPathErr: exec.ErrNotFound}
	got := check.NewCommandChecker().Run(context.Background(), cmdDef("node", "node", "", false), check.Runtime{Exec: stub})

	if got.Status != check.Warn {
		t.Fatalf("Status = %v, attendu warn", got.Status)
	}
}

func TestCommandChecker_PresentWithoutVersionDoesNotExecute(t *testing.T) {
	stub := &platform.StubRunner{}
	got := check.NewCommandChecker().Run(context.Background(), cmdDef("git", "git", "", true), check.Runtime{Exec: stub})

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass", got.Status)
	}
	if stub.RunCalled {
		t.Fatal("sans contrainte de version, le binaire ne doit pas être exécuté")
	}
}

func TestCommandChecker_VersionInRange(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{Stdout: []byte("git version 2.45.1")},
	}
	def := cmdDef("git", "git", ">=2.40.0", true)
	def.VersionArgs = []string{"--version"}

	got := check.NewCommandChecker().Run(context.Background(), def, check.Runtime{Exec: stub})

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (%s)", got.Status, got.Summary)
	}
	if got.Actual != "2.45.1" {
		t.Fatalf("Actual = %q, attendu 2.45.1", got.Actual)
	}
	if got.Expected != ">=2.40.0" {
		t.Fatalf("Expected = %q", got.Expected)
	}
	if !strings.Contains(got.Summary, "2.45.1") || !strings.Contains(got.Summary, ">=2.40.0") {
		t.Fatalf("Summary = %q, attendu version + contrainte", got.Summary)
	}
	if got.Detail != "" || strings.Contains(got.Summary, "git version") {
		t.Fatalf("la sortie brute du processus a fuité dans le résultat: %+v", got)
	}
	if len(stub.RunArgs) != 1 || stub.RunArgs[0] != "--version" {
		t.Fatalf("RunArgs = %v, attendu [--version] (pas de shell)", stub.RunArgs)
	}
}

func TestCommandChecker_VersionOutOfRange(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{Stdout: []byte("v18.20.4")},
	}
	got := check.NewCommandChecker().Run(context.Background(), cmdDef("node", "node", ">=20.0.0 <23.0.0", true), check.Runtime{Exec: stub})

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
	if got.Actual != "18.20.4" {
		t.Fatalf("Actual = %q, attendu 18.20.4", got.Actual)
	}
	if !strings.Contains(got.Summary, "18.20.4") || !strings.Contains(got.Summary, ">=20.0.0 <23.0.0") {
		t.Fatalf("Summary = %q", got.Summary)
	}
}

func TestCommandChecker_Timeout(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{TimedOut: true},
	}
	got := check.NewCommandChecker().Run(context.Background(), cmdDef("docker", "docker", ">=24.0.0", true), check.Runtime{Exec: stub})

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
	if !strings.Contains(got.Summary, "3 s") {
		t.Fatalf("Summary = %q, attendu mention du délai de 3 s", got.Summary)
	}
}

func TestCommandChecker_CustomVersionArgsAreNotShelled(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{Stdout: []byte("go version go1.23.2 darwin/arm64")},
	}
	def := cmdDef("go", "go", ">=1.22.0", true)
	def.VersionArgs = []string{"version"}
	def.VersionPattern = `go([0-9]+\.[0-9]+(?:\.[0-9]+)?)`

	got := check.NewCommandChecker().Run(context.Background(), def, check.Runtime{Exec: stub})

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (%s)", got.Status, got.Summary)
	}
	if got.Actual != "1.23.2" {
		t.Fatalf("Actual = %q, attendu 1.23.2", got.Actual)
	}
	if len(stub.RunArgs) != 1 || stub.RunArgs[0] != "version" {
		t.Fatalf("RunArgs = %v, attendu [version] (arguments séparés, pas `go version` en une chaîne)", stub.RunArgs)
	}
}

func TestCommandChecker_StripsANSIBeforeExtractingVersion(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{Stdout: []byte("\x1b[32m2.45.1\x1b[0m")},
	}
	got := check.NewCommandChecker().Run(context.Background(), cmdDef("git", "git", ">=2.40.0", true), check.Runtime{Exec: stub})

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (%s)", got.Status, got.Summary)
	}
	if got.Actual != "2.45.1" {
		t.Fatalf("Actual = %q, attendu 2.45.1 après suppression ANSI", got.Actual)
	}
}

func TestCommandChecker_UnparseableOutput(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{Stdout: []byte("not a version at all")},
	}
	got := check.NewCommandChecker().Run(context.Background(), cmdDef("git", "git", ">=2.40.0", true), check.Runtime{Exec: stub})

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
	if strings.Contains(got.Summary, "not a version") || got.Detail != "" {
		t.Fatalf("la sortie brute a fuité: %+v", got)
	}
}

func TestCommandChecker_ReadsVersionFromStderr(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{Stderr: []byte("tool 3.1.4 extra")},
	}
	got := check.NewCommandChecker().Run(context.Background(), cmdDef("tool", "tool", ">=3.0.0", true), check.Runtime{Exec: stub})

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (%s)", got.Status, got.Summary)
	}
	if got.Actual != "3.1.4" {
		t.Fatalf("Actual = %q, attendu 3.1.4 (stderr)", got.Actual)
	}
}

func TestCommandChecker_DefaultVersionArgs(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{Stdout: []byte("1.2.3")},
	}
	check.NewCommandChecker().Run(context.Background(), cmdDef("tool", "tool", ">=1.0.0", true), check.Runtime{Exec: stub})

	if len(stub.RunArgs) != 1 || stub.RunArgs[0] != "--version" {
		t.Fatalf("sans versionArgs, attendu [--version], got %v", stub.RunArgs)
	}
}
