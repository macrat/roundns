package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/macrat/landns/lib-landns"
	"github.com/macrat/landns/lib-landns/logger"
	"gopkg.in/yaml.v2"
)

var (
	configPath = flag.String("config", "./config.yml", "Path to config file.")
)

func main() {
	flag.Parse()

	f, err := os.Open(*configPath)
	if err != nil {
		fmt.Println(err)
		return
	}
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Println(err)
		return
	}
	var conf Config
	if err = yaml.Unmarshal(raw, &conf); err != nil {
		fmt.Println(err)
		return
	}

	if conf.Address == nil {
		conf.Address = &AddrConfig{IP: net.ParseIP("0.0.0.0"), Port: 53}
	}
	if conf.Metrics.Address == nil {
		conf.Metrics.Address = &AddrConfig{IP: net.ParseIP("0.0.0.0"), Port: 8553}
	}
	if len(conf.Metrics.Namespace) == 0 {
		conf.Metrics.Namespace = "roundns"
	}
	if conf.Log.Level == 0 {
		conf.Log.Level = logger.WarnLevel
	}

	log := logger.New(os.Stdout, logger.DebugLevel)
	metrics := landns.NewMetrics(conf.Metrics.Namespace)

	ctx := context.Background()

	rs := conf.MakeResolver(log)
	rs.RunHealthCheck(ctx)

	server := landns.Server{
		Name:      conf.Metrics.Namespace,
		Metrics:   metrics,
		Resolvers: rs,
	}

	log.Info("start server", logger.Fields{
		"metrics_address":   conf.Metrics.Address.String(),
		"metrics_namespace": conf.Metrics.Namespace,
		"address":           conf.Address.String(),
		"log_level":         conf.Log.Level,
	})

	log.Fatal(server.ListenAndServe(
		ctx,
		conf.Metrics.Address.TCPAddr(),
		conf.Address.UDPAddr(),
		"udp",
	).Error(), nil)
}
