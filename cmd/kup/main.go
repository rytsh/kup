package main

import (
	"context"

	"github.com/rakunlabs/into"
)

func main() {
	into.Init(
		run,
		into.WithMsgf("kup"),
		into.WithStartFn(nil), into.WithStopFn(nil),
	)
}

func run(ctx context.Context) error {
	return nil
}
