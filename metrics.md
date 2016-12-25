# Metrics documentation
This file contain documentation about exposed Prometheus metrics

#Server main
## Metrics details
Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Connections**   | `{NAMESPACE}_server_connections`| status [active, reading, writing, waiting, accepted, handled]

##Metrics output example
```
# Server Connections
nginx_server_connections{status="accepted"} 70606
```

#Server zones
## Metrics details
Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Requests**      | `{NAMESPACE}server_requests`    | code [2xx, 3xx, 4xx, 5xx, total], host _(or domain name)_
 **Bytes**         | `{NAMESPACE}server_bytes`       | direction [in, out], host _(or domain name)_
 **Cache**         | `{NAMESPACE}server_cache`       | status [bypass, expired, hit, miss, revalidated, scarce, stale, updating], host _(or domain name)_

##Metrics output example
```
# Server Requests
nginx_server_requests{code="1xx",host="test.domain.com"} 0

# Server Bytes
nginx_server_bytes{direction="in",host="test.domain.com"} 21

# Server Cache
nginx_server_cache{host="test.domain.com",status="bypass"} 2
```

#Upstreams
## Metrics details
Nginx data         | Name                            | Exposed informations     
------------------ | ------------------------------- | ------------------------
 **Requests**      | `{NAMESPACE}_upstream_requests` | code [2xx, 3xx, 4xx, 5xx and total], upstream _(or upstream name)_
 **Bytes**         | `{NAMESPACE}_upstream_bytes`    | direction [in, out], upstream _(or upstream name)_
 **Response time** | `{NAMESPACE}_upstream_response` | backend (or server), in_bytes, out_bytes, upstream _(or upstream name)_
 ~~Requests/sec~~  | `NOT EXPORTED YET`              |

##Metrics output example
```
# Upstream Requests
nginx_upstream_requests{code="1xx",upstream="XXX-XXXXX-3000"} 0

# Upstream Bytes
nginx_upstream_bytes{direction="in",upstream="XXX-XXXXX-3000"} 0

# Upstream Response time
nginx_upstream_response{host="10.2.15.10:3000",in_bytes="285025.000000",out_bytes="447594.000000",upstream="XXX-XXXXX-3000"} 99
```
