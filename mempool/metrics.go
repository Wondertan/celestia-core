package mempool

import (
	"encoding/json"
	"path/filepath"
	"time"

	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/discard"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/tendermint/tendermint/libs/os"
)

const (
	// MetricsSubsystem is a subsystem shared by all metrics exposed by this
	// package.
	MetricsSubsystem = "mempool"
)

// Metrics contains metrics exposed by this package.
// see MetricsProvider for descriptions.
type Metrics struct {
	// Size of the mempool.
	Size metrics.Gauge

	// Histogram of transaction sizes, in bytes.
	TxSizeBytes metrics.Histogram

	// FailedTxs defines the number of failed transactions. These were marked
	// invalid by the application in either CheckTx or RecheckTx.
	FailedTxs metrics.Counter

	// EvictedTxs defines the number of evicted transactions. These are valid
	// transactions that passed CheckTx and existed in the mempool but were later
	// evicted to make room for higher priority valid transactions that passed
	// CheckTx.
	EvictedTxs metrics.Counter

	// SuccessfulTxs defines the number of transactions that successfully made
	// it into a block.
	SuccessfulTxs metrics.Counter

	// Number of times transactions are rechecked in the mempool.
	RecheckTimes metrics.Counter

	// AlreadySeenTxs defines the number of transactions that entered the
	// mempool which were already present in the mempool. This is a good
	// indicator of the degree of duplication in message gossiping.
	AlreadySeenTxs metrics.Counter
}

// PrometheusMetrics returns Metrics build using Prometheus client library.
// Optionally, labels can be provided along with their values ("foo",
// "fooValue").
func PrometheusMetrics(namespace string, labelsAndValues ...string) *Metrics {
	labels := []string{}
	for i := 0; i < len(labelsAndValues); i += 2 {
		labels = append(labels, labelsAndValues[i])
	}
	return &Metrics{
		Size: prometheus.NewGaugeFrom(stdprometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "size",
			Help:      "Size of the mempool (number of uncommitted transactions).",
		}, labels).With(labelsAndValues...),

		TxSizeBytes: prometheus.NewHistogramFrom(stdprometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "tx_size_bytes",
			Help:      "Transaction sizes in bytes.",
			Buckets:   stdprometheus.ExponentialBuckets(1, 3, 17),
		}, labels).With(labelsAndValues...),

		FailedTxs: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "failed_txs",
			Help:      "Number of failed transactions.",
		}, labels).With(labelsAndValues...),

		EvictedTxs: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "evicted_txs",
			Help:      "Number of evicted transactions.",
		}, labels).With(labelsAndValues...),

		SuccessfulTxs: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "successful_txs",
			Help:      "Number of transactions that successfully made it into a block.",
		}, labels).With(labelsAndValues...),

		RecheckTimes: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "recheck_times",
			Help:      "Number of times transactions are rechecked in the mempool.",
		}, labels).With(labelsAndValues...),

		AlreadySeenTxs: prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: MetricsSubsystem,
			Name:      "already_seen_txs",
			Help:      "Number of transactions that entered the mempool but were already present in the mempool.",
		}, labels).With(labelsAndValues...),
	}
}

// NopMetrics returns no-op Metrics.
func NopMetrics() *Metrics {
	return &Metrics{
		Size:           discard.NewGauge(),
		TxSizeBytes:    discard.NewHistogram(),
		FailedTxs:      discard.NewCounter(),
		EvictedTxs:     discard.NewCounter(),
		SuccessfulTxs:  discard.NewCounter(),
		RecheckTimes:   discard.NewCounter(),
		AlreadySeenTxs: discard.NewCounter(),
	}
}

type JSONMetrics struct {
	filepath             string
	StartedAt            time.Time
	EndedAt              time.Time
	FailedTxs            uint64
	EvictedTxs           uint64
	SuccessfulTxs        uint64
	AlreadySeenTxs       uint64
	AlreadyRejectedTxs   uint64
	RequestedTxs         uint64
	RerequestedTxs       uint64
	LostTxs			  	 uint64
	FailedResponses      uint64
	SentTransactionBytes uint64
	SentStateBytes       uint64
	ReceivedTxBytes      uint64
	ReceivedStateBytes   uint64
}

func NewJSONMetrics(rootDir string) *JSONMetrics {
	path := filepath.Join(rootDir, "data", "mempool_metrics.json")
	return &JSONMetrics{
		filepath:  path,
		StartedAt: time.Now().UTC(),
	}
}

func (m *JSONMetrics) Save() {
	m.EndedAt = time.Now().UTC()
	content, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	os.MustWriteFile(m.filepath, content, 0644)
}
