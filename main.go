package main

import (
	"github.com/rull05/AnvBOT-go/anv"
)

func main() {
	bot, err := anv.NewAnv(nil)
	if err != nil {
		panic(err)
	}

	err = bot.Start()
	if err != nil {
		panic(err)
	}
	stop := make(chan struct{})
	<-stop
}
