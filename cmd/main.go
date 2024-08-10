package main

import (
	"fmt"
	"github.com/go-kit/log/level"
	"github.com/hq0101/clamd_exporter/pkg/exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/spf13/cobra"
	"net/http"
	"os"
	"time"
)

func main() {
	var listen, netType, address string
	var connTimeout, readTimeout time.Duration
	logger := promlog.New(&promlog.Config{})
	rootCmd := &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {
			reg := prometheus.NewRegistry()
			reg.MustRegister(exporter.New(netType, address, connTimeout, readTimeout, logger))
			http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{ErrorHandling: promhttp.ContinueOnError}))
			if err := http.ListenAndServe(listen, nil); err != nil {
				level.Error(logger).Log("msg", "failed to listen and serve", "err", err)
			}
		},
	}

	rootCmd.Flags().StringVarP(&listen, "listen", "l", ":8080", "listen address")
	rootCmd.Flags().StringVarP(&address, "address", "a", "", "ClamAV server address /var/run/clamav/clamd.ctl or 127.0.0.1:3310")
	rootCmd.Flags().StringVarP(&netType, "nettype", "n", "", "Network type (unix/tcp)")
	rootCmd.Flags().DurationVarP(&connTimeout, "conn_timeout", "t", 10*time.Second, "Connection timeout")
	rootCmd.Flags().DurationVarP(&readTimeout, "read_timeout", "r", 30*time.Second, "Read timeout")
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
