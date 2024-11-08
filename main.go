package main

import (
	"context"
	"errors"
	"log"
	"os"

	"github.com/urfave/cli/v2"
)

func main() {
	ctx := context.Background()

	app := &cli.App{
		Name:      "transliterate-markdown-attachments",
		Usage:     "utility for transliterating the names of markdown attachment files and updating links in the markup",
		UsageText: "transliterate-markdown-attachments convert -f /path/to/file.md",
		Commands: []*cli.Command{
			{
				Name: "convert",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "file",
						Aliases:  []string{"f"},
						Usage:    "path to markdown file",
						Required: true,
					},
				},
				Action: func(cliCtx *cli.Context) error {
					filePath, ok := cliCtx.Value("file").(string)
					if !ok {
						return errors.New("flag file is not string")
					}

					return convert(ctx, filePath)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
