package main

import (
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name:  "HashHelper",
		Usage: "A tool to help with the hash",
		Commands: []*cli.Command{
			{
				Name:    "dwarf",
				Aliases: []string{"d"},
				Usage:   "dwarf to c header file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "input",
						Aliases:     []string{"i"},
						Usage:       "input file path",
						Value:       "./",
						DefaultText: "./",
					},
					&cli.StringFlag{
						Name:        "output",
						Aliases:     []string{"o"},
						Usage:       "output file path",
						Value:       "./",
						DefaultText: "./",
					},
				},
				Action: func(c *cli.Context) error {
					ipath := c.String("input")
					//opath := c.String("output")
					err := DwarfHelper(ipath)
					if err != nil {
						return err
					}
					return nil
				},
			},
			{
				Name:    "pdb",
				Aliases: []string{"p"},
				Usage:   "pdb to c header file",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "path",
						Aliases:     []string{"p"},
						Usage:       "specify path",
						Value:       "./",
						DefaultText: "./",
					},
					&cli.StringFlag{
						Name:        "remove",
						Aliases:     []string{"r"},
						Usage:       "remove .fs256 file after verify",
						Value:       "false",
						DefaultText: "false",
					},
				},
				Action: func(c *cli.Context) error {
					ipath := c.String("input")
					opath := c.String("output")
					log.Println(ipath, opath)
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
