package main

import (
	"github.com/jojohappy/luxun/pkg/controller"
	"github.com/jojohappy/luxun/pkg/stream"
)

func main() {
	stream.Init()

	controller.Execute()
}
