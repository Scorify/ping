package ping

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/go-ping/ping"
)

type Schema struct {
	Target string `key:"target"`
}

func Validate(config string) error {
	conf := Schema{}

	err := schema.Unmarshal([]byte(config), &conf)
	if err != nil {
		return err
	}

	return nil
}

func Run(ctx context.Context, config string) error {
	schema := Schema{}

	err := json.Unmarshal([]byte(config), &schema)
	if err != nil {
		return err
	}

	// create pinger
	pinger, err := ping.NewPinger(schema.Target)
	if err != nil {
		return fmt.Errorf("failed to create pinger; err: %s", err)
	}

	pinger.Count = 1
	doneChan := make(chan error, 1)

	// run ping
	go func() {
		defer close(doneChan)
		doneChan <- pinger.Run()
	}()

	// handle ping output
	select {
	case err := <-doneChan:
		if err != nil {
			// error occurred during ping
			return err
		}

		stats := pinger.Statistics()

		if stats.PacketLoss == 0 {
			// successful ping
			return nil
		} else {
			// packet loss not 0
			return fmt.Errorf("packet loss not 0; packet loss: %f", stats.PacketLoss)
		}

	case <-ctx.Done():
		pinger.Stop()

		// context exceed; timeout
		return fmt.Errorf("timeout exceeded; err: %s", ctx.Err())
	}
}
