package main

import (
	"os"

	sdk "github.com/PlakarKorp/go-kloset-sdk"
	"github.com/PlakarKorp/integrations/github"
)

func main() {
	sdk.EntrypointImporter(os.Args, github.NewImporter)
}
