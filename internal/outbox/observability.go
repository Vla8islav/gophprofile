package outbox

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	outboxPublished = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gophprofile_outbox_published_total",
		Help: "Outbox events successfully published to the broker.",
	})
	outboxPublishFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gophprofile_outbox_publish_failures_total",
		Help: "Failed publish attempts; the event is retried next tick.",
	})
	outboxMarkFailures = promauto.NewCounter(prometheus.CounterOpts{
		Name: "gophprofile_outbox_mark_failures_total",
		Help: "Events published but not marked sent — each one becomes a duplicate delivery.",
	})
	outboxOldestPendingAge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "gophprofile_outbox_oldest_pending_age_seconds",
		Help: "Age of the oldest unpublished outbox event; 0 when the outbox is empty.",
	})
)
