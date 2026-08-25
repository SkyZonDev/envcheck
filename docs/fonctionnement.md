# Comment fonctionne envcheck

Ce document décrit le comportement réel du binaire : ce qu’il lit, ce qu’il exécute, comment il tranche, et ce qu’il refuse de faire.

`envcheck` ne modifie jamais la machine. Il ne crée pas de fichiers (sauf `envcheck init`), n’installe pas d’outils, n’écrit pas de variables d’environnement et n’appelle aucun service distant.

## 1. Intention

Le fichier `envcheck.yml` est un **contrat** versionné avec le dépôt. Le binaire en est l’évaluateur :

- le mainteneur déclare *quoi* vérifier ;
- `envcheck` répond *est-ce vrai sur cette machine, maintenant* ;
- le contributeur (ou la CI) agit ensuite : installer Git, démarrer Docker, copier `.env.example`, etc.

Cette séparation est volontaire. Un outil qui « répare » l’environnement mélangerait diagnostic et effet de bord. Ici, le diagnostic est complet même si une règle échoue, et le `hint` de chaque règle sert de prochain pas humain.

## 2. Pipeline

Sans sous-commande, `envcheck` se comporte comme `envcheck check`. Le flux est unique :

```text
┌─────────────┐     ┌──────────────┐     ┌─────────────────┐     ┌────────┐     ┌──────────────┐
│  Découverte │ ──► │ Décodage YAML│ ──► │ Validation      │ ──► │ Moteur │ ──► │ Rendu + code │
│  du fichier │     │ (strict)     │     │ du schéma       │     │        │     │ de sortie    │
└─────────────┘     └──────────────┘     └─────────────────┘     └────────┘     └──────────────┘
```

Une erreur aux trois premières étapes (fichier absent, YAML illisible, schéma invalide) **n’exécute aucune règle** et retourne le code `2`. Une fois le fichier accepté, toutes les règles sont évaluées.

### 2.1 Découverte

Sans `--config` :

1. `envcheck.yml` dans le répertoire courant ;
2. sinon `.envcheck.yml` dans le même répertoire.

Le répertoire courant est celui d’où la commande est lancée (`os.Getwd()`), pas l’emplacement du binaire. **Aucun parent n’est consulté.** Un monorepo ou un dossier personnel qui contiendrait un `envcheck.yml` ne peut donc pas « contaminer » un dépôt enfant.

Avec `--config CHEMIN`, ce fichier est chargé tel quel. S’il n’existe pas, le code est `2`.

`envcheck init` utilise la même découverte pour refuser d’écraser un fichier existant, sauf `--force`. Le fichier écrit s’appelle toujours `envcheck.yml`.

### 2.2 Décodage strict

Le YAML est lu puis décodé avec `KnownFields(true)` : toute clé absente du schéma est une erreur. Ce n’est pas un avertissement. Un fichier copié-collé avec un champ d’une future version (ou une faute de frappe) échoue tout de suite, au lieu d’être ignoré silencieusement.

### 2.3 Validation de schéma

Après le décodage, `envcheck` vérifie ce que le parseur YAML ne peut pas garantir :

- `version` vaut exactement `1` (numéro de schéma, pas la version du binaire) ;
- `checks` n’est pas vide ;
- chaque `id` est unique et suit `[a-z0-9][a-z0-9-]{0,62}` ;
- `type` est l’un de `command`, `docker`, `env`, `path` ;
- les champs obligatoires du type sont présents ;
- une contrainte `version` SemVer est syntaxiquement valide ;
- un `versionPattern` compile comme expression régulière ;
- un `kind` de chemin vaut `file`, `directory`, `any`, ou est omis.

Si une seule de ces règles casse, **aucune** vérification n’est lancée (code `2`). Un `versionPattern` invalide n’est donc pas découvert au milieu d’un rapport : il est rejeté avant.

### 2.4 Évaluation

Les règles s’exécutent **dans l’ordre du fichier**, une par une. Une règle en échec, en avertissement ou en erreur interne n’empêche jamais d’évaluer les suivantes. L’ordre du rapport JSON et de la sortie texte est donc stable et comparable d’une machine à l’autre.

