package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
)

type NginxVts struct {
	HostName     string `json:"hostName"`
	NginxVersion string `json:"nginxVersion"`
	LoadMsec     int64  `json:"loadMsec"`
	NowMsec      int64  `json:"nowMsec"`
	Connections  struct {
		Active   uint64 `json:"active"`
		Reading  uint64 `json:"reading"`
		Writing  uint64 `json:"writing"`
		Waiting  uint64 `json:"waiting"`
		Accepted uint64 `json:"accepted"`
		Handled  uint64 `json:"handled"`
		Requests uint64 `json:"requests"`
	} `json:"connections"`
	ServerZones   map[string]Server              `json:"serverZones"`
	UpstreamZones map[string][]Upstream          `json:"upstreamZones"`
	FilterZones   map[string]map[string]Upstream `json:"filterZones"`
	CacheZones    map[string]Cache               `json:"cacheZones"`
}

type Server struct {
	RequestCounter uint64 `json:"requestCounter"`
	InBytes        uint64 `json:"inBytes"`
	OutBytes       uint64 `json:"outBytes"`
	RequestMsec    uint64 `json:"requestMsec"`
	Responses      struct {
		OneXx       uint64 `json:"1xx"`
		TwoXx       uint64 `json:"2xx"`
		ThreeXx     uint64 `json:"3xx"`
		FourXx      uint64 `json:"4xx"`
		FiveXx      uint64 `json:"5xx"`
		Miss        uint64 `json:"miss"`
		Bypass      uint64 `json:"bypass"`
		Expired     uint64 `json:"expired"`
		Stale       uint64 `json:"stale"`
		Updating    uint64 `json:"updating"`
		Revalidated uint64 `json:"revalidated"`
		Hit         uint64 `json:"hit"`
		Scarce      uint64 `json:"scarce"`
	} `json:"responses"`
	OverCounts struct {
		MaxIntegerSize float64 `json:"maxIntegerSize"`
		RequestCounter uint64  `json:"requestCounter"`
		InBytes        uint64  `json:"inBytes"`
		OutBytes       uint64  `json:"outBytes"`
		OneXx          uint64  `json:"1xx"`
		TwoXx          uint64  `json:"2xx"`
		ThreeXx        uint64  `json:"3xx"`
		FourXx         uint64  `json:"4xx"`
		FiveXx         uint64  `json:"5xx"`
		Miss           uint64  `json:"miss"`
		Bypass         uint64  `json:"bypass"`
		Expired        uint64  `json:"expired"`
		Stale          uint64  `json:"stale"`
		Updating       uint64  `json:"updating"`
		Revalidated    uint64  `json:"revalidated"`
		Hit            uint64  `json:"hit"`
		Scarce         uint64  `json:"scarce"`
	} `json:"overCounts"`
}

type Upstream struct {
	Server         string `json:"server"`
	RequestCounter uint64 `json:"requestCounter"`
	InBytes        uint64 `json:"inBytes"`
	OutBytes       uint64 `json:"outBytes"`
	Responses      struct {
		OneXx   uint64 `json:"1xx"`
		TwoXx   uint64 `json:"2xx"`
		ThreeXx uint64 `json:"3xx"`
		FourXx  uint64 `json:"4xx"`
		FiveXx  uint64 `json:"5xx"`
	} `json:"responses"`
	ResponseMsec uint64 `json:"responseMsec"`
	RequestMsec  uint64 `json:"requestMsec"`
	Weight       uint64 `json:"weight"`
	MaxFails     uint64 `json:"maxFails"`
	FailTimeout  uint64 `json:"failTimeout"`
	Backup       bool   `json:"backup"`
	Down         bool   `json:"down"`
	OverCounts   struct {
		MaxIntegerSize float64 `json:"maxIntegerSize"`
		RequestCounter uint64  `json:"requestCounter"`
		InBytes        uint64  `json:"inBytes"`
		OutBytes       uint64  `json:"outBytes"`
		OneXx          uint64  `json:"1xx"`
		TwoXx          uint64  `json:"2xx"`
		ThreeXx        uint64  `json:"3xx"`
		FourXx         uint64  `json:"4xx"`
		FiveXx         uint64  `json:"5xx"`
	} `json:"overCounts"`
}

