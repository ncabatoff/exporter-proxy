package main

import (
	"flag"
	"io"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"strings"
)

func main() {
	var (
		listenAddress = flag.String("web.listen-address", "",
			"Address on which to expose metrics and web interface.")
		metricsPath = flag.String("web.telemetry-path", "/metrics",
			"Path under which to expose metrics.")
		localport = flag.Int("localport", 0,
			"port bound to localhost from which to proxy metrics")
	)
	flag.Parse()

	if *localport == 0 {
		log.Fatal("Must specify a local port to proxy")
	}

	if !strings.Contains(*listenAddress, ":") {
		*listenAddress += ":" + strconv.Itoa(*localport)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
			<head><title>Exporter Proxy</title></head>
			<body>
			<h1>Exporter Proxy</h1>
			<p><a href="` + *metricsPath + `">Metrics</a></p>
			</body>
			</html>`))
	})

	url := "http://localhost:" + strconv.Itoa(*localport) + *metricsPath
	client := &http.Client{}

	http.HandleFunc(*metricsPath, func(w http.ResponseWriter, r *http.Request) {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("error creating request: %v", err)
		}
		req.Header.Set("Accept-Encoding", r.Header.Get("Accept-Encoding"))
		req.Header.Set("Accept", r.Header.Get("Accept"))
		resp, err := client.Do(req)

		code := 500
		if err != nil {
			log.Printf("proxied exporter returned an error: %v", err)
		} else {
			code = resp.StatusCode
		}

		for k, vals := range resp.Header {
			for _, val := range vals {
				w.Header().Add(k, val)
			}
		}
		w.WriteHeader(code)

		_, err = io.Copy(w, resp.Body)
		if err != nil {
			log.Printf("error writing response: %v", err)
		}

		resp.Body.Close()
	})

	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("Unable to setup HTTP server: %v", err)
	}
}