Chaque contrôle reçoit un *runtime* injecté : répertoire courant, copie des variables d’environnement, lanceur de processus, système de fichiers. Les contrôles ne parlent pas au terminal. Ils produisent un résultat ; le rendu vient après.

### 2.5 Rendu et code de sortie

- `--format text` (défaut) : une ligne par contrôle, un hint indenté en cas d’échec, un bilan « N/M contrôles requis réussis ».
- `--format json` : un seul objet JSON indenté sur `stdout`. Rien d’autre n’y est écrit.
- `--quiet` : masque les lignes `pass` dans le rendu texte. Le JSON reste complet.

Le processus est « passé » (`passed: true`, code `0`) si et seulement s’il n’y a **aucun** statut `fail` ni `error`. Les `warn` (échecs optionnels) ne font pas échouer le run.

## 3. Verdicts

Chaque contrôle produit un statut parmi quatre :

| Statut | Marqueur texte | Signification | Compte pour le code `1` ? |
| --- | --- | --- | --- |
| `pass` | `✓` | La règle est satisfaite | Non |
| `fail` | `✗` | Règle **obligatoire** non satisfaite | Oui |
| `warn` | `!` | Règle **optionnelle** (`required: false`) non satisfaite | Non |
| `error` | `‼` | Type inconnu du moteur, ou contrainte SemVer irrecevable à l’exécution | Oui, si la règle est obligatoire |

`required` vaut `true` par défaut. Il faut l’écrire `required: false` pour qu’un échec devienne un avertissement.

Conséquence pratique : vous pouvez déclarer un outil « souhaitable mais pas bloquant » (un linter, un client optionnel) sans faire échouer l’onboarding ni la CI.

## 4. Les quatre contrôles

### 4.1 `command`

Question : cet exécutable est-il dans le `PATH` ? Si une contrainte `version` est donnée, sa version extraite la respecte-t-elle ?

1. `LookPath(command)` — équivalent de « le binaire est-il trouvable ? ». Pas d’exécution encore.
2. Sans champ `version` : succès dès que le binaire existe.
3. Avec `version` : lancement de `command` + `versionArgs` (défaut : `--version`), délai 3 s.
4. Concaténation de stdout et stderr, suppression des séquences ANSI, extraction via `versionPattern` (défaut : premier triplet `X.Y.Z`).
5. Comparaison SemVer (bibliothèque Masterminds) contre la contrainte.

Échecs typiques :

- binaire absent du `PATH` ;
- commande trop lente (timeout) ;
- aucune version extractible ;
- version extraite hors plage (`Détectée : 18.20.4 ; attendue : >=20.0.0 <23.0.0.`).

Le champ `actual` du rapport contient la version extraite, jamais la sortie brute du processus.

