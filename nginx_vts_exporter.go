package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type NginxVts struct {
	NginxVersion string `json:"nginxVersion"`
	LoadMsec     int64  `json:"loadMsec"`
	NowMsec      int64  `json:"nowMsec"`
	Connections  struct {
		Active   int `json:"active"`
		Reading  int `json:"reading"`
		Writing  int `json:"writing"`
		Waiting  int `json:"waiting"`
		Accepted int `json:"accepted"`
		Handled  int `json:"handled"`
		Requests int `json:"requests"`
	} `json:"connections"`
	ServerZones   map[string]Server     `json:"serverZones"`
	UpstreamZones map[string][]Upstream `json:"upstreamZones"`
	CacheZones    map[string]Cache      `json:"cacheZones"`
}

type Server struct {
	RequestCounter int `json:"requestCounter"`
	InBytes        int `json:"inBytes"`
	OutBytes       int `json:"outBytes"`
	Responses      struct {
		OneXx       int `json:"1xx"`
		TwoXx       int `json:"2xx"`
		ThreeXx     int `json:"3xx"`
		FourXx      int `json:"4xx"`
		FiveXx      int `json:"5xx"`
		Miss        int `json:"miss"`
		Bypass      int `json:"bypass"`
		Expired     int `json:"expired"`
		Stale       int `json:"stale"`
		Updating    int `json:"updating"`
		Revalidated int `json:"revalidated"`
		Hit         int `json:"hit"`
		Scarce      int `json:"scarce"`
	} `json:"responses"`
	OverCounts struct {
		MaxIntegerSize float64 `json:"maxIntegerSize"`
		RequestCounter int     `json:"requestCounter"`
		InBytes        int     `json:"inBytes"`
		OutBytes       int     `json:"outBytes"`
		OneXx          int     `json:"1xx"`
		TwoXx          int     `json:"2xx"`
		ThreeXx        int     `json:"3xx"`
		FourXx         int     `json:"4xx"`
		FiveXx         int     `json:"5xx"`
		Miss           int     `json:"miss"`
		Bypass         int     `json:"bypass"`
		Expired        int     `json:"expired"`
		Stale          int     `json:"stale"`
		Updating       int     `json:"updating"`
		Revalidated    int     `json:"revalidated"`
		Hit            int     `json:"hit"`
		Scarce         int     `json:"scarce"`
	} `json:"overCounts"`
}

type Upstream struct {
	Server         string `json:"server"`
	RequestCounter int    `json:"requestCounter"`
	InBytes        int    `json:"inBytes"`
	OutBytes       int    `json:"outBytes"`
	Responses      struct {
		OneXx   int `json:"1xx"`
		TwoXx   int `json:"2xx"`
		ThreeXx int `json:"3xx"`
		FourXx  int `json:"4xx"`
		FiveXx  int `json:"5xx"`
	} `json:"responses"`
	ResponseMsec int  `json:"responseMsec"`
	Weight       int  `json:"weight"`
	MaxFails     int  `json:"maxFails"`
	FailTimeout  int  `json:"failTimeout"`
	Backup       bool `json:"backup"`
	Down         bool `json:"down"`
	OverCounts   struct {
		MaxIntegerSize float64 `json:"maxIntegerSize"`
		RequestCounter int     `json:"requestCounter"`
		InBytes        int     `json:"inBytes"`
		OutBytes       int     `json:"outBytes"`
		OneXx          int     `json:"1xx"`
		TwoXx          int     `json:"2xx"`
		ThreeXx        int     `json:"3xx"`
		FourXx         int     `json:"4xx"`
		FiveXx         int     `json:"5xx"`
	} `json:"overCounts"`
}

type Cache struct {
	MaxSize   int `json:"maxSize"`
	UsedSize  int `json:"usedSize"`
	InBytes   int `json:"inBytes"`
	OutBytes  int `json:"outBytes"`
	Responses struct {
		Miss        int `json:"miss"`
		Bypass      int `json:"bypass"`
		Expired     int `json:"expired"`
		Stale       int `json:"stale"`
		Updating    int `json:"updating"`
		Revalidated int `json:"revalidated"`
		Hit         int `json:"hit"`
		Scarce      int `json:"scarce"`
	} `json:"responses"`
	OverCounts struct {
		MaxIntegerSize float64 `json:"maxIntegerSize"`
		InBytes        int     `json:"inBytes"`
		OutBytes       int     `json:"outBytes"`
		Miss           int     `json:"miss"`
		Bypass         int     `json:"bypass"`
		Expired        int     `json:"expired"`
		Stale          int     `json:"stale"`
		Updating       int     `json:"updating"`
		Revalidated    int     `json:"revalidated"`
		Hit            int     `json:"hit"`
		Scarce         int     `json:"scarce"`
	} `json:"overCounts"`
}

const namespace = "nginx"

