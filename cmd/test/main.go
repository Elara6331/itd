package main

import (
	"fmt"
	"time"

	"go.arsenm.dev/itd/api"
)

func main() {
	itd, _ := api.New(api.DefaultAddr)
	defer itd.Close()

	fmt.Println(itd.Address())

	mCh, cancel, _ := itd.WatchMotion()

	go func() {
		time.Sleep(10 * time.Second)
		cancel()
		fmt.Println("canceled")
	}()

	for m := range mCh {
		fmt.Println(m)
	}

}
