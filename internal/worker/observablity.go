package worker

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var thumbnailDuration = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "gophprofile_worker_thumbnail_duration_seconds",
	Help:    "Time to generate all thumbnail variants for one avatar (successful runs only).",
	Buckets: prometheus.DefBuckets,
})
