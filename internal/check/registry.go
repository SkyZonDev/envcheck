package check

// Registry associe un type de contrôle YAML ("env", "command", ...) au
// Checker qui sait l'évaluer.
type Registry struct {
	checkers map[string]Checker
}

func NewRegistry(checkers ...Checker) *Registry {
	r := &Registry{checkers: make(map[string]Checker, len(checkers))}
	for _, c := range checkers {
		r.checkers[c.Type()] = c
	}
	return r
}

func (r *Registry) Get(checkType string) (Checker, bool) {
	c, ok := r.checkers[checkType]
	return c, ok
}

// DefaultRegistry câble tous les Checkers du MVP (env, path, command, docker).
// Un type absent du registre devient "error" (ou "warn" si optionnel).
func DefaultRegistry() *Registry {
	return NewRegistry(
		NewEnvChecker(),
		NewPathChecker(),
		NewCommandChecker(),
		NewDockerChecker(),
	)
}
