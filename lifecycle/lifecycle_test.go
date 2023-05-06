package lifecycle

import (
	"context"
	"fmt"
	"testing"
)

func TestLifecycle(t *testing.T) {
	lf := new(Lifecycle)

	lf.Append(
		&Hook{
			Name: "a",
			OnStart: func(ctx context.Context) error {
				fmt.Println("program a start")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				fmt.Println("program a stop")
				return nil
			},
		},
		&Hook{
			Name: "b",
			OnStart: func(ctx context.Context) error {
				fmt.Println("program b start")
				return nil
			},
			OnStop: func(ctx context.Context) error {
				fmt.Println("program b stop")
				return nil
			},
		},
	)

	lf.Start(context.Background())
	lf.Stop(context.Background())
}
