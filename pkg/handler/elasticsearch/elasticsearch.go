package elasticsearch

import (
	"context"
	"flag"
	"fmt"
	"reflect"
	"sync"
	"time"

	elastic "gopkg.in/olivere/elastic.v5"
)

const (
	defaultBulkAction        = 1000
	defaultBulkSize          = 3
	defaultBulkFlushInterval = 10 * time.Second
	defaultMaxRetries        = 3
	defaultBulkWorker        = 3
	defaultHealthcheck       = true
)

var (
	defaultIndex  = flag.String("es-index", "kubernetes-events", "index name of kubernetes events")
	defaultEsUrls = flag.String("es-urls", "", "Endpoints of Events backend elasticsearch")
)

type ElasticClient struct {
	Client        *elastic.Client
	bulkProcessor *elastic.BulkProcessor
	index         string
	healthcheck   bool
	maxRetries    int
	urls          []string
	q             chan interface{}
	shutdownCh    chan struct{}
}

var (
	esClient *ElasticClient
	once     sync.Once
)

func NewElasticStorage(index string, urls ...string) (*ElasticClient, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(urls...),
		elastic.SetHealthcheck(defaultHealthcheck),
		elastic.SetMaxRetries(defaultMaxRetries))
	if nil != err {
		return nil, err
	}

	bp, err := client.BulkProcessor().
		Name("Luxun-Elastic").
		Workers(defaultBulkWorker).
		BulkActions(defaultBulkAction).
		BulkSize(defaultBulkSize).
		FlushInterval(defaultBulkFlushInterval).
		Stats(true).
		Do(context.Background())
	if nil != err {
		return nil, err
	}

	return &ElasticClient{
		Client:        client,
		bulkProcessor: bp,
		index:         index,
		healthcheck:   defaultHealthcheck,
		maxRetries:    defaultMaxRetries,
		urls:          urls,
		q:             make(chan interface{}, defaultBulkAction),
		shutdownCh:    make(chan struct{}),
	}, nil
}

func (es *ElasticClient) Run() {
	go func() {
		es.runQueueRoutine()
	}()
}

func (es *ElasticClient) Shutdown() {
	close(es.shutdownCh)
	es.bulkProcessor.Close()
}

func (es *ElasticClient) runQueueRoutine() {
	defer close(es.q)
	var t time.Time
	for {
		select {
		case ev, ok := <-es.q:
			if ok {
				v := reflect.ValueOf(ev)
				if v.Kind() == reflect.Struct {
					tv := v.FieldByName("Time")
					if tv.IsValid() {
						t = tv.Interface().(time.Time)
						es.bulkProcessor.Add(elastic.NewBulkIndexRequest().Index(fmt.Sprintf("%s-%s", es.index, t.Format("2006.01.02"))).Type(es.index).Doc(ev))
					} else {
						fmt.Printf("no time filed was found in %v\n", ev)
						es.bulkProcessor.Add(elastic.NewBulkIndexRequest().Index(es.index).Type(es.index).Doc(ev))
					}
				} else {
					fmt.Printf("kind of %v is not struct\n", ev)
				}
			}
		case <-es.shutdownCh:
			return
		}
	}
}

func (es *ElasticClient) Bulk(v ...interface{}) error {
	var err error
	func() {
		for _, ev := range v {
			select {
			case es.q <- ev:
			default:
				err = fmt.Errorf("elasticsearch queue blocked")
			}
		}
	}()
	if err != nil {
		return err
	}
	return nil
}

func ElasticSource() *ElasticClient {
	flag.Parse()
	var err error
	once.Do(func() {
		esClient, err = NewElasticStorage(*defaultIndex, *defaultEsUrls)
		if nil != err {
			panic(err)
		}
		esClient.Run()
	})
	return esClient
}
