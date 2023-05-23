package metrics

import (
	"log"
	"net/http"
	"net/url"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var RdBlk = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_read",
	Help: "The total number of blocks read",
})

var RdBlkErr = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_read_error",
	Help: "The total number of corrupted or malformed blocks read",
})

var RdBlkMiss = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_read_missed",
	Help: "The total number of blocks missed between reads",
})

var RdBlkStale = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_read_stale",
	Help: "The total number of stale blocks read",
})

var RdBlkData = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_read_data",
	Help: "The total number of data blocks read",
})

var RdBlkKeep = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_read_keepalive",
	Help: "The total number of keepalive blocks read",
})

var RdErr = promauto.NewCounter(prometheus.CounterOpts{
	Name: "read_errors",
	Help: "The total number of read errors",
})

var RdSvcTime = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "read_service_time",
	Help:    "Read service time",
	Buckets: []float64{0.0005, 0.001, 0.005, 0.01, 0.015, 0.02, 0.03},
})

var WrBlk = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_written",
	Help: "The total number of blocks written",
})

var WrBlkData = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_written_data",
	Help: "The total number of data blocks written",
})

var WrBlkKeep = promauto.NewCounter(prometheus.CounterOpts{
	Name: "blocks_written_keepalive",
	Help: "The total number of keepalive blocks written",
})

var WrErr = promauto.NewCounter(prometheus.CounterOpts{
	Name: "write_errors",
	Help: "The total number of write errors",
})

var WrSvcTime = promauto.NewHistogram(prometheus.HistogramOpts{
	Name:    "write_service_time",
	Help:    "Write service time",
	Buckets: []float64{0.0005, 0.001, 0.005, 0.01, 0.015, 0.02, 0.03},
})

var RxPkt = promauto.NewCounter(prometheus.CounterOpts{
	Name: "packets_rx",
	Help: "The total number of recieved packets",
})

var RxBytes = promauto.NewCounter(prometheus.CounterOpts{
	Name: "bytes_rx",
	Help: "The total number of recieved bytes",
})

var RxErr = promauto.NewCounter(prometheus.CounterOpts{
	Name: "rx_errors",
	Help: "The total number of receive errors",
})

var TxPkt = promauto.NewCounter(prometheus.CounterOpts{
	Name: "packets_tx",
	Help: "The total number of transmitted packets",
})

var TxBytes = promauto.NewCounter(prometheus.CounterOpts{
	Name: "bytes_tx",
	Help: "The total number of sent bytes",
})

var TxErr = promauto.NewCounter(prometheus.CounterOpts{
	Name: "tx_errors",
	Help: "The total number of transmit errors",
})

var PeerRxState = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "peer_rx_state",
	Help: "Current peer rx state",
})

var PeerTxState = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "peer_tx_state",
	Help: "Current peer tx state",
})

func Serve(promAddr string) {
	promUrl := url.URL{
		Scheme: "http",
		Host:   promAddr,
		Path:   "/metrics",
	}
	http.Handle(promUrl.Path, promhttp.Handler())
	log.Printf("prometheus: client starting at %s", promUrl.String())
	err := http.ListenAndServe(promAddr, nil)
	if err != nil {
		log.Printf("prometheus: unable to start: %s", err)
	}
}
