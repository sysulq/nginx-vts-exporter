package main

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
