package main

import (
	"context"
	"fmt"

	"github.com/macrat/landns/lib-landns"
)

type ResolverWithHealthCheck interface {
	landns.Resolver

	RunHealthCheck(context.Context)
}

type ResolverWithHealthCheckSet []ResolverWithHealthCheck

func (rs ResolverWithHealthCheckSet) RunHealthCheck(ctx context.Context) {
	for _, r := range rs {
		r.RunHealthCheck(ctx)
	}
}

func (rs ResolverWithHealthCheckSet) Resolve(w landns.ResponseWriter, r landns.Request) error {
	for _, x := range rs {
		if err := x.Resolve(w, r); err != nil {
			return err
		}
	}
	return nil
}

func (rs ResolverWithHealthCheckSet) RecursionAvailable() bool {
	res := false
	for _, r := range rs {
		if r.RecursionAvailable() {
			res = true
		}
	}
	return res
}

func (rs ResolverWithHealthCheckSet) Close() error {
	for _, r := range rs {
		if err := r.Close(); err != nil {
			return err
		}
	}
	return nil
}

type LoadBalanceResolver struct {
	Qtype    uint16
	Domain   string
	Hosts    []Host
	Strategy Strategy
	Metrics  *landns.Metrics
}

func (lb LoadBalanceResolver) RunHealthCheck(ctx context.Context) {
	for _, h := range lb.Hosts {
		if h.Health != nil {
			go h.Health.Run(ctx)
		}
	}
}

func (lb LoadBalanceResolver) getStatus() (alives []bool, selectable bool) {
	alives = make([]bool, len(lb.Hosts))
	for i, h := range lb.Hosts {
		alives[i] = h.Health.IsHealthy()
		if alives[i] {
			selectable = true
		}
	}
	return
}

func (lb LoadBalanceResolver) Resolve(w landns.ResponseWriter, r landns.Request) error {
	if r.Qtype == lb.Qtype && r.Name == lb.Domain {
		alives, ok := lb.getStatus()
		if !ok {
			return fmt.Errorf("all hosts is down")
		}

		idx, err := lb.Strategy.Select(alives)
		if err != nil {
			return err
		}

		w.Add(lb.Hosts[idx].Record)
	}
	return nil
}

func (lb LoadBalanceResolver) RecursionAvailable() bool {
	return false
}

func (lb LoadBalanceResolver) Close() error {
	return nil
}
