# Configuration YAML

`envcheck` lit un seul fichier : `envcheck.yml` ou `.envcheck.yml` dans le répertoire courant, ou le chemin passé à `--config`.

Le schéma porte le numéro `version: 1`, indépendant de la version du binaire. Les clés inconnues sont **refusées**.

## Fichier

```yaml
version: 1                 # obligatoire, seule valeur acceptée : 1
projectName: api-platform  # optionnel, informatif (n'apparaît pas dans le rapport JSON actuel)

checks:                    # obligatoire, au moins une entrée
  - id: git
    type: command
    command: git
    version: ">=2.40.0"
    hint: "Installez Git depuis https://git-scm.com/downloads."
```

| Champ | Obligatoire | Description |
| --- | --- | --- |
| `version` | oui | Numéro de schéma. Doit valoir `1`. |
| `projectName` | non | Libellé du projet. Ignoré par le moteur ; utile pour les humains qui lisent le YAML. |
| `checks` | oui | Liste ordonnée des règles. L’ordre d’évaluation et d’affichage est celui de cette liste. |

## Champs communs à chaque contrôle

```yaml
- id: mon-controle          # obligatoire, unique
  type: command             # obligatoire : command | docker | env | path
  required: true            # optionnel, défaut true
  hint: "Que faire ensuite." # optionnel, affiché seulement en cas de non-réussite
```

| Champ | Obligatoire | Description |
| --- | --- | --- |
| `id` | oui | Identifiant unique. Alphabet : `[a-z0-9][a-z0-9-]{0,62}` (1 à 63 caractères, commence par une lettre ou un chiffre, tirets ensuite). Sert de clé dans le rapport JSON. |
| `type` | oui | `command`, `docker`, `env` ou `path`. |
| `required` | non | `true` (défaut) : un échec devient `fail` et code de sortie `1`. `false` : un échec devient `warn` et n’empêche pas le code `0`. |
| `hint` | non | Conseil affiché sous la ligne d’échec (texte) et recopié dans le JSON. Pas évalué, pas exécuté. |

Les champs spécifiques à un type peuvent figurer dans le YAML ; s’ils n’appartiennent pas au type, ils sont simplement ignorés après validation (ils restent des clés *connues* du schéma). Inversement, une clé vraiment inconnue (`versoin:`, `cmd:`, …) fait échouer le chargement.

## Type `command`

Vérifie qu’un exécutable est dans le `PATH`, et optionnellement que sa version SemVer satisfait une contrainte.

```yaml
- id: node
  type: command
  command: node
  version: ">=20.0.0 <23.0.0"
  versionArgs: ["--version"]
  versionPattern: "([0-9]+\\.[0-9]+\\.[0-9]+)"
  hint: "Installez Node 20 depuis https://nodejs.org/."
```

