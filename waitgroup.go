package main

import (
	"log/slog"
	"sync"
)

type WaitGroup struct {
	*sync.WaitGroup
}

func (wg WaitGroup) Done(c string) {
	log.Info("Component stopped", slog.String("name", c))
	wg.WaitGroup.Done()
}
