package main

import (
	"net"
	"testing"

	"github.com/macrat/landns/lib-landns"
	"github.com/macrat/landns/lib-landns/testutil"
	"github.com/miekg/dns"
)

func TestLoadBalanceResolver(t *testing.T) {
	rs := LoadBalanceResolver{
		Qtype:  dns.TypeA,
		Domain: "test.local.",
		Hosts: []Host{
			{Record: landns.AddressRecord{Name: "test.local.", Address: net.ParseIP("127.0.1.2"), TTL: 300}},
			{Record: landns.AddressRecord{Name: "test.local.", Address: net.ParseIP("127.0.1.3"), TTL: 300}},
		},
		Strategy: &RoundRobinStrategy{},
	}

	tests := []struct {
		res string
		req landns.Request
	}{
		{"test.local. 300 IN A 127.0.1.3", landns.NewRequest("test.local.", dns.TypeA, false)},
		{"test.local. 300 IN A 127.0.1.2", landns.NewRequest("test.local.", dns.TypeA, false)},
		{"test.local. 300 IN A 127.0.1.3", landns.NewRequest("test.local.", dns.TypeA, false)},
		{"test.local. 300 IN A 127.0.1.2", landns.NewRequest("test.local.", dns.TypeA, false)},
		{"", landns.NewRequest("hoge.local.", dns.TypeA, false)},
		{"", landns.NewRequest("test.local.", dns.TypeCNAME, false)},
	}

	for _, tt := range tests {
		res := testutil.NewDummyResponseWriter()

		if err := rs.Resolve(res, tt.req); err != nil {
			t.Errorf("failed to resolve: %s: %s", tt.req, err)
			continue
		}

		if len(tt.res) == 0 {
			if len(res.Records) > 0 {
				t.Errorf("unexpected response count: expected 0 but got %d", len(res.Records))
			}
			continue
		}

		if len(res.Records) != 1 {
			t.Errorf("unexpected response count: expected 1 but got %d", len(res.Records))
		} else if res.Records[0].String() != tt.res {
			t.Errorf("unexpected response:\nexpected: %s\nactual:   %s", res.Records[0], tt.res)
		}
	}
}

func TestLoadBalanceResolver_allDown(t *testing.T) {
	rs := LoadBalanceResolver{
		Qtype:  dns.TypeA,
		Domain: "test.local.",
		Hosts: []Host{
			{Record: landns.AddressRecord{Name: "test.local.", Address: net.ParseIP("127.0.1.2"), TTL: 300}, Health: &HealthMonitor{}},
			{Record: landns.AddressRecord{Name: "test.local.", Address: net.ParseIP("127.0.1.3"), TTL: 300}, Health: &HealthMonitor{}},
		},
		Strategy: &RoundRobinStrategy{},
	}

	req := landns.NewRequest("test.local.", dns.TypeA, false)
	res := testutil.NewDummyResponseWriter()

	if err := rs.Resolve(res, req); err == nil {
		t.Errorf("expected cause error but not cauesd")
	} else if err.Error() != "all hosts is down" {
		t.Errorf("unexpected error: %s", err)
	}
}
