package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

// Represents observable action instance
type CounterAction struct {
	labels []string
	vec    *prometheus.CounterVec
}

func (a CounterAction) Inc(status Status) {
	a.vec.
		WithLabelValues(append([]string{status.String()}, a.labels...)...).
		Inc()
}

func (a CounterAction) Add(status Status, value float64) {
	a.vec.
		WithLabelValues(append([]string{status.String()}, a.labels...)...).
		Add(value)
}

// Represents counter metric
type Counter struct {
	vec              *prometheus.CounterVec
	additionalLabels []string
}

// Returns new Counter instance
func NewCounter(name, help string, additionalLabels []string) *Counter {
	labels := append([]string{"status"}, additionalLabels...)
	m := Counter{
		vec: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: name,
				Help: help,
			},
			labels,
		),
		additionalLabels: additionalLabels,
	}
	prometheus.MustRegister(m.vec)
	return &m
}

// Returns new Action for Counter metric collector
// Panics if the labels count is invalid
func (m Counter) NewAction(labels ...string) *CounterAction {
	if len(labels) != len(m.additionalLabels) {
		panic(fmt.Errorf("action and metric labels count mismatch: %d != %d", len(labels), len(m.additionalLabels)))
	}
	return &CounterAction{
		labels: labels,
		vec:    m.vec,
	}
}
