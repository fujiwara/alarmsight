package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/fujiwara/alarmsight"
	"github.com/fujiwara/lamblocal"
	"github.com/handlename/ssmwrap"
)

func main() {
	if p := os.Getenv("SSM_PATH"); p != "" {
		slog.Info(fmt.Sprintf("exporting SSM parameters from %s", p))
		if err := ssmwrap.Export(ssmwrap.ExportOptions{
			Paths:   []string{p},
			Retries: 3,
		}); err != nil {
			slog.Error(fmt.Sprintf("failed to export SSM parameters: %s", err))
			os.Exit(1)
		}
	}
	app := alarmsight.NewCLI()
	lamblocal.Run(context.Background(), app.Handler)
}
