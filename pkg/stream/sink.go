package stream

import (
	"fmt"

	"github.com/jojohappy/luxun/pkg/handler/elasticsearch"
	"github.com/jojohappy/luxun/pkg/model"
)

type Sink struct {
	input  <-chan *model.Event
	stopCh chan struct{}
}

func NewSink() *Sink {
	return &Sink{
		stopCh: make(chan struct{}),
	}
}

func (s *Sink) SetInput(in <-chan *model.Event) {
	s.input = in
}

func (s *Sink) Exec() <-chan error {
	result := make(chan error)
	go func() {
		var err error
		for {
			select {
			case ev, opened := <-s.input:
				if !opened {
					continue
				}
				// send to es
				err = sendToEs(ev)
				if nil != err {
					fmt.Printf("failed to push event to elasticsearch: %s\n", err.Error())
					result <- err
				}
			case <-s.stopCh:
				close(result)
				return
			}
		}
	}()
	return result
}

func (s *Sink) Stop() {
	close(s.stopCh)
}

func sendToEs(ev *model.Event) error {
	return elasticsearch.ElasticSource().Bulk(ev)
}
