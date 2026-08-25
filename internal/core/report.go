package core

import "github.com/SkyZonDev/envcheck/internal/check"

// Summary compte les résultats par statut, redondant avec Checks mais
// pratique pour un consommateur JSON qui ne veut pas recompter lui-même.
type Summary struct {
	Pass  int `json:"pass"`
	Fail  int `json:"fail"`
	Warn  int `json:"warn"`
	Error int `json:"error"`
}

// RunReport est le rapport immuable produit par core.Run. schemaVersion est
// distinct de la version du binaire : c'est le contrat public consommé par
// --format json en CI (section 9 de la spec).
type RunReport struct {
	SchemaVersion int            `json:"schemaVersion"`
	ToolVersion   string         `json:"toolVersion"`
	ConfigPath    string         `json:"configPath"`
	StartedAt     string         `json:"startedAt"`
	DurationMS    int64          `json:"durationMs"`
	Passed        bool           `json:"passed"`
	Summary       Summary        `json:"summary"`
	Checks        []check.Result `json:"checks"`
}
