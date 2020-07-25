package server

import (
	"fmt"
	"log"
	"sync/atomic"
	"time"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	v2cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"

	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
)

var (
	localhost = "127.0.0.1"
	version   int32
)

func DemoData() v2cache.Snapshot {

	var remoteHost = "localhost"
	var virtualHostName = "local_service"
	var clusterName = "demo"
	var listenerName = "demo_listener"

	log.Printf(">>>>>>>>>>>>>>>>>>> creating cluster %v  with  %s", clusterName, remoteHost)

	// address config
	h := &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  remoteHost,
			Protocol: core.SocketAddress_TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: uint32(8000),
			},
		},
	}}

	// cluster
	c := []cache.Resource{
		&v2.Cluster{
			Name:                 clusterName,
			ConnectTimeout:       ptypes.DurationProto(2 * time.Second),
			ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_LOGICAL_DNS},
			DnsLookupFamily:      v2.Cluster_V4_ONLY,
			LbPolicy:             v2.Cluster_ROUND_ROBIN,
			Hosts:                []*core.Address{h},
		},
	}

	log.Printf(">>>>>>>>>>>>>>>>>>> creating listener %s", listenerName)

	var targetRegex = ".*"
	var routeConfigName = "local_route"

	v := route.VirtualHost{
		Name:    virtualHostName,
		Domains: []string{"test.local.bimmer-tech.com"},

		Routes: []*route.Route{{
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_SafeRegex{
					SafeRegex: &matcher.RegexMatcher{EngineType: &matcher.RegexMatcher_GoogleRe2{}, Regex: targetRegex},
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					HostRewriteSpecifier: &route.RouteAction_HostRewrite{
						HostRewrite: remoteHost,
					},
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: clusterName,
					},
				},
			},
		}}}

	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_RouteConfig{
			RouteConfig: &v2.RouteConfiguration{
				Name:         routeConfigName,
				VirtualHosts: []*route.VirtualHost{&v},
			},
		},
		HttpFilters: []*hcm.HttpFilter{{
			Name: wellknown.Router,
		}},
	}

	pbst, err := ptypes.MarshalAny(manager)
	if err != nil {
		panic(err)
	}

	var l = []cache.Resource{
		&v2.Listener{
			Name: listenerName,
			Address: &core.Address{
				Address: &core.Address_SocketAddress{
					SocketAddress: &core.SocketAddress{
						Protocol: core.SocketAddress_TCP,
						Address:  localhost,
						PortSpecifier: &core.SocketAddress_PortValue{
							PortValue: 10000,
						},
					},
				},
			},
			FilterChains: []*listener.FilterChain{{
				Filters: []*listener.Filter{{
					Name: wellknown.HTTPConnectionManager,
					ConfigType: &listener.Filter_TypedConfig{
						TypedConfig: pbst,
					},
				}},
			}},
		}}

	atomic.AddInt32(&version, 1)
	log.Printf(">>>>>>>>>>>>>>>>>>> creating snapshot Version %s", fmt.Sprint(version))
	return v2cache.NewSnapshot(fmt.Sprint(version), nil, c, nil, l, nil)
}
