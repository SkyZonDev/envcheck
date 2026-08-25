package core

// Codes de sortie du CLI, tels que définis dans la spécification technique.
const (
	// ExitOK : exécution valide, tous les contrôles obligatoires réussissent.
	ExitOK = 0
	// ExitFail : exécution valide, au moins un contrôle obligatoire échoue.
	ExitFail = 1
	// ExitConfigError : fichier absent, YAML invalide, schéma invalide ou erreur d'usage.
	ExitConfigError = 2
	// ExitInternalError : erreur interne inattendue.
	ExitInternalError = 3
)
