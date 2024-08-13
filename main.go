package main

import (
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"html"
	"log"
	"net/http"
	"time"
)

var totalRunRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name: "http_requests_total",
		Help: "The total number of run HTTP requests.",
	},
)

var runDuration = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name: "run_duration_seconds",
		Help: "The duration of the run.",
	},
)

func factorial(n uint64) uint64 {
	if n < 2 {
		return 1
	}
	fact := uint64(1)
	for i := uint64(2); i <= n; i++ {
		fact *= i
	}
	return fact
}

func euler(iterations int) {

	for i := 0; i < iterations; i++ {
		var n uint64
		var e, term float64
		for i := int(0); i < 65; i++ {
			term = float64(1) / float64(factorial(n))
			e += term
			n++
		}
	}
}

func run(w http.ResponseWriter, req *http.Request) {
	totalRunRequests.Inc()
	start := time.Now()
	query := req.URL.Query()
	switch function := query.Get("function"); function {
	case "euler":
		euler(500_000)
	case "":
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Error: missing function parameter in query string")
	default:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Error: unknown function '%s'\n", html.EscapeString(function))
	}
	duration := time.Since(start).Seconds()
	_, _ = fmt.Fprintf(w, "%f", duration)
	runDuration.Set(duration)

}

func main() {
	prometheus.MustRegister(totalRunRequests)
	prometheus.MustRegister(runDuration)
	prometheus.MustRegister(collectors.NewBuildInfoCollector())

	http.HandleFunc("/run", run)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8090", nil))
}
