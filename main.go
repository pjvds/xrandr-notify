package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
	"github.com/urfave/cli/v2"
)

type Batch struct {
	Events []xgb.Event
}

func main() {
	log := log.New(os.Stderr, "", log.LstdFlags)
	app := &cli.App{
		Name: "randr-notify",
		Flags: []cli.Flag{
			&cli.DurationFlag{
				Name:    "accumulation-timeout",
				Aliases: []string{"a"},
				Usage:   "The maximum duration to wait to add another event to the batch.",
				Value:   1 * time.Second,
			},
		},
		Action: func(c *cli.Context) error {
			x, err := xgb.NewConn()
			if err != nil {
				return cli.Exit(err.Error(), 100)
			}

			if err := randr.Init(x); err != nil {
				log.Fatal(err.Error())
				return cli.Exit(err.Error(), 101)
			}

			// Get the root window on the default screen.
			root := xproto.Setup(x).DefaultScreen(x).Root

			// Tell RandR to send us events.
			if err := randr.SelectInputChecked(x, root,
				randr.NotifyMaskScreenChange|
					randr.NotifyMaskCrtcChange|
					randr.NotifyMaskOutputChange|
					randr.NotifyMaskOutputProperty).Check(); err != nil {
				return cli.Exit(err.Error(), 102)
			}

			encoder := json.NewEncoder(os.Stdout)
			events := make(chan xgb.Event)

			// Write x events to channel
			go func() {
				for {
					e, err := x.WaitForEvent()
					if err != nil {
						log.Fatal(err.Error())
						return
					}

					events <- e
				}
			}()

			// Monitor x events channel
			for e := range events {
				batch := Batch{
					Events: []xgb.Event{
						e,
					},
				}

				// accumulate events before write them as json output
			loop:
				for {
					select {
					case e, ok := <-events:
						if !ok {
							break
						}

						batch.Events = append(batch.Events, e)
					case <-time.After(c.Duration("accumulation-timeout")):
						break loop
					}
				}

				if err := encoder.Encode(batch); err != nil {
					log.Fatal(err)
				}
			}

			return nil
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