Pour un binaire dont `--version` n’imprime pas un triplet classique (`go version go1.23.0 linux/arm64`), fournissez `versionArgs` et `versionPattern`. Voir [configuration.md](configuration.md#type-command).

### 4.2 `docker`

Question : le client `docker` est-il dans le `PATH`, et le démon répond-il à `docker info` ?

Les messages distinguent volontairement quatre situations :

| Situation | Résumé |
| --- | --- |
| `docker` introuvable | « Le client Docker est introuvable dans le PATH. » |
| Timeout (3 s) | « La commande Docker n’a pas répondu en moins de 3 s. » |
| Permission refusée | « Le client Docker est disponible, mais les permissions sont insuffisantes. » |
| Autre échec de `docker info` | « Le client Docker est disponible, mais le démon est inaccessible. » |

La détection « permissions » inspecte stdout/stderr **en interne** (motifs du type `permission denied`). Rien de cette sortie n’est recopié dans le rapport.

`envcheck` ne vérifie ni Compose, ni Kubernetes, ni la possibilité de tirer une image. Il répond seulement : « un démon Docker utilisable est-il là ? »

### 4.3 `env`

Question : cette variable existe-t-elle dans l’environnement du processus ? Est-elle non vide, sauf `allowEmpty: true` ?

Trois issues :

- absente → échec (ou avertissement si optionnel) ;
- présente mais vide, et `allowEmpty` est faux (défaut) → échec ;
- définie (éventuellement vide si autorisé) → succès.

Le champ `actual` reste **toujours vide**. La valeur n’est ni loguée, ni mise dans le JSON, ni imprimée en texte. Seul le *fait* (absente / vide / définie) est observable.

`envcheck` ne charge pas les fichiers `.env`. Si votre projet les utilise, exportez les variables (direnv, `set -a; source .env`, le runner CI, etc.) **avant** d’appeler `envcheck`.

### 4.4 `path`

Question : ce chemin existe-t-il, et son type correspond-il à `kind` ?

Les chemins relatifs sont résolus par rapport au répertoire courant du processus, pas par rapport au fichier YAML. Un `path: ./migrations` lancé depuis un sous-dossier cherchera `sous-dossier/migrations`.

| `kind` | Succès si |
| --- | --- |
| omis ou `any` | le chemin existe (fichier ou dossier) |
| `file` | c’est un fichier (pas un dossier) |
| `directory` | c’est un dossier |

Le champ `actual` vaut `file` ou `directory` — une indication de type, pas un contenu.

## 5. Exécution des commandes

Toute commande système passe par le même lanceur :

- `exec.CommandContext(nom, args...)` : l’exécutable et les arguments sont **séparés**. Aucun shell (`sh -c`, `cmd /C`) n’est invoqué. Un YAML malveillant du type `command: "node; rm -rf /"` cherche un binaire dont le nom contient `; rm -rf /`, il n’exécute pas un script.
- Délai dur de **3 secondes**. Au-delà, le processus enfant est tué ; le contrôle échoue (timeout), ce n’est pas une erreur interne.
- Capture plafonnée à **32 KiB** par flux (stdout et stderr). Le surplus est jeté pour ne pas bloquer l’enfant sur un pipe plein, et n’entre jamais dans un rapport.
- Les séquences ANSI sont retirées avant toute extraction de version.

Il n’existe pas de champ YAML « commande libre » (`versionCommand: "git --version | head"`). Seuls `command` (un nom) et `versionArgs` (une liste d’arguments) sont acceptés.

## 6. Ce qui n’est jamais observé depuis l’extérieur

| Donnée | Où elle circule | Où elle n’apparaît pas |
| --- | --- | --- |
| Valeur d’une variable d’environnement | Map interne au processus | Rapport JSON, sortie texte, `stderr`, `actual` |
| Sortie brute d’un binaire | Buffer plafonné, le temps d’extraire une version ou un motif Docker | Rapport JSON |
| Secrets dans le YAML | Le YAML est un fichier local que vous versionnez | `envcheck` n’y lit que des noms (`envName`), jamais des valeurs à afficher |

Le binaire ne fait aucun appel réseau, n’envoie pas de télémétrie, et n’écrit rien hors de `envcheck init`.

Voir aussi [SECURITY.md](../SECURITY.md).

## 7. Architecture interne

Le point d’entrée `cmd/envcheck` injecte la version compilée et délègue. Toute la logique vit sous `internal/` : aucune API Go publique n’est exposée pour l’instant.

```text
cmd/envcheck          binaire, version via ldflags
internal/app          Cobra : check, init, flags, codes de sortie
internal/config       découverte, YAML strict, validation
internal/core         orchestration, rapport, codes 0/1/2/3
internal/check        un Checker par type (command, docker, env, path)
internal/platform     os/exec sans shell, FS, délai, capture
internal/render       texte (couleur, quiet) et JSON
internal/template     modèle statique de envcheck init
```

Règle de conception : un Checker ne connaît ni Cobra, ni le terminal. Il reçoit un `Runtime` (CWD, env, runner, FS). Les tests substituent ces dépendances : la suite CI n’a pas besoin d’un démon Docker réel, ni des outils que le YAML du dépôt utilisateur déclarerait.

La version affichée par `--version` vaut `dev` en compilation locale. Les releases injectent le tag Git (`-X main.version=…`).

## 8. Limites du MVP

`envcheck` 0.1.0 ne fait pas :

- installer ou mettre à jour un outil ;
- charger `.env` / `.envrc` ;
- inspecter Docker Compose, Kubernetes, ou des images ;
- exécuter des scripts arbitraires ou des plugins distants ;
- détecter automatiquement la stack d’un dépôt ;
- remonter les répertoires parents à la recherche d’un YAML.

Ces limites sont le produit, pas un oubli : un diagnostic fiable, local et non intrusif.
