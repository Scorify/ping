package ping

import (
	"context"
	"fmt"

	"github.com/go-ping/ping"
	"github.com/scorify/schema"
)

type Schema struct {
	Target string `key:"target"`
	Count int `key:"ping count"`
}

func Validate(config string) error {
	conf := Schema{}

	err := schema.Unmarshal([]byte(config), &conf)
	if err != nil {
		return err
	}

	if conf.Count < 1 || conf.Count >= 1000 {
		return fmt.Errorf("ping count must be between 1 and 1000")
	}
	
	return nil
}

func Run(ctx context.Context, config string) error {
	conf := Schema{}

	err := schema.Unmarshal([]byte(config), &conf)
	if err != nil {
		return err
	}

	// create pinger
	pinger, err := ping.NewPinger(conf.Target)
	if err != nil {
		return fmt.Errorf("failed to create pinger; err: %s", err)
	}

	pinger.Count = conf.Count
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

		if stats.PacketLoss < 100 {
			// at least one successful ping
			return nil
		} else {
			// packet loss is 100%
			return fmt.Errorf("total packet loss; packet loss: %f", stats.PacketLoss)
		}

	case <-ctx.Done():
		pinger.Stop()

		// context exceed; timeout
		return fmt.Errorf("timeout exceeded; err: %s", ctx.Err())
	}
}
