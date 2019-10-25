package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var defaultObj = map[float64]float64{0.5: 0.05, 0.75: 0.05, 0.95: 0.05, 0.99: 0.05}

// Represents observable action instance
type Action struct {
	ts     time.Time
	labels []string
	vec    *prometheus.SummaryVec
}

func (a Action) Observe(status Status) {
	a.vec.
		WithLabelValues(append([]string{status.String()}, a.labels...)...).
		Observe(float64(time.Now().Sub(a.ts).Nanoseconds()))
}

// Represents latency metric
type Latency struct {
	latency          *prometheus.SummaryVec
	additionalLabels []string
}

// Returns new Latency instance
func NewLatency(name, help string, obj map[float64]float64, additionalLabels []string) *Latency {
	if obj == nil {
		obj = defaultObj
	}
	labels := append([]string{"status"}, additionalLabels...)
	m := Latency{
		latency: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       name,
				Help:       help,
				Objectives: obj,
			},
			labels,
		),
		additionalLabels: additionalLabels,
	}
	prometheus.MustRegister(m.latency)
	return &m
}

func (l Latency) NewAction(labels ...string) *Action {
	if len(labels) != len(l.additionalLabels) {
		panic(fmt.Errorf("action and metric labels count mismatch: %d != %d", len(labels), len(l.additionalLabels)))
	}
	return &Action{
		ts:     time.Now(),
		labels: labels,
		vec:    l.latency,
	}
}
