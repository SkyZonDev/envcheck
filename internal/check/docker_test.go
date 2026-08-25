package check_test

import (
	"context"
	"errors"
	"os/exec"
	"strings"
	"testing"

	"github.com/SkyZonDev/envcheck/internal/check"
	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/platform"
)

func dockerDef(required bool) config.Check {
	return config.Check{
		ID:       "docker-daemon",
		Type:     config.TypeDocker,
		Required: &required,
		Hint:     "Démarrez Docker Desktop ou le service Docker.",
	}
}

func TestDockerChecker_ClientMissing(t *testing.T) {
	stub := &platform.StubRunner{LookPathErr: exec.ErrNotFound}
	got := check.NewDockerChecker().Run(context.Background(), dockerDef(true), check.Runtime{Exec: stub})

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
	if !strings.Contains(got.Summary, "client") || !strings.Contains(strings.ToLower(got.Summary), "introuvable") {
		t.Fatalf("Summary = %q, attendu client introuvable", got.Summary)
	}
	if stub.RunCalled {
		t.Fatal("sans client, docker info ne doit pas être lancé")
	}
}

func TestDockerChecker_ClientMissingIsWarnWhenOptional(t *testing.T) {
	stub := &platform.StubRunner{LookPathErr: exec.ErrNotFound}
	got := check.NewDockerChecker().Run(context.Background(), dockerDef(false), check.Runtime{Exec: stub})
	if got.Status != check.Warn {
		t.Fatalf("Status = %v, attendu warn", got.Status)
	}
}

func TestDockerChecker_DaemonUnreachable(t *testing.T) {
	raw := "Cannot connect to the Docker daemon at unix:///var/run/docker.sock. Is the docker daemon running?"
	stub := &platform.StubRunner{
		Result: platform.RunResult{ExitCode: 1, Err: errExit, Stderr: []byte(raw)},
	}
	got := check.NewDockerChecker().Run(context.Background(), dockerDef(true), check.Runtime{Exec: stub})

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
	if !strings.Contains(got.Summary, "daemon") && !strings.Contains(got.Summary, "démon") {
		t.Fatalf("Summary = %q, attendu daemon inaccessible (client présent)", got.Summary)
	}
	if strings.Contains(got.Summary, "introuvable") {
		t.Fatalf("ne doit pas confondre client absent et daemon down: %q", got.Summary)
	}
	assertNoRawLeak(t, got, raw)
	if len(stub.RunArgs) != 1 || stub.RunArgs[0] != "info" {
		t.Fatalf("RunArgs = %v, attendu [info] (pas de shell)", stub.RunArgs)
	}
}

func TestDockerChecker_PermissionDenied(t *testing.T) {
	raw := "Got permission denied while trying to connect to the Docker daemon socket at unix:///var/run/docker.sock"
	stub := &platform.StubRunner{
		Result: platform.RunResult{ExitCode: 1, Err: errExit, Stderr: []byte(raw)},
	}
	got := check.NewDockerChecker().Run(context.Background(), dockerDef(true), check.Runtime{Exec: stub})

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
	if !strings.Contains(strings.ToLower(got.Summary), "permission") {
		t.Fatalf("Summary = %q, attendu permissions insuffisantes", got.Summary)
	}
	assertNoRawLeak(t, got, raw)
}

func TestDockerChecker_Timeout(t *testing.T) {
	stub := &platform.StubRunner{Result: platform.RunResult{TimedOut: true}}
	got := check.NewDockerChecker().Run(context.Background(), dockerDef(true), check.Runtime{Exec: stub})

	if got.Status != check.Fail {
		t.Fatalf("Status = %v, attendu fail", got.Status)
	}
	if !strings.Contains(got.Summary, "3 s") {
		t.Fatalf("Summary = %q, attendu mention du délai", got.Summary)
	}
}

func TestDockerChecker_DaemonOK(t *testing.T) {
	stub := &platform.StubRunner{
		Result: platform.RunResult{ExitCode: 0, Stdout: []byte("Server Version: 27.0.3\n")},
	}
	got := check.NewDockerChecker().Run(context.Background(), dockerDef(true), check.Runtime{Exec: stub})

	if got.Status != check.Pass {
		t.Fatalf("Status = %v, attendu pass (%s)", got.Status, got.Summary)
	}
	assertNoRawLeak(t, got, "Server Version")
}

func assertNoRawLeak(t *testing.T, got check.Result, rawFragment string) {
	t.Helper()
	for _, field := range []string{got.Summary, got.Detail, got.Actual, got.Expected} {
		if rawFragment != "" && strings.Contains(field, rawFragment) {
			t.Fatalf("sortie brute Docker fuitée dans %+v", got)
		}
	}
}

// errExit simule un processus terminé hors zéro. Le Checker s'appuie sur
// ExitCode / Err non nil, pas sur le type exact.
var errExit = errors.New("exit status 1")
