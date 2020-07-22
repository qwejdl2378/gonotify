package main

import (
	"fmt"
	"github.com/qwejdl2378/gonotify/tail"
)

func main() {
	t := tail.New("/workspace/tmp/openresty-config/logs/access.log")
	go t.StartTrack()
	for {
		select {
			case t := <- t.Lines:
				fmt.Println(t.Text)
		}
	}
}
