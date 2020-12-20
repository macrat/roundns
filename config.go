package main

import (
	"fmt"
	"net"
	"time"

	"github.com/go-yaml/yaml"
	"github.com/macrat/landns/lib-landns"
	"github.com/macrat/landns/lib-landns/logger"
	"github.com/miekg/dns"
)

type QType uint16

func (q *QType) UnmarshalText(text []byte) error {
	switch string(text) {
	case "A":
		*q = QType(dns.TypeA)
	case "AAAA":
		*q = QType(dns.TypeAAAA)
	case "CNAME":
		*q = QType(dns.TypeCNAME)
	case "TXT":
		*q = QType(dns.TypeTXT)
	}
	return nil
}

func (q QType) MarshalText() ([]byte, error) {
	x := landns.QtypeToString(uint16(q))
	return []byte(x), nil
}

type CommandConfig []string

func (c *CommandConfig) UnmarshalText(text []byte) error {
	err := yaml.Unmarshal(text, (*[]string)(c))
	if err == nil {
		return nil
	}

	var s string
	err = yaml.Unmarshal(text, &s)
	if err != nil {
		return err
	}
	*c = CommandConfig{"sh", "-c", s}

	return nil
}

type HealthConfig struct {
	Command  CommandConfig `yaml:"command"`
	Interval time.Duration `yaml:"interval"`
}

func (h *HealthConfig) UnmarshalText(text []byte) error {
	var x struct {
		Command  CommandConfig `yaml:"command"`
		Interval time.Duration `yaml:"interval"`
	}
	if err := yaml.Unmarshal(text, &x); err == nil {
		h.Command = x.Command
		h.Interval = x.Interval
		return nil
	}

	var cmd CommandConfig
	if err := yaml.Unmarshal(text, &cmd); err != nil {
		return err
	}

	h.Command = cmd
	h.Interval = 5 * time.Minute
	return nil
}

type TTL uint32

func (t *TTL) UnmarshalText(text []byte) error {
	if err := yaml.Unmarshal(text, (*uint32)(t)); err == nil {
		return nil
	}

	var d time.Duration
	if err := yaml.Unmarshal(text, &d); err != nil {
		return err
	}

	*t = TTL(d.Seconds())
	return nil
}

type StrategyConfig string

func (s *StrategyConfig) UnmarshalText(text []byte) error {
	if err := yaml.Unmarshal(text, (*string)(s)); err != nil {
		return err
	}

	if *s != "lound-robin" {
		return fmt.Errorf("unsupported strategy: %s", s)
	}

	return nil
}

func (s *StrategyConfig) MakeStrategy() Strategy {
	return &RoundRobinStrategy{}
}

type ServiceConfig struct {
	FQDN     landns.Domain  `yaml:"fqdn"`
	Type     QType          `yaml:"type"`
	TTL      TTL            `yaml:"ttl,omitempty"`
	Health   HealthConfig   `yaml:"health,omitempty"`
	Strategy StrategyConfig `yaml:"strategy,omitempty"`
	Hosts    []string       `yaml:"hosts"`
}

func (s ServiceConfig) host2record(host string) landns.Record {
	ttl := uint32(s.TTL)

	switch uint16(s.Type) {
	case dns.TypeA, dns.TypeAAAA:
		return landns.AddressRecord{Name: s.FQDN, TTL: ttl, Address: net.ParseIP(host)}
	case dns.TypeCNAME:
		return landns.CnameRecord{Name: s.FQDN, TTL: ttl, Target: landns.Domain(host)}
	case dns.TypeTXT:
		return landns.TxtRecord{Name: s.FQDN, TTL: ttl, Text: host}
	default:
		return nil
	}
}

func (s ServiceConfig) MakeResolver(log logger.Logger, metrics *landns.Metrics) ResolverWithHealthCheck {
	var hosts []Host
	for _, h := range s.Hosts {
		monitor := &HealthMonitor{
			Command:  s.Health.Command,
			Env:      map[string]string{"HOST": h},
			Interval: s.Health.Interval,
			Logger:   log,
		}
		hosts = append(hosts, Host{
			Record: s.host2record(h),
			Health: monitor,
		})
	}

	return LoadBalanceResolver{
		Qtype:    uint16(s.Type),
		Domain:   s.FQDN.String(),
		Hosts:    hosts,
		Strategy: s.Strategy.MakeStrategy(),
		Metrics:  metrics,
	}
}

type AddrConfig net.TCPAddr

func (a *AddrConfig) UnmarshalText(text []byte) error {
	addr, err := net.ResolveTCPAddr("tcp4", string(text))
	if err != nil {
		return err
	}
	*a = AddrConfig(*addr)
	return nil
}

func (a AddrConfig) TCPAddr() *net.TCPAddr {
	return (*net.TCPAddr)(&a)
}

func (a AddrConfig) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{IP: a.IP, Port: a.Port}
}

func (a AddrConfig) String() string {
	return (*net.TCPAddr)(&a).String()
}

type MetricsConfig struct {
	Address   *AddrConfig `yaml:"address"`
	Namespace string      `yaml:"namescape"`
}

type LogConfig struct {
	Level logger.Level `yaml:"level"`
}

type Config struct {
	Metrics  MetricsConfig   `yaml:"metrics"`
	Address  *AddrConfig     `yaml:"address"`
	Log      LogConfig       `yaml:"log"`
	TTL      TTL             `yaml:"ttl"`
	Services []ServiceConfig `yaml:"services"`
}

func (c Config) MakeResolver(log logger.Logger, metrics *landns.Metrics) ResolverWithHealthCheck {
	var rs ResolverWithHealthCheckSet

	ttl := c.TTL
	if ttl <= 0 {
		ttl = TTL(300)
	}

	for _, s := range c.Services {
		if s.TTL <= 0 {
			s.TTL = ttl
		}

		rs = append(rs, s.MakeResolver(log, metrics))
	}

	return rs
}
