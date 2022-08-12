package wait

import (
	"github.com/avast/retry-go"
	"github.com/pkg/errors"
	"loggie-test/pkg/resources"
	"time"
)

func ForReady(resource resources.Resource, option ...retry.Option) error {

	if len(option) == 0 {
		option = append(option, retry.Delay(1*time.Second))
		option = append(option, retry.Attempts(60))
	}

	if err := retry.Do(func() error {
		ready, err := resource.Ready()
		if err != nil {
			return retry.Unrecoverable(err)
		}

		if ready {
			return nil
		}

		return errors.Errorf(" %s not ready", resource.Name())
	}, option...); err != nil {
		return err
	}

	return nil
}

func Sleep(t time.Duration) {
	time.Sleep(t)
}
