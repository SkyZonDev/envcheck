package platform

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

// TestExecRunnerHelper n'est pas un vrai test : c'est le processus enfant
// lancé par ExecRunner dans les tests ci-dessous. Il s'arrête immédiatement
// si la variable d'environnement n'est pas posée.
func TestExecRunnerHelper(t *testing.T) {
	if os.Getenv("ENVCHECK_WANT_HELPER") != "1" {
		return
	}
	args := helperArgs()
	if len(args) == 0 {
		os.Exit(0)
	}
	switch args[0] {
	case "echo":
		fmt.Print(strings.Join(args[1:], " "))
	case "sleep":
		time.Sleep(5 * time.Second)
	case "big":
		fmt.Print(strings.Repeat("x", 100_000))
	case "exit":
		os.Exit(3)
	default:
		fmt.Fprintf(os.Stderr, "helper inconnu: %s\n", args[0])
		os.Exit(2)
	}
	os.Exit(0)
}

func helperArgs() []string {
	for i, a := range os.Args {
		if a == "--" {
			return os.Args[i+1:]
		}
	}
	return nil
}

func helperRunner(t *testing.T) *ExecRunner {
	t.Helper()
	return &ExecRunner{
		Timeout:  2 * time.Second,
		extraEnv: []string{"ENVCHECK_WANT_HELPER=1"},
	}
}

func helperCall(subcommand string, extra ...string) (string, []string) {
	args := append([]string{"-test.run=TestExecRunnerHelper", "--", subcommand}, extra...)
	return os.Args[0], args
}

func TestExecRunner_CapturesStdoutWithoutShell(t *testing.T) {
	r := helperRunner(t)
	name, args := helperCall("echo", "hello", "world")

	res := r.Run(context.Background(), name, args)
	if res.TimedOut {
		t.Fatalf("timeout inattendu: %v", res.Err)
	}
	if res.Err != nil {
		t.Fatalf("Run = %v (stderr=%q)", res.Err, res.Stderr)
	}
	// Le binaire de test écrit "=== RUN ..." sur stdout avant le helper :
	// on vérifie la charge utile, pas l'égalité stricte.
	if got := string(res.Stdout); !strings.Contains(got, "hello world") {
		t.Fatalf("stdout = %q, attendu de contenir %q", got, "hello world")
	}
}

func TestExecRunner_Timeout(t *testing.T) {
	r := helperRunner(t)
	r.Timeout = 200 * time.Millisecond
	name, args := helperCall("sleep")

	start := time.Now()
	res := r.Run(context.Background(), name, args)
	elapsed := time.Since(start)

	if !res.TimedOut {
		t.Fatalf("TimedOut = false, attendu true (err=%v)", res.Err)
	}
	if elapsed > 2*time.Second {
		t.Fatalf("le timeout n'a pas tué le processus assez vite: %s", elapsed)
	}
}

func TestExecRunner_CapsCaptureAt32KiB(t *testing.T) {
	r := helperRunner(t)
	r.MaxCapture = 64
	name, args := helperCall("big")

	res := r.Run(context.Background(), name, args)
	if res.Err != nil && !res.TimedOut {
		t.Fatalf("Run = %v", res.Err)
	}
	if len(res.Stdout) != 64 {
		t.Fatalf("len(stdout) = %d, attendu 64", len(res.Stdout))
	}
}

func TestExecRunner_LookPathRejectsShellExpression(t *testing.T) {
	r := NewExecRunner()
	if _, err := r.LookPath("true && true"); err == nil {
		t.Fatal("LookPath ne doit pas résoudre une expression shell")
	}
}

func TestExecRunner_NonZeroExit(t *testing.T) {
	r := helperRunner(t)
	name, args := helperCall("exit")

	res := r.Run(context.Background(), name, args)
	if res.TimedOut {
		t.Fatal("timeout inattendu")
	}
	if res.ExitCode != 3 {
		t.Fatalf("ExitCode = %d, attendu 3 (err=%v)", res.ExitCode, res.Err)
	}
}