type Exporter struct {
	URI   string
	mutex sync.RWMutex

	serverMetrics, upstreamMetrics, cacheMetrics map[string]*prometheus.GaugeVec
}

func newServerMetric(metricName string, docString string, labels []string, namespace string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "server_" + metricName,
			Help:      docString,
		},
		labels,
	)
}

func newUpstreamMetric(metricName string, docString string, labels []string, namespace string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "upstream_" + metricName,
			Help:      docString,
		},
		labels,
	)
}

func newCacheMetric(metricName string, docString string, labels []string, namespace string) *prometheus.GaugeVec {
	return prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Name:      "cache_" + metricName,
			Help:      docString,
		},
		labels,
	)
}

func NewExporter(uri string, namespace string) *Exporter {

	return &Exporter{
		URI: uri,
		serverMetrics: map[string]*prometheus.GaugeVec{
			"connections": newServerMetric("connections", "nginx connections", []string{"status"}, namespace),
			"requests":    newServerMetric("requests", "requests counter", []string{"host", "code"}, namespace),
			"bytes":       newServerMetric("bytes", "request/response bytes", []string{"host", "direction"}, namespace),
			"cache":       newServerMetric("cache", "cache counter", []string{"host", "status"}, namespace),
		},
		upstreamMetrics: map[string]*prometheus.GaugeVec{
			"requests": newUpstreamMetric("requests", "requests counter", []string{"upstream", "code"}, namespace),
			"bytes":    newUpstreamMetric("bytes", "request/response bytes", []string{"upstream", "direction"}, namespace),
		},
		cacheMetrics: map[string]*prometheus.GaugeVec{
			"requests": newCacheMetric("requests", "cache requests counter", []string{"zone", "status"}, namespace),
			"bytes":    newCacheMetric("bytes", "cache request/response bytes", []string{"zone", "direction"}, namespace),
		},
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.serverMetrics {
		m.Describe(ch)
	}
	for _, m := range e.upstreamMetrics {
		m.Describe(ch)
	}
	for _, m := range e.cacheMetrics {
		m.Describe(ch)
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.resetMetrics()
	e.scrape()

	e.collectMetrics(ch)
}

func (e *Exporter) resetMetrics() {
	for _, m := range e.serverMetrics {
		m.Reset()
	}
	for _, m := range e.upstreamMetrics {
		m.Reset()
	}
	for _, m := range e.cacheMetrics {
		m.Reset()
	}
}

func (e *Exporter) collectMetrics(metrics chan<- prometheus.Metric) {
	for _, m := range e.serverMetrics {
		m.Collect(metrics)
	}
	for _, m := range e.upstreamMetrics {
		m.Collect(metrics)
	}
	for _, m := range e.cacheMetrics {
		m.Collect(metrics)
	}
}

func (e *Exporter) scrape() {
	body, err := fetchHTTP(e.URI, 2*time.Second)()
	if err != nil {
		log.Println("fetchHTTP failed", err)
		return
	}
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		log.Println("ioutil.ReadAll failed", err)
		return
	}

	var nginxVtx NginxVts
	err = json.Unmarshal(data, &nginxVtx)
	if err != nil {
		log.Println("json.Unmarshal failed", err)
		return
	}

	e.serverMetrics["connections"].WithLabelValues("active").Set(float64(nginxVtx.Connections.Active))
	e.serverMetrics["connections"].WithLabelValues("reading").Set(float64(nginxVtx.Connections.Reading))
	e.serverMetrics["connections"].WithLabelValues("waiting").Set(float64(nginxVtx.Connections.Waiting))
	e.serverMetrics["connections"].WithLabelValues("writing").Set(float64(nginxVtx.Connections.Writing))
	e.serverMetrics["connections"].WithLabelValues("accepted").Set(float64(nginxVtx.Connections.Accepted))
	e.serverMetrics["connections"].WithLabelValues("handled").Set(float64(nginxVtx.Connections.Handled))
	e.serverMetrics["connections"].WithLabelValues("requests").Set(float64(nginxVtx.Connections.Requests))

	for host, s := range nginxVtx.ServerZones {
		e.serverMetrics["requests"].WithLabelValues(host, "total").Set(float64(s.RequestCounter))
		e.serverMetrics["requests"].WithLabelValues(host, "1xx").Set(float64(s.Responses.OneXx))
		e.serverMetrics["requests"].WithLabelValues(host, "2xx").Set(float64(s.Responses.TwoXx))
		e.serverMetrics["requests"].WithLabelValues(host, "3xx").Set(float64(s.Responses.ThreeXx))
		e.serverMetrics["requests"].WithLabelValues(host, "4xx").Set(float64(s.Responses.FourXx))
		e.serverMetrics["requests"].WithLabelValues(host, "5xx").Set(float64(s.Responses.FiveXx))

		e.serverMetrics["cache"].WithLabelValues(host, "bypass").Set(float64(s.Responses.Bypass))
		e.serverMetrics["cache"].WithLabelValues(host, "expired").Set(float64(s.Responses.Expired))
		e.serverMetrics["cache"].WithLabelValues(host, "hit").Set(float64(s.Responses.Hit))
		e.serverMetrics["cache"].WithLabelValues(host, "miss").Set(float64(s.Responses.Miss))
		e.serverMetrics["cache"].WithLabelValues(host, "revalidated").Set(float64(s.Responses.Revalidated))
		e.serverMetrics["cache"].WithLabelValues(host, "scarce").Set(float64(s.Responses.Scarce))
		e.serverMetrics["cache"].WithLabelValues(host, "stale").Set(float64(s.Responses.Stale))
		e.serverMetrics["cache"].WithLabelValues(host, "updating").Set(float64(s.Responses.Updating))

		e.serverMetrics["bytes"].WithLabelValues(host, "in").Set(float64(s.InBytes))
		e.serverMetrics["bytes"].WithLabelValues(host, "out").Set(float64(s.OutBytes))
	}

	for name, upstreamList := range nginxVtx.UpstreamZones {
		for _, s := range upstreamList {
			e.upstreamMetrics["requests"].WithLabelValues(name, "total").Add(float64(s.RequestCounter))
			e.upstreamMetrics["requests"].WithLabelValues(name, "1xx").Add(float64(s.Responses.OneXx))
			e.upstreamMetrics["requests"].WithLabelValues(name, "2xx").Add(float64(s.Responses.TwoXx))
			e.upstreamMetrics["requests"].WithLabelValues(name, "3xx").Add(float64(s.Responses.ThreeXx))
			e.upstreamMetrics["requests"].WithLabelValues(name, "4xx").Add(float64(s.Responses.FourXx))
			e.upstreamMetrics["requests"].WithLabelValues(name, "5xx").Add(float64(s.Responses.FiveXx))

			e.upstreamMetrics["bytes"].WithLabelValues(name, "in").Add(float64(s.InBytes))
			e.upstreamMetrics["bytes"].WithLabelValues(name, "out").Add(float64(s.OutBytes))
		}
	}

	for zone, s := range nginxVtx.CacheZones {
		e.cacheMetrics["requests"].WithLabelValues(zone, "bypass").Set(float64(s.Responses.Bypass))
		e.cacheMetrics["requests"].WithLabelValues(zone, "expired").Set(float64(s.Responses.Expired))
		e.cacheMetrics["requests"].WithLabelValues(zone, "hit").Set(float64(s.Responses.Hit))
		e.cacheMetrics["requests"].WithLabelValues(zone, "miss").Set(float64(s.Responses.Miss))
		e.cacheMetrics["requests"].WithLabelValues(zone, "revalidated").Set(float64(s.Responses.Revalidated))
		e.cacheMetrics["requests"].WithLabelValues(zone, "scarce").Set(float64(s.Responses.Scarce))
		e.cacheMetrics["requests"].WithLabelValues(zone, "stale").Set(float64(s.Responses.Stale))
		e.cacheMetrics["requests"].WithLabelValues(zone, "updating").Set(float64(s.Responses.Updating))

		e.cacheMetrics["bytes"].WithLabelValues(zone, "in").Set(float64(s.InBytes))
		e.cacheMetrics["bytes"].WithLabelValues(zone, "out").Set(float64(s.OutBytes))
	}
}

func fetchHTTP(uri string, timeout time.Duration) func() (io.ReadCloser, error) {
	http.DefaultClient.Timeout = timeout

	return func() (io.ReadCloser, error) {
		resp, err := http.DefaultClient.Get(uri)
		if err != nil {
			return nil, err
		}
		if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
			resp.Body.Close()
			return nil, fmt.Errorf("HTTP status %d", resp.StatusCode)
		}
		return resp.Body, nil
	}
}

var (
	listenAddress   = flag.String("telemetry.address", ":9913", "Address on which to expose metrics.")
	metricsEndpoint = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metrics.")
	metricsNamespace = flag.String("metrics.namespace", "nginx", "Prometheus metrics namespace.")
	nginxScrapeURI  = flag.String("nginx.scrape_uri", "http://localhost/status", "URI to nginx stub status page")
	insecure        = flag.Bool("insecure", true, "Ignore server certificate if using https")
)

func main() {
	flag.Parse()

	exporter := NewExporter(*nginxScrapeURI, *metricsNamespace)
	prometheus.MustRegister(exporter)
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle(*metricsEndpoint, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Nginx Exporter</title></head>
			<body>
			<h1>Nginx Exporter</h1>
			<p><a href="` + *metricsEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Starting Server at : %s", *listenAddress)
	log.Printf("Metrics endpoint: %s", *metricsEndpoint)
	log.Printf("Metrics namespace: %s", *metricsNamespace)
	log.Printf("Scraping information from : %s", *nginxScrapeURI)
	log.Fatal(http.ListenAndServe(*listenAddress, nil))
}
