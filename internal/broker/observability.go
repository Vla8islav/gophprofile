package broker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel"
)

var consumerEvents = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "gophprofile_consumer_events_total",
	Help: "Consumed message handling outcomes; retry counts attempts, not messages.",
}, []string{"result"})

var tracer = otel.Tracer("gophprofile/broker")

type kafkaHeaderCarrier struct{ headers *[]kafka.Header }

func (c kafkaHeaderCarrier) Get(key string) string {
	for _, h := range *c.headers {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

func (c kafkaHeaderCarrier) Set(key, value string) {
	*c.headers = append(*c.headers, kafka.Header{Key: key, Value: []byte(value)})
}

func (c kafkaHeaderCarrier) Keys() []string {
	keys := make([]string, 0, len(*c.headers))
	for _, h := range *c.headers {
		keys = append(keys, h.Key)
	}
	return keys
}
