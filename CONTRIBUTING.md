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

`main` est protégé : on n'y pousse pas. Tout passe par une pull request.

1. Branche depuis `main` à jour (`release/vX.Y.Z` ou `docs/…`).
2. Rédigez la section `## [X.Y.Z]` dans `CHANGELOG.md` **pour un humain** :
   ce qui change pour l'utilisateur, pas la liste des commits. Laissez
   `## [Unreleased]` vide au-dessus.
3. Ouvrez une PR vers `main`. La CI (tests + snapshot GoReleaser) doit passer.
4. Après le merge, depuis `main` à jour, poussez **uniquement le tag** :

```bash
git checkout main
git pull origin main
git tag -a v0.1.0 -m "v0.1.0"
git push origin v0.1.0
```

`.github/workflows/release.yml` se déclenche sur `v*`. Il extrait la
section CHANGELOG correspondant au tag (`scripts/extract-release-notes.sh`)
et la passe à GoReleaser : c'est ce texte qui apparaît dans « What's Changed »,
pas le journal git. S'il n'y a pas de section pour cette version, la release
échoue.

GoReleaser injecte la version via `-X main.version=…`, publie les archives
et `checksums.txt`. Ne jamais forcer un tag déjà poussé.

Pour corriger les notes d'une release déjà publiée (sans retaguer) :

```bash
bash scripts/extract-release-notes.sh v0.1.0 notes.md
gh release edit v0.1.0 --notes-file notes.md
```

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
