package uos

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMonitoring() {
	if config.Monitoring.PortPPROF > 0 {
		go func() {
			LogInfo("starting PPROF web interface")
			pprofMux := http.NewServeMux()

			pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
			pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
			pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
			pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
			pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

			err := http.ListenAndServe(fmt.Sprintf(":%d", config.Monitoring.PortPPROF), pprofMux)
			if err != nil {
				LogErrorObj("profiling web interface stopped", err)
			}
		}()
	}

	if config.Monitoring.PortMetrics > 0 {
		go func() {
			LogInfo("starting metrics server")
			metricsMux := http.NewServeMux()

			metricsMux.Handle("/metrics", promhttp.Handler())

			err := http.ListenAndServe(fmt.Sprintf(":%d", config.Monitoring.PortMetrics), metricsMux)
			if err != nil {
				LogErrorObj("profiling web interface stopped", err)
			}
		}()
	}
}
