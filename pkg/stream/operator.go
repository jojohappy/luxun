package stream

import (
	"fmt"

	"github.com/luxun/pkg/model"
)

type opFunc func(en *model.Event) (*model.Event, error)

type Operator struct {
	input  <-chan *model.Event
	output chan *model.Event
	fn     opFunc
	stopCh chan struct{}
}

func NewOperator(fn opFunc) *Operator {
	return &Operator{
		output: make(chan *model.Event, 1),
		fn:     fn,
		stopCh: make(chan struct{}),
	}
}

func (o *Operator) Exec() {
	go func() {
		defer func() {
			close(o.output)
		}()
		var e *model.Event
		var err error
		for {
			select {
			case en, opened := <-o.input:
				if !opened {
					continue
				}
				e, err = o.fn(en)
				if nil != err {
					fmt.Printf("failed to process event: %s. skipped\n", err.Error())
				}
				o.output <- e
			case <-o.stopCh:
				return
			}
		}
	}()
}

func (o *Operator) SetInput(in <-chan *model.Event) {
	o.input = in
}

func (o *Operator) GetOutput() <-chan *model.Event {
	return o.output
}

func (o *Operator) Stop() {
	close(o.stopCh)
}

func filter(en *model.Event) (*model.Event, error) {
	return en, nil
}
