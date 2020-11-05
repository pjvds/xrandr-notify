package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/xgb"
	"github.com/BurntSushi/xgb/randr"
	"github.com/BurntSushi/xgb/xproto"
)

type Batch struct {
	Events []xgb.Event
}

func main() {
	log := log.New(os.Stderr, "", log.LstdFlags)

	x, err := xgb.NewConn()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	if err := randr.Init(x); err != nil {
		log.Fatal(err.Error())
		return
	}

	// Get the root window on the default screen.
	root := xproto.Setup(x).DefaultScreen(x).Root

	// Tell RandR to send us events. (I think these are all of them, as of 1.3.)
	if err := randr.SelectInputChecked(x, root,
		randr.NotifyMaskScreenChange|
			randr.NotifyMaskCrtcChange|
			randr.NotifyMaskOutputChange|
			randr.NotifyMaskOutputProperty).Check(); err != nil {
		log.Fatal(err.Error())
		return
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
			case <-time.After(1):
				break loop
			}
		}

		if err := encoder.Encode(batch); err != nil {
			log.Fatal(err)
		}
	}
}
