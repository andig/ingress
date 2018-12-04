package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/urfave/cli.v1"

	. "github.com/sevlyar/go-daemon"
)

func main() {
	app := cli.NewApp()
	app.Name = "ingress"
	app.Usage = "data ingestion for volkszaehler.org"
	app.HideVersion = true
	app.Flags = []cli.Flag{
		// general
		cli.BoolFlag{
			Name: "daemonize, d",
		},
	}

	app.Action = func(c *cli.Context) {
		fmt.Println(os.Args)
		if c.NArg() > 0 {
			log.Fatalf("Unexpected arguments: %v", c.Args())
		}

		log.Println("ingress - data ingestion for volkszaehler.org")

		if c.Bool("daemonize") {
			// if true {
			fmt.Println("daemonizing")

			// run in background
			// context := new(Context)
			context := &Context{
				PidFileName: "ingress.pid",
				PidFilePerm: 0644,
				LogFileName: "ingress.log",
				LogFilePerm: 0640,
				WorkDir:     "./",
				// WorkDir:     "/Users/andig/htdocs/ingress",
				Umask: 027,
				Args:  []string{"ingress", "-d"},
			}

			child, err := context.Reborn()
			if err != nil {
				log.Fatal("Unable to run: ", err)
			}

			if child != nil {
				log.Println("Daemonized")
				// PostParent()
			} else {
				defer context.Release()
				// PostChild()
				log.Println("Daemon running")
			}
		} else {
			// run in foreground
			fmt.Println("not daemonizing")
			// PostChild()
		}
	}

	app.Run(os.Args)
}
