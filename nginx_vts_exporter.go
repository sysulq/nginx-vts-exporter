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
	RequestMsec  int  `json:"requestMsec"`
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

type Exporter struct {
	URIs            []URI
	serverMetrics   map[string]*prometheus.Desc
	upstreamMetrics map[string]*prometheus.Desc
	cacheMetrics    map[string]*prometheus.Desc
}

type configType struct {
	TelemetryAddress  string `json:"telemetryAddress"`
	TelemetryEndpoint string `json:"telemetryEndpoint"`
	MetricsNamespace  string `json:"metricsNamespace"`
	NginxScrapeURIs   []URI  `json:"nginxScrapeURIs"`
}

type URI struct {
	HostName string `json:"hostName"`
	Uri      string `json:"uri"`
}

type cmd struct {
	showVersion      *bool
	listenAddress    *string
	metricsEndpoint  *string
	metricsNamespace *string
	nginxScrapeURI   *string
	configFile       *string
	insecure         *bool
}

func newServerMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(config.MetricsNamespace, "server", metricName),
		docString, labels, nil,
	)
}

func newUpstreamMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(config.MetricsNamespace, "upstream", metricName),
		docString, labels, nil,
	)
}

func newCacheMetric(metricName string, docString string, labels []string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(config.MetricsNamespace, "cache", metricName),
		docString, labels, nil,
	)
}

func NewExporter(uris []URI) *Exporter {

	return &Exporter{
		URIs: uris,
		serverMetrics: map[string]*prometheus.Desc{
			"connections": newServerMetric("connections", "nginx connections", []string{"status", "hostName"}),
			"requests":    newServerMetric("requests", "requests counter", []string{"host", "code", "hostName"}),
			"bytes":       newServerMetric("bytes", "request/response bytes", []string{"host", "direction", "hostName"}),
			"cache":       newServerMetric("cache", "cache counter", []string{"host", "status", "hostName"}),
		},
		upstreamMetrics: map[string]*prometheus.Desc{
			"requests": newUpstreamMetric("requests", "requests counter", []string{"upstream", "code", "hostName"}),
			"bytes":    newUpstreamMetric("bytes", "request/response bytes", []string{"upstream", "direction", "hostName"}),
			"response": newUpstreamMetric("response", "The average of only upstream response processing times in milliseconds", []string{"upstream", "backend", "hostName"}),
			"request":  newUpstreamMetric("request", "The average of request processing times including upstream in milliseconds.", []string{"upstream", "backend", "hostName"}),
		},
		cacheMetrics: map[string]*prometheus.Desc{
			"requests": newCacheMetric("requests", "cache requests counter", []string{"zone", "status", "hostName"}),
			"bytes":    newCacheMetric("bytes", "cache request/response bytes", []string{"zone", "direction", "hostName"}),
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
	for _, m := range e.cacheMetrics {
		ch <- m
	}
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {

	for _, uri := range e.URIs {

		body, err := fetchHTTP(uri.Uri, 2*time.Second)()
		if err != nil {
			log.Println("fetchHTTP failed", err)
			continue
		}

		data, err := ioutil.ReadAll(body)
		if err != nil {
			log.Println("ioutil.ReadAll failed", err)
			continue
		}

		var nginxVtx NginxVts
		err = json.Unmarshal(data, &nginxVtx)
		if err != nil {
			log.Printf("json.Unmarshal failed. URI=%s, ERROR=%s ", uri.Uri, err)
			continue
		}

		hostName := uri.HostName
		if len(hostName) == 0 {
			hostName = nginxVtx.HostName
		}

		Collect(nginxVtx, hostName, ch, e)

		//log.Printf("successfull scrape form hostName=%s", hostName)

		body.Close()
	}
}

func Collect(nginxVtx NginxVts, hostName string, ch chan<- prometheus.Metric, e *Exporter) {

	// connections
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Active), "active", hostName)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Reading), "reading", hostName)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Waiting), "waiting", hostName)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Writing), "writing", hostName)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Accepted), "accepted", hostName)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Handled), "handled", hostName)
	ch <- prometheus.MustNewConstMetric(e.serverMetrics["connections"], prometheus.GaugeValue, float64(nginxVtx.Connections.Requests), "requests", hostName)

	// ServerZones
	for host, s := range nginxVtx.ServerZones {
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.RequestCounter), host, "total", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.OneXx), host, "1xx", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.TwoXx), host, "2xx", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.ThreeXx), host, "3xx", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.FourXx), host, "4xx", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["requests"], prometheus.CounterValue, float64(s.Responses.FiveXx), host, "5xx", hostName)

		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Bypass), host, "bypass", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Expired), host, "expired", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Hit), host, "hit", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Miss), host, "miss", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Revalidated), host, "revalidated", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Scarce), host, "scarce", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Stale), host, "stale", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["cache"], prometheus.CounterValue, float64(s.Responses.Updating), host, "updating", hostName)

		ch <- prometheus.MustNewConstMetric(e.serverMetrics["bytes"], prometheus.CounterValue, float64(s.InBytes), host, "in", hostName)
		ch <- prometheus.MustNewConstMetric(e.serverMetrics["bytes"], prometheus.CounterValue, float64(s.OutBytes), host, "out", hostName)
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

			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["response"], prometheus.GaugeValue, float64(s.ResponseMsec), name, s.Server, hostName)
			ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["request"], prometheus.GaugeValue, float64(s.RequestMsec), name, s.Server, hostName)
		}

		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, total, name, "total", hostName)
		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, one, name, "1xx", hostName)
		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, two, name, "2xx", hostName)
		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, three, name, "3xx", hostName)
		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, four, name, "4xx", hostName)
		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["requests"], prometheus.CounterValue, five, name, "5xx", hostName)

		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["bytes"], prometheus.CounterValue, inbytes, name, "in", hostName)
		ch <- prometheus.MustNewConstMetric(e.upstreamMetrics["bytes"], prometheus.CounterValue, outbytes, name, "out", hostName)
	}

	// CacheZones
	for zone, s := range nginxVtx.CacheZones {
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Bypass), zone, "bypass", hostName)
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Expired), zone, "expired", hostName)
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Hit), zone, "hit", hostName)
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Miss), zone, "miss", hostName)
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Revalidated), zone, "revalidated", hostName)
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Scarce), zone, "scarce", hostName)
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Stale), zone, "stale", hostName)
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["requests"], prometheus.CounterValue, float64(s.Responses.Updating), zone, "updating", hostName)

		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["bytes"], prometheus.CounterValue, float64(s.InBytes), zone, "in", hostName)
		ch <- prometheus.MustNewConstMetric(e.cacheMetrics["bytes"], prometheus.CounterValue, float64(s.OutBytes), zone, "out", hostName)
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

