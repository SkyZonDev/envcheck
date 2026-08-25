package platform

import "time"

// CommandTimeout est le délai dur appliqué à chaque sous-processus
// (section 1 et 6 de la spec). Au-delà, le Runner tue le processus et
// signale TimedOut : le Checker traduit ça en fail, pas en erreur interne.
const CommandTimeout = 3 * time.Second

// Clock isole time.Now pour figer l'horloge dans les tests.
type Clock interface {
	Now() time.Time
}

type systemClock struct{}

func (systemClock) Now() time.Time { return time.Now() }

// SystemClock renvoie l'horloge murale du processus.
func SystemClock() Clock { return systemClock{} }

// FrozenClock renvoie toujours Instant. Utile pour des tests déterministes
// du calcul de DurationMS.
type FrozenClock struct {
	Instant time.Time
}

func (c FrozenClock) Now() time.Time { return c.Instant }