type Cache struct {
	MaxSize   uint64 `json:"maxSize"`
	UsedSize  uint64 `json:"usedSize"`
	InBytes   uint64 `json:"inBytes"`
	OutBytes  uint64 `json:"outBytes"`
	Responses struct {
		Miss        uint64 `json:"miss"`
		Bypass      uint64 `json:"bypass"`
		Expired     uint64 `json:"expired"`
		Stale       uint64 `json:"stale"`
		Updating    uint64 `json:"updating"`
		Revalidated uint64 `json:"revalidated"`
		Hit         uint64 `json:"hit"`
		Scarce      uint64 `json:"scarce"`
	} `json:"responses"`
	OverCounts struct {
		MaxIntegerSize float64 `json:"maxIntegerSize"`
		InBytes        uint64  `json:"inBytes"`
		OutBytes       uint64  `json:"outBytes"`
		Miss           uint64  `json:"miss"`
		Bypass         uint64  `json:"bypass"`
		Expired        uint64  `json:"expired"`
		Stale          uint64  `json:"stale"`
		Updating       uint64  `json:"updating"`
		Revalidated    uint64  `json:"revalidated"`
		Hit            uint64  `json:"hit"`
		Scarce         uint64  `json:"scarce"`
	} `json:"overCounts"`
}

type Exporter struct {
	URI string

	infoMetric                                                  *prometheus.Desc
	serverMetrics, upstreamMetrics, filterMetrics, cacheMetrics map[string]*prometheus.Desc
}

func newServerMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(*metricsNamespace, "server", metricName),
		docString, labels, nil,
	)
}

func newUpstreamMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(*metricsNamespace, "upstream", metricName),
		docString, labels, nil,
	)
}

func newFilterMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(*metricsNamespace, "filter", metricName),
		docString, labels, nil,
	)
}

func newCacheMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(*metricsNamespace, "cache", metricName),
		docString, labels, nil,
	)
}

func NewExporter(uri string) *Exporter {
	return &Exporter{
		URI:        uri,
		infoMetric: newServerMetric("info", "nginx info", []string{"hostName", "nginxVersion"}),
		serverMetrics: map[string]*prometheus.Desc{
			"connections": newServerMetric("connections", "nginx connections", []string{"status"}),
			"requests":    newServerMetric("requests", "requests counter", []string{"host", "code"}),
			"bytes":       newServerMetric("bytes", "request/response bytes", []string{"host", "direction"}),
			"cache":       newServerMetric("cache", "cache counter", []string{"host", "status"}),
			"requestMsec": newServerMetric("requestMsec", "average of request processing times in milliseconds", []string{"host"}),
		},
		upstreamMetrics: map[string]*prometheus.Desc{
			"requests":     newUpstreamMetric("requests", "requests counter", []string{"upstream", "code", "backend"}),
			"bytes":        newUpstreamMetric("bytes", "request/response bytes", []string{"upstream", "direction"}),
			"responseMsec": newUpstreamMetric("responseMsec", "average of only upstream/backend response processing times in milliseconds", []string{"upstream", "backend"}),
			"requestMsec":  newUpstreamMetric("requestMsec", "average of request processing times in milliseconds", []string{"upstream", "backend"}),
		},
		filterMetrics: map[string]*prometheus.Desc{
			"requests":     newFilterMetric("requests", "requests counter", []string{"filter", "filterName", "code"}),
			"bytes":        newFilterMetric("bytes", "request/response bytes", []string{"filter", "filterName", "direction"}),
			"responseMsec": newFilterMetric("responseMsec", "average of only upstream/backend response processing times in milliseconds", []string{"filter", "filterName"}),
			"requestMsec":  newFilterMetric("requestMsec", "average of request processing times in milliseconds", []string{"filter", "filterName"}),
		},
		cacheMetrics: map[string]*prometheus.Desc{
			"requests": newCacheMetric("requests", "cache requests counter", []string{"zone", "status"}),
			"bytes":    newCacheMetric("bytes", "cache request/response bytes", []string{"zone", "direction"}),
		},
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range e.serverMetrics {
		ch <- m
	}
	for _, m := range e.upstreamMetrics {
		ch <- m
	}
	for _, m := range e.filterMetrics {
		ch <- m
	}
	for _, m := range e.cacheMetrics {
		ch <- m
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	body, err := fetchHTTP(e.URI, time.Duration(*nginxScrapeTimeout)*time.Second)()
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

	// info
	uptime := (nginxVtx.NowMsec - nginxVtx.LoadMsec)/1000
	ch <- prometheus.MustNewConstMetric(e.infoMetric, prometheus.GaugeValue, float64(uptime), nginxVtx.HostName, nginxVtx.NginxVersion)

	// connections
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Active), "active")
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Reading), "reading")
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Waiting), "waiting")
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Writing), "writing")
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Accepted), "accepted")
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Handled), "handled")
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Requests), "requests")

	// ServerZones
	for host, s := range nginxVtx.ServerZones {
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.RequestCounter), host, "total")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.OneXx), host, "1xx")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.TwoXx), host, "2xx")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.ThreeXx), host, "3xx")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.FourXx), host, "4xx")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.FiveXx), host, "5xx")

		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Bypass), host, "bypass")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Expired), host, "expired")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Hit), host, "hit")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Miss), host, "miss")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Revalidated), host, "revalidated")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Scarce), host, "scarce")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Stale), host, "stale")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Updating), host, "updating")

		ch <- prometheus.MustNewConstMetric(e.serverMetrics["bytes"], prometheus.CounterValue, float64(s.InBytes), host, "in")
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["bytes"], prometheus.CounterValue, float64(s.OutBytes), host, "out")

		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requestMsec"], prometheus.GaugeValue, float64(s.RequestMsec), host)

	}

	// UpstreamZones
	for name, upstreamList := range nginxVtx.UpstreamZones {
		var total, one, two, three, four, five, inbytes, outbytes float64
		for _, s := range upstreamList {
			total += float64(s.RequestCounter)
			one += float64(s.Responses.OneXx)
			two += float64(s.Responses.TwoXx)
			three += float64(s.Responses.ThreeXx)
			four += float64(s.Responses.FourXx)
			five += float64(s.Responses.FiveXx)

			inbytes += float64(s.InBytes)
			outbytes += float64(s.OutBytes)

			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["responseMsec"], prometheus.GaugeValue, float64(s.ResponseMsec), name, s.Server)
			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requestMsec"], prometheus.GaugeValue, float64(s.RequestMsec), name, s.Server)

			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, total, name, "total", s.Server)
			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, one, name, "1xx", s.Server)
			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, two, name, "2xx", s.Server)
			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, three, name, "3xx", s.Server)
			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, four, name, "4xx", s.Server)
			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, five, name, "5xx", s.Server)
		}

		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["bytes"], prometheus.CounterValue, inbytes, name, "in")
		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["bytes"], prometheus.CounterValue, outbytes, name, "out")
	}
	// FilterZones
	for filter, values := range nginxVtx.FilterZones {
		for name, stat := range values {
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["responseMsec"], prometheus.GaugeValue, float64(stat.ResponseMsec), filter, name)
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["requestMsec"], prometheus.GaugeValue, float64(stat.RequestMsec), filter, name)
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["requests"], prometheus.CounterValue, float64(stat.RequestCounter), filter, name, "total")
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["requests"], prometheus.CounterValue, float64(stat.Responses.OneXx), filter, name, "1xx")
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["requests"], prometheus.CounterValue, float64(stat.Responses.TwoXx), filter, name, "2xx")
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["requests"], prometheus.CounterValue, float64(stat.Responses.ThreeXx), filter, name, "3xx")
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["requests"], prometheus.CounterValue, float64(stat.Responses.FourXx), filter, name, "4xx")
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["requests"], prometheus.CounterValue, float64(stat.Responses.FiveXx), filter, name, "5xx")

			ch <- prometheus.MustNewConstMetric(e.filterMetrics["bytes"], prometheus.CounterValue, float64(stat.InBytes), filter, name, "in")
			ch <- prometheus.MustNewConstMetric(e.filterMetrics["bytes"], prometheus.CounterValue, float64(stat.OutBytes), filter, name, "out")
		}
	}

	// CacheZones
	for zone, s := range nginxVtx.CacheZones {
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Bypass), zone, "bypass")
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Expired), zone, "expired")
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Hit), zone, "hit")
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Miss), zone, "miss")
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Revalidated), zone, "revalidated")
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Scarce), zone, "scarce")
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Stale), zone, "stale")
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Updating), zone, "updating")

		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["bytes"], prometheus.CounterValue, float64(s.InBytes), zone, "in")
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["bytes"], prometheus.CounterValue, float64(s.OutBytes), zone, "out")
	}
}

