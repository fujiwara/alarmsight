package main

import (
	"context"

	"github.com/fujiwara/alarmsight"
	"github.com/fujiwara/lamblocal"
)

func main() {
	app := alarmsight.NewCLI()
	lamblocal.Run(context.Background(), app.Handler)
}
