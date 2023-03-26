package main

import (
	"sync"

	"go.arsenm.dev/logger/log"
)

type WaitGroup struct {
	*sync.WaitGroup
}

func (wg WaitGroup) Done(c string) {
	log.Info("Component stopped").Str("name", c).Send()
	wg.WaitGroup.Done()
}