| Champ | Obligatoire | Défaut | Description |
| --- | --- | --- | --- |
| `command` | oui | — | Nom de l’exécutable, tel que `LookPath` le chercherait (`node`, `git`, `go`). Pas une ligne de shell. |
| `version` | non | (aucune) | Contrainte [SemVer](https://semver.org/) Masterminds, par ex. `>=2.40.0`, `>=20.0.0 <23.0.0`, `~1.2`. Absent : seule la présence du binaire compte. |
| `versionArgs` | non | `["--version"]` | Arguments passés **séparément** au binaire pour obtenir la version. |
| `versionPattern` | non | `([0-9]+\.[0-9]+\.[0-9]+)` | Expression régulière Go. Si un groupe de capture existe, c’est lui qui est pris ; sinon le match entier. |

### Extraire une version atypique

`go version` imprime `go version go1.23.0 linux/arm64`, pas `1.23.0` isolé. Le motif par défaut (triplet `X.Y.Z`) échouerait. Il faut donc :

```yaml
- id: go
  type: command
  command: go
  version: ">=1.22.0"
  versionArgs: ["version"]
  versionPattern: "go([0-9]+\\.[0-9]+(?:\\.[0-9]+)?)"
  hint: "Installez Go depuis https://go.dev/dl/."
```

Le YAML est un fichier YAML : les backslashes du motif doivent être échappés (`\\.`).

Une contrainte `version` ou un `versionPattern` **invalides** (non compilables) sont rejetés au chargement, code `2`, avant toute exécution.

## Type `docker`

Aucun champ spécifique. Le contrôle lance `docker info`.

```yaml
- id: docker-daemon
  type: docker
  required: true
  hint: "Démarrez Docker Desktop ou le service Docker."
```

Messages selon la cause : client absent, démon inaccessible, permissions insuffisantes, timeout. Voir [fonctionnement.md](fonctionnement.md#42-docker).

## Type `env`

Vérifie la **présence** d’une variable, jamais sa valeur.

```yaml
- id: database-url
  type: env
  envName: DATABASE_URL
  allowEmpty: false
  hint: "Copiez .env.example vers .env et renseignez DATABASE_URL."
```

| Champ | Obligatoire | Défaut | Description |
| --- | --- | --- | --- |
| `envName` | oui | — | Nom de la variable (`DATABASE_URL`, `CI`, …). |
| `allowEmpty` | non | `false` | Si `true`, une variable définie mais vide réussit. Si `false`, vide = échec. |

`envcheck` lit l’environnement du **processus** (`os.Environ()`). Il ne parse pas `.env`. Pour tester une variable locale :

```bash
export DATABASE_URL=postgres://localhost/app
envcheck
```

## Type `path`

Vérifie qu’un fichier ou un dossier existe.

```yaml
- id: migrations
  type: path
  path: ./migrations
  kind: directory
  hint: "Créez le dossier migrations à la racine du dépôt."
```

| Champ | Obligatoire | Défaut | Description |
| --- | --- | --- | --- |
| `path` | oui | — | Chemin relatif (au répertoire courant d’exécution) ou absolu. |
| `kind` | non | `any` | `file`, `directory` ou `any`. Valeur vide = `any`. |

Un `kind` autre que ces trois valeurs est une erreur de schéma (code `2`).

Les chemins relatifs sont joints au CWD du processus. Lancez `envcheck` depuis la racine du dépôt, ou passez `--config` *et* placez-vous au bon endroit.

## Contrôles optionnels

```yaml
- id: optional-linter
  type: command
  command: golangci-lint
  required: false
  hint: "golangci-lint est recommandé mais non bloquant."
```

Si `golangci-lint` est absent : statut `warn`, marqueur `!`, code de sortie `0` (pourvu que le reste passe).

## Ce qui est rejeté

| Situation | Code | Exemple de message |
| --- | --- | --- |
| Aucun YAML dans le CWD | `2` | aucun fichier envcheck.yml ou .envcheck.yml trouvé… |
| Clé inconnue | `2` | YAML invalide : … field xxx not found … |
| `version:` autre que `1` | `2` | version N non supportée |
| `checks` vide | `2` | aucun contrôle déclaré |
| `id` dupliqué | `2` | id dupliqué : "git" |
| `id` hors motif | `2` | l'id doit respecter `[a-z0-9][a-z0-9-]{0,62}` |
| `type` inconnu | `2` | type "plugin" inconnu |
| `command` / `envName` / `path` manquant | `2` | le type X requiert le champ Y |
| Contrainte SemVer illisible | `2` | contrainte de version invalide |
| `versionPattern` non compilable | `2` | versionPattern est invalide |
| `kind` hors liste | `2` | kind doit être file, directory ou any |

## Exemples

Un modèle prêt à adapter : [`examples/envcheck.yml`](../examples/envcheck.yml).

Le dépôt `envcheck` lui-même utilise un YAML plus court (`envcheck.yml` à la racine) : Go présent avec une version minimale, et `go.mod` en fichier.
