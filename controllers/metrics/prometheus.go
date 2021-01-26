package metrics

import (
	"fmt"
	"sync"

	k8gbv1beta1 "github.com/AbsaOSS/k8gb/api/v1beta1"
	"github.com/AbsaOSS/k8gb/controllers/depresolver"
	"github.com/prometheus/client_golang/prometheus"
	crm "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	gslbSubsystem   = "gslb"
	HealthyStatus   = "Healthy"
	UnhealthyStatus = "Unhealthy"
	NotFoundStatus  = "NotFound"
)

type PrometheusMetrics struct {
	healthyRecordsMetric        *prometheus.GaugeVec
	ingressHostsPerStatusMetric *prometheus.GaugeVec
	once                        sync.Once
}

// NewPrometheusMetrics creates new prometheus metrics instance
func NewPrometheusMetrics(config depresolver.Config) (metrics *PrometheusMetrics) {
	metrics = new(PrometheusMetrics)
	metrics.healthyRecordsMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: config.K8gbNamespace,
			Subsystem: gslbSubsystem,
			Name:      "healthy_records",
			Help:      "Number of healthy records observed by K8GB.",
		},
		[]string{"namespace", "name"},
	)
	metrics.ingressHostsPerStatusMetric = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: config.K8gbNamespace,
			Subsystem: gslbSubsystem,
			Name:      "ingress_hosts_per_status",
			Help:      "Number of managed hosts observed by K8GB.",
		},
		[]string{"namespace", "name", "status"},
	)
	return
}

func (m *PrometheusMetrics) UpdateIngressHostsPerStatusMetric(gslb *k8gbv1beta1.Gslb, serviceHealth map[string]string) error {
	var healthyHostsCount, unhealthyHostsCount, notFoundHostsCount int
	for _, hs := range serviceHealth {
		switch hs {
		case HealthyStatus:
			healthyHostsCount++
		case UnhealthyStatus:
			unhealthyHostsCount++
		default:
			notFoundHostsCount++
		}
	}
	m.ingressHostsPerStatusMetric.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": HealthyStatus}).
		Set(float64(healthyHostsCount))
	m.ingressHostsPerStatusMetric.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": UnhealthyStatus}).
		Set(float64(unhealthyHostsCount))
	m.ingressHostsPerStatusMetric.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name, "status": NotFoundStatus}).
		Set(float64(notFoundHostsCount))
	return nil
}

func (m *PrometheusMetrics) UpdateHealthyRecordsMetric(gslb *k8gbv1beta1.Gslb, healthyRecords map[string][]string) error {
	var hrsCount int
	for _, hrs := range healthyRecords {
		hrsCount += len(hrs)
	}
	m.healthyRecordsMetric.With(prometheus.Labels{"namespace": gslb.Namespace, "name": gslb.Name}).Set(float64(hrsCount))
	return nil
}

// Register prometheus metrics. Read register documentation, but shortly:
// You can register metric with given name only once
func (m *PrometheusMetrics) Register() (err error) {
	m.once.Do(func() {
		if err = crm.Registry.Register(m.healthyRecordsMetric); err != nil {
			return
		}
		if err = crm.Registry.Register(m.ingressHostsPerStatusMetric); err != nil {
			return
		}
	})
	if err != nil {
		return fmt.Errorf("can't register prometheus metrics: %s", err)
	}
	return
}

// Unregister prometheus metrics
func (m *PrometheusMetrics) Unregister() {
	crm.Registry.Unregister(m.healthyRecordsMetric)
	crm.Registry.Unregister(m.ingressHostsPerStatusMetric)
}

// GetHealthyRecordsMetric retrieves actual copy of healthy record metric
// TODO: consider to implement concrete metrics as a functions which returns metrics as slices/maps or structures
func (m *PrometheusMetrics) GetHealthyRecordsMetric() prometheus.GaugeVec {
	return *m.healthyRecordsMetric
}

// GetIngressHostsPerStatusMetric retrieves actual copy of ingress host metric
// TODO: consider to implement concrete metrics as a functions which returns metrics as slices/maps or structures
func (m *PrometheusMetrics) GetIngressHostsPerStatusMetric() prometheus.GaugeVec {
	return *m.ingressHostsPerStatusMetric
}