func readConfig(config *configType, configFile *string) *string {

	log.Println("Read config file %s", *configFile)

	file, e := ioutil.ReadFile(*configFile)
	if e != nil {
		fmt.Printf("File error: %v\n", e)
	}

	err := json.Unmarshal(file, &config)
	if err != nil {
		log.Printf("Error parse config file %s. %s", configFile, err)
	}
	log.Printf("Read from config file. Results: %+v\n", config)

	if *cmd_args.listenAddress != listenAddress_def || config.TelemetryAddress == "" {
		config.TelemetryAddress = *cmd_args.listenAddress
	}
	if *cmd_args.metricsEndpoint != metricsEndpoint_def || config.TelemetryEndpoint == "" {
		config.TelemetryEndpoint = *cmd_args.metricsEndpoint
	}
	if *cmd_args.metricsNamespace != metricsNamespace_def || config.MetricsNamespace == "" {
		config.MetricsNamespace = *cmd_args.metricsNamespace
	}
	add_cmd_scripe := true
	for _, val := range config.NginxScrapeURIs {
		if val.Uri == *cmd_args.nginxScrapeURI {
			add_cmd_scripe = false
			break
		}
	}
	if *cmd_args.nginxScrapeURI == nginxScrapeURI_def && len(config.NginxScrapeURIs) != 0 {
		add_cmd_scripe = false
	}
	if add_cmd_scripe {
		config.NginxScrapeURIs = append(config.NginxScrapeURIs, URI{HostName: "", Uri: *cmd_args.nginxScrapeURI})
	}
	return nil
}

var (
	cmd_args cmd
	config   configType

	listenAddress_def    = ":9913"
	metricsEndpoint_def  = "/metrics"
	metricsNamespace_def = "nginx"
	nginxScrapeURI_def   = "http://localhost/status/format/json"
	configFile_def       = "/etc/nginx-vts-exporter/config.json"
)

func init() {
	prometheus.MustRegister(version.NewCollector("nginx_vts_exporter"))
}

func main() {

	cmd_args.showVersion = flag.Bool("version", false, "Print version information.")
	cmd_args.listenAddress = flag.String("telemetry.address", listenAddress_def, "Address on which to expose metrics.")
	cmd_args.metricsEndpoint = flag.String("telemetry.endpoint", metricsEndpoint_def, "Path under which to expose metrics.")
	cmd_args.metricsNamespace = flag.String("metrics.namespace", metricsNamespace_def, "Prometheus metrics namespace.")
	cmd_args.nginxScrapeURI = flag.String("nginx.scrape_uri", nginxScrapeURI_def, "URI to nginx VTS module json page")
	cmd_args.configFile = flag.String("config.file", configFile_def, "path to config.json file")
	cmd_args.insecure = flag.Bool("insecure", true, "Ignore server certificate if using https")

	flag.Parse()

	if *cmd_args.showVersion {
		fmt.Fprintln(os.Stdout, version.Print("nginx_vts_exporter"))
		os.Exit(0)
	}

	log.Printf("Starting nginx_vts_exporter %s", version.Info())
	log.Printf("Build context %s", version.BuildContext())

	readConfig(&config, cmd_args.configFile)

	log.Printf("Starting with config: %+v", config)

	exporter := NewExporter(config.NginxScrapeURIs)
	prometheus.MustRegister(exporter)
	prometheus.Unregister(prometheus.NewProcessCollector(os.Getpid(), ""))
	prometheus.Unregister(prometheus.NewGoCollector())

	http.Handle(config.TelemetryEndpoint, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Nginx Exporter</title></head>
			<body>
			<h1>Nginx Exporter</h1>
			<p><a href="` + config.TelemetryEndpoint + `">Metrics</a></p>
			</body>
			</html>`))
	})

	log.Printf("Starting Server at : %s", config.TelemetryAddress)
	log.Printf("Metrics endpoint: %s", config.TelemetryEndpoint)
	log.Printf("Metrics namespace: %s", config.MetricsNamespace)
	log.Printf("Scraping information from : %+v", config.NginxScrapeURIs)
	log.Fatal(http.ListenAndServe(config.TelemetryAddress, nil))
}
