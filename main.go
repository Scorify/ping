package ping

import (
	"context"
	"fmt"

	"github.com/go-ping/ping"
	"github.com/scorify/schema"
)

type Schema struct {
	// Target is the address to ping
	Target string `key:"target"`

	// Count is the number of ping requests to send
	// Default is 1
	Count int `key:"count" default:"1"`

	// SuccessfulCount is the number of successful ping requests required
	// Default is 1
	// SuccessfulCount must be less than or equal to Count and greater than 0
	SuccessfulCount int `key:"successful_count" default:"1"`
}

func Validate(config string) error {
	conf := Schema{}

	err := schema.Unmarshal([]byte(config), &conf)
	if err != nil {
		return err
	}

	if conf.Count < 1 || conf.Count > 10 {
		return fmt.Errorf("invalid count; count must be between 1 and 10; count: %d", conf.Count)
	}

	if conf.SuccessfulCount < 1 || conf.SuccessfulCount > conf.Count {
		return fmt.Errorf("invalid successful_count; successful_count must be between 1 and count; successful_count: %d, count: %d", conf.SuccessfulCount, conf.Count)
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

		if stats.PacketsRecv < conf.SuccessfulCount {
			return fmt.Errorf("ping failed; received %d packets, expected at least %d", stats.PacketsRecv, conf.SuccessfulCount)
		}

		// ping successful
		return nil
	case <-ctx.Done():
		pinger.Stop()

		stats := pinger.Statistics()

		if stats.PacketsRecv < conf.SuccessfulCount {
			return fmt.Errorf("ping failed; received %d packets, expected at least %d", stats.PacketsRecv, conf.SuccessfulCount)
		}

		// ping successful
		return nil
	}
}
