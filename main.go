package main

import (
	"github.com/luxun/pkg/controller"
	"github.com/luxun/pkg/stream"
)

func main() {
	stream.Init()

	controller.Execute()
}
