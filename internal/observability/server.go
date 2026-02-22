package observability

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const shutdownTimeout = 5 * time.Second

type Server struct {
	addr     string
	httpSrv  *http.Server
	recorder Recorder
}

func NewServer(addr string) *Server {
	registry := prometheus.NewRegistry()
	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	recorder := NewPrometheusRecorder(registry)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	httpSrv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &Server{
		addr:     addr,
		httpSrv:  httpSrv,
		recorder: recorder,
	}
}

func (s *Server) Recorder() Recorder {
	return s.recorder
}

func (s *Server) Addr() string {
	return s.addr
}

func (s *Server) Run(ctx context.Context) error {
	shutdownDone := make(chan struct{})
	go func() {
		defer close(shutdownDone)
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		_ = s.httpSrv.Shutdown(shutdownCtx)
	}()

	err := s.httpSrv.ListenAndServe()
	<-shutdownDone
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("observability server: %w", err)
	}
	return nil
}
