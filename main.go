package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"html"
	"log"
	"net/http"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

var totalRunRequests = prometheus.NewCounter(
	prometheus.CounterOpts{
		Name:      "http_requests_total",
		Help:      "The total number of run HTTP requests.",
		Namespace: "load_test",
		Subsystem: "stress_tester_go",
	},
)

var runDuration = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Name:      "run_duration_seconds",
		Help:      "The duration of the run.",
		Namespace: "load_test",
		Subsystem: "stress_tester_go",
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
	// 500_000 iterations takes around 500ms
	for i := 0; i < iterations; i++ {
		var n uint64
		var e, term float64
		for j := int(0); j < 65; j++ {
			term = float64(1) / float64(factorial(n))
			e += term
			n++
		}
	}
}

func keygen(iterations int) error {
	for i := 0; i < iterations; i++ {
		privateKey, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return err
		}

		pub, err := ssh.NewPublicKey(&privateKey.PublicKey)
		if err != nil {
			return err
		}

		_ = ssh.FingerprintSHA256(pub)
	}
	return nil
}

func run(w http.ResponseWriter, req *http.Request) {
	totalRunRequests.Inc()
	start := time.Now()
	query := req.URL.Query()

	iterations := 1
	if query.Has("iterations") {
		i, err := strconv.Atoi(query.Get("iterations"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "Error: invalid value for iterations: %v", err)
		}
		iterations = i
	}

	switch function := query.Get("function"); function {
	case "euler":
		euler(iterations)
	case "keygen":
		err := keygen(iterations)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = fmt.Fprintf(w, "Error: error in key generation: %v", err)
		}
	case "":
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Error: missing function parameter in query string")
	default:
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintf(w, "Error: unknown function '%s'\n", html.EscapeString(function))
	}

	duration := time.Since(start).Seconds()
	_, _ = fmt.Fprintf(w, "\nElapsed Time: %f", duration)
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
