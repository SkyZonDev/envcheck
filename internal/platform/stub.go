package platform

import (
	"context"
	"os/exec"
)

// StubRunner est un Runner scripté pour les tests des Checkers command /
// docker. Il n'exécute rien : LookPath et Run renvoient ce qu'on a préparé.
type StubRunner struct {
	// LookPathErr, si non nil, fait échouer LookPath (typiquement exec.ErrNotFound).
	LookPathErr error
	// Result est renvoyé tel quel par Run.
	Result RunResult
	// RunCalled passe à true dès que Run est invoqué — pour vérifier qu'un
	// contrôle sans contrainte de version ne lance pas le binaire.
	RunCalled    bool
	LookPathName string
	RunName      string
	RunArgs      []string
}

func (s *StubRunner) LookPath(file string) (string, error) {
	s.LookPathName = file
	if s.LookPathErr != nil {
		return "", s.LookPathErr
	}
	return "/stub/" + file, nil
}

func (s *StubRunner) Run(_ context.Context, name string, args []string) RunResult {
	s.RunCalled = true
	s.RunName = name
	s.RunArgs = append([]string(nil), args...)
	if s.LookPathErr != nil && s.Result.Err == nil {
		return RunResult{Err: exec.ErrNotFound}
	}
	return s.Result
}
