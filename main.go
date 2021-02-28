package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/containers/podman/v2/cmd/podman/registry"
	"github.com/containers/podman/v2/libpod/define"
	"github.com/containers/podman/v2/pkg/domain/entities"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
)

const (
	namespace = "podman"
)

var (
	flagAddr        = flag.String("l", ":9901", "Address to listen on")
	statOpts        entities.ContainerStatsOptions
	collectorLabels = []string{"id", "name"}
)

type Collector struct {
	sync.Mutex
	numCtrs  prometheus.Gauge
	ctrName  *prometheus.GaugeVec
	pids     *prometheus.GaugeVec
	cpuperc  *prometheus.GaugeVec
	memusage *prometheus.GaugeVec
	memlimit *prometheus.GaugeVec
	memperc  *prometheus.GaugeVec
	netin    *prometheus.GaugeVec
	netout   *prometheus.GaugeVec
	blkin    *prometheus.GaugeVec
	blkout   *prometheus.GaugeVec
}

func NewCollector() *Collector {
	return &Collector{
		numCtrs: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "num_ctrs",
				Help:      "Number of running containers",
			},
		),
		pids: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "pids",
				Help:      "Number of running pids in the container",
			},
			collectorLabels,
		),
		cpuperc: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "cpuperc",
				Help:      "percentage cpu",
			},
			collectorLabels,
		),
		memusage: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memusage",
				Help:      "memory usage",
			},
			collectorLabels,
		),
		memlimit: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memlimit",
				Help:      "memory limit",
			},
			collectorLabels,
		),
		memperc: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "memperc",
				Help:      "memory percentage",
			},
			collectorLabels,
		),
		netin: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "netin",
				Help:      "network in",
			},
			collectorLabels,
		),
		netout: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "netout",
				Help:      "network out",
			},
			collectorLabels,
		),
		blkin: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "blkin",
				Help:      "blocks in",
			},
			collectorLabels,
		),
		blkout: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "blkout",
				Help:      "blocks out",
			},
			collectorLabels,
		),
	}
}

func init() {
	_ = registry.PodmanConfig()
	_, err := registry.NewContainerEngine(&cobra.Command{}, []string{})
	statOpts.Stream = false
	if err != nil {
		log.Fatal(err)
	}
}

func getAllStats() ([]define.ContainerStats, error) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	reports, err := registry.ContainerEngine().ContainerStats(
		registry.Context(),
		[]string{},
		statOpts,
	)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("deadline exceeded")
		case s := <-reports:
			return s.Stats, nil
		}
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.numCtrs.Desc()
	c.pids.Describe(ch)
	c.cpuperc.Describe(ch)
	c.memusage.Describe(ch)
	c.memlimit.Describe(ch)
	c.memperc.Describe(ch)
	c.netin.Describe(ch)
	c.netout.Describe(ch)
	c.blkin.Describe(ch)
	c.blkout.Describe(ch)
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.Lock()
	defer c.Unlock()

	c.pids.Reset()
	c.cpuperc.Reset()
	c.memusage.Reset()
	c.memlimit.Reset()
	c.memperc.Reset()
	c.netin.Reset()
	c.netout.Reset()
	c.blkin.Reset()
	c.blkout.Reset()

	reports, err := getAllStats()
	if err != nil {
		log.Printf("getAllStats() err: %v", err)
		return
	}

	c.numCtrs.Set(float64(len(reports)))

	for _, ctr := range reports {
		c.pids.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.PIDs))
		c.cpuperc.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.CPU))
		c.memusage.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.MemUsage))
		c.memlimit.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.MemLimit))
		c.memperc.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.MemPerc))
		c.netin.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.NetInput))
		c.netout.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.NetOutput))
		c.blkin.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.BlockInput))
		c.blkout.WithLabelValues(ctr.ContainerID, ctr.Name).Set(float64(ctr.BlockOutput))
	}
	ch <- c.numCtrs
	c.pids.Collect(ch)
	c.cpuperc.Collect(ch)
	c.memusage.Collect(ch)
	c.memlimit.Collect(ch)
	c.memperc.Collect(ch)
	c.netin.Collect(ch)
	c.netout.Collect(ch)
	c.blkin.Collect(ch)
	c.blkout.Collect(ch)
}

func main() {
	flag.Parse()
	prometheus.MustRegister(NewCollector())
	log.Fatalf("ListenAndServe error: %v", http.ListenAndServe(
		*flagAddr,
		promhttp.Handler(),
	))
}
