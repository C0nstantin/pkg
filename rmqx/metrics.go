package rmqx

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
	"strings"
)

type metrics struct {
	MsgsHandled  prometheus.Counter
	MsgsReceived prometheus.Counter
	MsgsRejected prometheus.Counter
}

var workerMetrics *metrics

func initMetrics() {
	namespace := "que_system"
	appName := "worker"
	if os.Getenv("WORKER") != "" {
		appName = strings.ReplaceAll(os.Getenv("WORKER"), "-", "_")
	}
	if os.Getenv("NAMESPACE") != "" {
		namespace = strings.ReplaceAll(os.Getenv("NAMESPACE"), "-", "_")
	}

	workerMetrics = &metrics{
		MsgsHandled: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "rmqx_worker",
			Name:      appName + "_rmq_messages_handled_total",
			Help:      "Number of done handled messages", // "Number of messages sent",
		}),
		MsgsReceived: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "rmqx_worker",
			Name:      appName + "_rmq_messages_received_total",
			Help:      "Number of messages received",
		}),
		MsgsRejected: promauto.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "rmqx_worker",
			Name:      appName + "_rmq_messages_rejected_total",
			Help:      "Number of messages rejected",
		}),
	}

	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
		w.WriteHeader(http.StatusOK)
	}))
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Printf("ERROR failed to start prometheus service:  %s", err)
	}
}
