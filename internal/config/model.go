package config

// Types de contrôles supportés en 0.1.0 (section 3.2 de la spec).
const (
	TypeCommand = "command"
	TypeDocker  = "docker"
	TypeEnv     = "env"
	TypePath    = "path"
)

// File représente le contenu décodé de envcheck.yml / .envcheck.yml.
type File struct {
	Version     int     `yaml:"version"`
	ProjectName string  `yaml:"projectName,omitempty"`
	Checks      []Check `yaml:"checks"`
}

// Check représente une règle déclarative. Tous les champs spécifiques à un
// type (command/docker/env/path) coexistent dans une seule struct plutôt
// qu'une hiérarchie polymorphe : plus simple à décoder depuis YAML, et la
// validation (validate.go) impose déjà les champs requis par type.
type Check struct {
	ID       string `yaml:"id"`
	Type     string `yaml:"type"`
	Required *bool  `yaml:"required,omitempty"`
	Hint     string `yaml:"hint,omitempty"`

	// type: command
	Command        string   `yaml:"command,omitempty"`
	Version        string   `yaml:"version,omitempty"`
	VersionArgs    []string `yaml:"versionArgs,omitempty"`
	VersionPattern string   `yaml:"versionPattern,omitempty"`

	// type: env
	EnvName    string `yaml:"envName,omitempty"`
	AllowEmpty bool   `yaml:"allowEmpty,omitempty"`

	// type: path
	Path string `yaml:"path,omitempty"`
	Kind string `yaml:"kind,omitempty"`
}

// IsRequired distingue la valeur par défaut (true) d'un `required: false`
// explicite, grâce au pointeur *bool sur le champ YAML.
func (c Check) IsRequired() bool {
	return c.Required == nil || *c.Required
}
