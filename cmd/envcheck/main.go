package main

import (
	"os"

	"github.com/SkyZonDev/envcheck/internal/app"
)

// version est injectée au build via -ldflags "-X main.version=vX.Y.Z".
// En développement (go run / go build sans ldflags), elle vaut "dev".
var version = "dev"

func main() {
	os.Exit(app.Execute(version))
}