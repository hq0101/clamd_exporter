package exporter

import (
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/hq0101/go-clamav/pkg/clamav"
	"github.com/prometheus/client_golang/prometheus"
	"regexp"
	"strconv"
	"time"
)

type Exporter struct {
	address          string
	logger           log.Logger
	clamdClient      *clamav.ClamClient
	up               *prometheus.Desc
	version          *prometheus.GaugeVec
	poolCount        *prometheus.Desc
	threadsLive      *prometheus.Desc
	threadsIdle      *prometheus.Desc
	threadsMax       *prometheus.Desc
	queue            *prometheus.Desc
	memoryHeap       *prometheus.Desc
	memoryUsed       *prometheus.Desc
	memoryMmap       *prometheus.Desc
	memoryFree       *prometheus.Desc
	memoryReleasable *prometheus.Desc
	memoryPoolsTotal *prometheus.Desc
	memoryPoolsUsed  *prometheus.Desc
}

func New(networkType, address string, connTimeout, readTimeout time.Duration, logger log.Logger) *Exporter {
	return &Exporter{
		address:     address,
		logger:      logger,
		clamdClient: clamav.NewClamClient(networkType, address, connTimeout, readTimeout),
		up: prometheus.NewDesc(
			"clamd_up",
			"Current health status of the server (1 = UP, 0 = DOWN).",
			nil,
			nil,
		),
		version: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "clamd_version",
				Help: "Clamd version information.",
			},
			[]string{"engine", "dbver", "dbtime"},
		),
		poolCount: prometheus.NewDesc(
			"clamd_pool_count",
			"Number of pool.",
			nil,
			nil,
		),
		threadsLive: prometheus.NewDesc(
			"clamd_threads_live",
			"Number of live threads.",
			nil,
			nil,
		),
		threadsIdle: prometheus.NewDesc(
			"clamd_threads_idle",
			"Number of idle threads.",
			nil,
			nil,
		),
		threadsMax: prometheus.NewDesc(
			"clamd_threads_max",
			"Maximum number of threads.",
			nil,
			nil,
		),
		queue: prometheus.NewDesc(
			"clamd_queue",
			"Number of queue.",
			nil,
			nil,
		),
		memoryHeap: prometheus.NewDesc(
			"clamd_memory_heap_bytes",
			"Memory heap.",
			nil,
			nil,
		),
		memoryUsed: prometheus.NewDesc(
			"clamd_memory_used_bytes",
			"Memory used.",
			nil,
			nil,
		),
		memoryMmap: prometheus.NewDesc(
			"clamd_memory_mmap_bytes",
			"Memory mmap.",
			nil,
			nil,
		),
		memoryFree: prometheus.NewDesc(
			"clamd_memory_free_bytes",
			"Memory free.",
			nil,
			nil,
		),
		memoryReleasable: prometheus.NewDesc(
			"clamd_memory_releasable_bytes",
			"Memory releasable.",
			nil,
			nil,
		),
		memoryPoolsTotal: prometheus.NewDesc(
			"clamd_memory_pools_total_bytes",
			"Memory pools total.",
			nil,
			nil,
		),
		memoryPoolsUsed: prometheus.NewDesc(
			"clamd_memory_pools_used_bytes",
			"Memory pools used.",
			nil,
			nil,
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- e.up
	e.version.Describe(ch)
	ch <- e.poolCount
	ch <- e.threadsLive
	ch <- e.threadsIdle
	ch <- e.threadsMax
	ch <- e.queue
	ch <- e.memoryHeap
	ch <- e.memoryUsed
	ch <- e.memoryMmap
	ch <- e.memoryFree
	ch <- e.memoryReleasable
	ch <- e.memoryPoolsTotal
	ch <- e.memoryPoolsUsed
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	result, err := e.clamdClient.Ping()
	if err != nil || result != "PONG" {
		ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 0)
		level.Error(e.logger).Log("msg", "Ping failed", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.up, prometheus.GaugeValue, 1)

	versionStr, err := e.clamdClient.Version()
	if err != nil {
		e.version.WithLabelValues("", "", "").Set(0)
		ch <- e.version.WithLabelValues("", "", "")
		level.Error(e.logger).Log("msg", "Failed to get ClamAV version", "err", err)
		return
	}
	versionInfo, err := parseVersion(versionStr)
	if err != nil {
		level.Error(e.logger).Log("msg", "Failed to parse ClamAV version", "err", err)
		e.version.WithLabelValues("", "", "").Set(1)
		ch <- e.version.WithLabelValues("", "", "")
		return
	}

	e.version.WithLabelValues(versionInfo.Engine, strconv.Itoa(versionInfo.DBVer), versionInfo.DBTime.Format("2006-01-02 15:04:05")).Set(1)
	ch <- e.version.WithLabelValues(versionInfo.Engine, strconv.Itoa(versionInfo.DBVer), versionInfo.DBTime.Format("2006-01-02 15:04:05"))

	stats, err := e.clamdClient.Stats()
	if err != nil {
		level.Error(e.logger).Log("msg", "Failed to get ClamAV stats", "err", err)
		return
	}
	ch <- prometheus.MustNewConstMetric(e.poolCount, prometheus.GaugeValue, float64(stats.Pools))
	ch <- prometheus.MustNewConstMetric(e.threadsLive, prometheus.GaugeValue, float64(stats.ThreadsLive))
	ch <- prometheus.MustNewConstMetric(e.threadsIdle, prometheus.GaugeValue, float64(stats.ThreadsIdle))
	ch <- prometheus.MustNewConstMetric(e.threadsMax, prometheus.GaugeValue, float64(stats.ThreadsMax))
	ch <- prometheus.MustNewConstMetric(e.queue, prometheus.GaugeValue, float64(stats.QueueItems))
	ch <- prometheus.MustNewConstMetric(e.memoryHeap, prometheus.GaugeValue, stats.MemHeap*1024*1024)
	ch <- prometheus.MustNewConstMetric(e.memoryUsed, prometheus.GaugeValue, stats.MemUsed*1024*1024)
	ch <- prometheus.MustNewConstMetric(e.memoryMmap, prometheus.GaugeValue, stats.MemMmap*1024*1024)
	ch <- prometheus.MustNewConstMetric(e.memoryFree, prometheus.GaugeValue, stats.MemFree*1024*1024)
	ch <- prometheus.MustNewConstMetric(e.memoryReleasable, prometheus.GaugeValue, stats.MemReleasable*1024*1024)
	ch <- prometheus.MustNewConstMetric(e.memoryPoolsTotal, prometheus.GaugeValue, stats.MemPoolsTotal*1024*1024)
	ch <- prometheus.MustNewConstMetric(e.memoryPoolsUsed, prometheus.GaugeValue, stats.MemPoolsUsed*1024*1024)
}

type VersionInfo struct {
	Engine string
	DBVer  int
	DBTime time.Time
}

func parseVersion(versionStr string) (*VersionInfo, error) {
	re := regexp.MustCompile(`ClamAV (\d+\.\d+\.\d+)/(\d+)/(.* \d+:\d+:\d+ \d+)`)
	matches := re.FindStringSubmatch(versionStr)

	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid version format: %s", versionStr)
	}

	engine := matches[1]
	dbver, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("failed to parse build number: %v", err)
	}

	layout := "Mon Jan _2 15:04:05 2006"
	date, err := time.Parse(layout, matches[3])
	if err != nil {
		return nil, fmt.Errorf("failed to parse date: %v", err)
	}

	return &VersionInfo{
		Engine: engine,
		DBVer:  dbver,
		DBTime: date,
	}, nil
}
