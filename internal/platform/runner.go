package platform

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"time"
)

// MaxCaptureBytes plafonne stdout et stderr séparément. Au-delà, les
// octets sont jetés (pour que le processus enfant ne se bloque pas sur
// un pipe plein) et n'entrent jamais dans un rapport.
const MaxCaptureBytes = 32 * 1024

// RunResult est le résultat brut d'un sous-processus. Aucun Checker ne
// doit copier Stdout/Stderr dans le rapport JSON : ils servent uniquement
// à extraire une version, après StripANSI.
type RunResult struct {
	ExitCode int
	Stdout   []byte
	Stderr   []byte
	TimedOut bool
	Err      error
}

// Runner encapsule la recherche d'un exécutable et son lancement. Les
// arguments sont toujours une slice distincte : jamais de shell, jamais
// de `sh -c` / `cmd /C` (section 6 de la spec).
type Runner interface {
	LookPath(file string) (string, error)
	Run(ctx context.Context, name string, args []string) RunResult
}

// ExecRunner est l'implémentation réelle, basée sur os/exec.
type ExecRunner struct {
	Timeout    time.Duration // 0 -> CommandTimeout
	MaxCapture int           // 0 -> MaxCaptureBytes
	// extraEnv est concaténé à os.Environ() du processus enfant. Réservé
	// aux tests du helper process ; la production le laisse nil.
	extraEnv []string
}

// NewExecRunner construit un runner aux constantes de la spec (3 s, 32 KiB).
func NewExecRunner() *ExecRunner {
	return &ExecRunner{}
}

func (r *ExecRunner) LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

func (r *ExecRunner) Run(ctx context.Context, name string, args []string) RunResult {
	timeout := r.Timeout
	if timeout <= 0 {
		timeout = CommandTimeout
	}
	limit := r.MaxCapture
	if limit <= 0 {
		limit = MaxCaptureBytes
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// name et args sont transmis séparément : exec.CommandContext n'invoque
	// jamais de shell, même si name contient des espaces ou des métacaractères.
	cmd := exec.CommandContext(ctx, name, args...)
	if len(r.extraEnv) > 0 {
		cmd.Env = append(os.Environ(), r.extraEnv...)
	}

	stdout := newCappedBuffer(limit)
	stderr := newCappedBuffer(limit)
	cmd.Stdout = stdout
	cmd.Stderr = stderr

	err := cmd.Run()
	res := RunResult{
		Stdout: stdout.bytes(),
		Stderr: stderr.bytes(),
	}

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		res.TimedOut = true
		res.Err = ctx.Err()
		return res
	}
	if err != nil {
		res.Err = err
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			res.ExitCode = exit.ExitCode()
		}
		return res
	}
	if cmd.ProcessState != nil {
		res.ExitCode = cmd.ProcessState.ExitCode()
	}
	return res
}

// cappedBuffer accepte toutes les écritures (pour ne pas bloquer l'enfant)
// mais n'en conserve que les `limit` premiers octets.
type cappedBuffer struct {
	buf   bytes.Buffer
	limit int
}

func newCappedBuffer(limit int) *cappedBuffer {
	return &cappedBuffer{limit: limit}
}

func (c *cappedBuffer) Write(p []byte) (int, error) {
	remaining := c.limit - c.buf.Len()
	if remaining > 0 {
		if len(p) > remaining {
			c.buf.Write(p[:remaining])
		} else {
			c.buf.Write(p)
		}
	}
	return len(p), nil
}

func (c *cappedBuffer) bytes() []byte {
	out := make([]byte, c.buf.Len())
	copy(out, c.buf.Bytes())
	return out
}
