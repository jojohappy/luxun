package stream

import (
	"testing"
	"time"

	"github.com/jojohappy/luxun/pkg/model"
)

var count int

func testOp(en *model.Event) (*model.Event, error) {
	count++
	return en, nil
}

func TestStream(t *testing.T) {
	count = 0

	Init()
	Stop()
	defaultStream = NewStream()

	op := NewOperator(testOp)
	op.SetInput(defaultStream.input)
	defaultStream.ops = append(defaultStream.ops, op)
	sink := NewSink()
	sink.SetInput(op.GetOutput())
	defaultStream.sink = sink
	defaultStream.start()

	ev := &model.Event{}

	Process(ev)
	time.Sleep(time.Second)
	Stop()

	if count != 1 {
		t.Fatalf("excepted 1, got %d", count)
	}
	count = 0
}
