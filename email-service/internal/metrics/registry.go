package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

// Registry is the custom prometheus registry for this service
var Registry = prometheus.NewRegistry()

func init() {
	// Register default collectors
	Registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	Registry.MustRegister(collectors.NewGoCollector())
}
