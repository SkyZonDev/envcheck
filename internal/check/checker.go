package check

import (
	"context"
	"time"

	"github.com/SkyZonDev/envcheck/internal/config"
	"github.com/SkyZonDev/envcheck/internal/platform"
)

// Status reprend exactement les quatre valeurs du RunReport (section 4 de
// la spec) : un contrôle optionnel qui échoue devient "warn", jamais "fail".
type Status string

const (
	Pass  Status = "pass"
	Fail  Status = "fail"
	Warn  Status = "warn"
	Error Status = "error"
)

// Result est le résultat immuable d'un contrôle. Actual est réservé à une
// version détectée ou une indication inoffensive (ex: "directory") — pour
// un contrôle env, il reste toujours vide : la valeur d'une variable n'est
// jamais lue dans le rapport (section 4 de la spec).
type Result struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Required   bool   `json:"required"`
	Status     Status `json:"status"`
	Summary    string `json:"summary"`
	Detail     string `json:"detail,omitempty"`
	Hint       string `json:"hint,omitempty"`
	Expected   string `json:"expected,omitempty"`
	Actual     string `json:"actual,omitempty"`
	DurationMS int64  `json:"durationMs"`
}

// Runtime regroupe les dépendances dont un Checker a besoin pour s'exécuter,
// injectées plutôt qu'accédées globalement afin de rester testable sans la
// machine hôte (section 5 de la spec). Exec et FS proviennent de
// internal/platform ; nil retombe sur les adaptateurs OS réels.
type Runtime struct {
	CWD  string
	Env  map[string]string
	Exec platform.Runner
	FS   platform.FileSystem
	// Now permet de figer l'horloge dans les tests. nil -> time.Now().
	Now func() time.Time
}

func (rt Runtime) now() time.Time {
	if rt.Now != nil {
		return rt.Now()
	}
	return time.Now()
}

func (rt Runtime) runner() platform.Runner {
	if rt.Exec != nil {
		return rt.Exec
	}
	return platform.NewExecRunner()
}

func (rt Runtime) fs() platform.FileSystem {
	if rt.FS != nil {
		return rt.FS
	}
	return platform.OSFS{}
}

// Checker évalue une définition de contrôle d'un type donné et produit un
// Result. Un Checker ne doit jamais lire d'arguments CLI ni écrire dans le
// terminal (voir le tableau des responsabilités, section 4 de la spec).
type Checker interface {
	Type() string
	Run(ctx context.Context, def config.Check, rt Runtime) Result
}

// statusForFailure applique la règle commune à tous les types de contrôle :
// un échec sur une règle required: false est un avertissement, pas un échec.
func statusForFailure(def config.Check) Status {
	if def.IsRequired() {
		return Fail
	}
	return Warn
}
