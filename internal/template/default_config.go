package template

// DefaultConfig est le modèle statique écrit par `envcheck init`.
// Il illustre les quatre types du MVP ; l'utilisateur l'adapte à son dépôt.
const DefaultConfig = `# Généré par envcheck init. Adaptez les contrôles à votre dépôt.
# Les valeurs des variables d'environnement n'apparaissent jamais dans le rapport.
version: 1
projectName: my-project

checks:
  - id: git
    type: command
    command: git
    version: ">=2.40.0"
    hint: "Installez Git depuis https://git-scm.com/downloads."

  - id: docker-daemon
    type: docker
    hint: "Démarrez Docker Desktop ou le service Docker."

  - id: database-url
    type: env
    envName: DATABASE_URL
    hint: "Copiez .env.example vers .env et renseignez DATABASE_URL."

  - id: migrations
    type: path
    path: ./migrations
    kind: directory
`