func fetchHTTP(uri string, timeout time.Duration) func() (io.ReadCloser, error) {
	http.DefaultClient.Timeout = timeout
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: *insecure}

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
	showVersion        = flag.Bool("version", false, "Print version information.")
	listenAddress      = flag.String("telemetry.address", ":9913", "Address on which to expose metrics.")
	metricsEndpoint    = flag.String("telemetry.endpoint", "/metrics", "Path under which to expose metrics.")
	metricsNamespace   = flag.String("metrics.namespace", "nginx", "Prometheus metrics namespace.")
	nginxScrapeURI     = flag.String("nginx.scrape_uri", "http://localhost/status", "URI to nginx stub status page")
	insecure           = flag.Bool("insecure", true, "Ignore server certificate if using https")
	nginxScrapeTimeout = flag.Int("nginx.scrape_timeout", 2, "The number of seconds to wait for an HTTP response from the nginx.scrape_uri")
)

func init() {
	prometheus.MustRegister(version.NewCollector("nginx_vts_exporter"))
}

func main() {
	flag.Parse()

	if *showVersion {
		fmt.Fprintln(os.Stdout, version.Print("nginx_vts_exporter"))
		os.Exit(0)
	}

	log.Printf("Starting nginx_vts_exporter %s", version.Info())
	log.Printf("Build context %s", version.BuildContext())

	exporter := NewExporter(*nginxScrapeURI)
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
