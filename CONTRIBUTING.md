# Contribuer

Merci. Le périmètre du MVP est volontairement étroit : un binaire, un YAML,
un verdict. Les PRs qui restent dans ce cadre passent plus vite.

## Prérequis

Go 1.27 ou plus récent (voir `go.mod`). Docker n'est **pas** requis pour
lancer la suite de tests : le contrôle `docker` s'appuie sur un `Runner`
injecté.

## Commandes

```bash
go vet ./...
go test ./...
go build -o envcheck ./cmd/envcheck
```

Sur une pull request, la CI exécute les tests sur Linux, macOS et Windows,
puis GoReleaser en snapshot. Un job **sans** `setup-go` extrait l'archive
Linux amd64, vérifie `checksums.txt`, et lance le binaire (`--version`,
`init --dry-run`).

## Release

Réservé aux mainteneurs. Un tag semver `v*` déclenche
`.github/workflows/release.yml` :

```bash
git tag v0.1.0
git push origin v0.1.0
```

GoReleaser injecte la version via `-X main.version=…`, publie les
archives et `checksums.txt` sur GitHub Releases. Ne jamais forcer un
tag déjà poussé.

## Documentation

Le contrat utilisateur (CLI, YAML, codes de sortie, rapport JSON) est décrit
dans [`docs/`](docs/README.md). Un changement de comportement observable doit
y être reflété, pas seulement dans le code.

## Conventions

- La logique métier vit sous `internal/`. `cmd/envcheck` ne fait qu'injecter
  la version et appeler `app.Execute`.
- Les Checkers ne parlent pas au terminal. Ils reçoivent un `Runtime`
  (env, FS, Runner) pour rester testables sans la machine hôte.
- `os/exec` : exécutable et arguments séparés, jamais de shell.
- Aucune valeur de variable d'environnement dans un rapport, un test ou une
  issue. Utilisez des fixtures fictives.

## Signaler un bug

Utilisez le modèle GitHub. Retirez tokens, mots de passe et extraits `.env`
avant de coller une config ou une sortie JSON.
