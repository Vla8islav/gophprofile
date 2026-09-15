package broker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var consumerEvents = promauto.NewCounterVec(prometheus.CounterOpts{
	Name: "gophprofile_consumer_events_total",
	Help: "Consumed message handling outcomes; retry counts attempts, not messages.",
}, []string{"result"})
