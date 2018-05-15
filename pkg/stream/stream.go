package stream

import (
	"fmt"

	"github.com/jojohappy/luxun/pkg/model"
)

type Stream struct {
	input chan *model.Event
	ops   []*Operator
	sink  *Sink
}

var defaultStream *Stream

func NewStream() *Stream {
	return &Stream{
		input: make(chan *model.Event, 1),
		ops:   make([]*Operator, 0),
	}
}

func Init() {
	defaultStream = NewStream()

	filterOp := NewOperator(filter)
	filterOp.SetInput(defaultStream.input)
	defaultStream.ops = append(defaultStream.ops, filterOp)

	storeOp := NewOperator(store)
	storeOp.SetInput(filterOp.GetOutput())
	defaultStream.ops = append(defaultStream.ops, storeOp)

	sink := NewSink()
	sink.SetInput(storeOp.GetOutput())
	defaultStream.sink = sink

	defaultStream.start()
}

func Process(ev ...*model.Event) {
	go func() {
		for _, e := range ev {
			defaultStream.input <- e
		}
	}()
}

func Stop() {
	defaultStream.stop()
}

func (s *Stream) start() {
	go func() {
		for _, op := range s.ops {
			op.Exec()
		}
		r := s.sink.Exec()
		for {
			select {
			case err := <-r:
				if nil != err {
					fmt.Printf("failed to sink %s\n", err.Error())
				}
			}
		}
	}()
}

func (s *Stream) stop() {
	for _, op := range s.ops {
		op.Stop()
	}

	s.sink.Stop()
}
