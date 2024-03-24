package main

import (
	"polling_service/initiator"
	"polling_service/poller"
	"polling_service/scheduler"
)

func main() {
	go initiator.Init()
	go scheduler.Init()
	go poller.Init()
	select {}
}
