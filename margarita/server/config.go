package server

import (
	"fmt"
	"log"
	url2 "net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/countableset/poolside/margarita/api"
	"github.com/countableset/poolside/margarita/config"

	"github.com/golang/protobuf/ptypes"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	auth "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	"github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	"github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
)

var (
	version int32
	tlsName = "poolside.dev"
)

// ListenForConfigurationChanges listens and applies changes to cache
func ListenForConfigurationChanges(cache cache.SnapshotCache) {
	for configs := range api.ConfigUpdateChan {
		configs = append(configs, api.Configuration{
			Domain: config.GetMargaritaDomain(),
			Proxy:  fmt.Sprintf("http://%s:%d", config.GetMargaritaHost(), config.GetMargaritaPort()),
		})
		snap := CreateNewSnapShot(configs)
		cache.SetSnapshot("id_1", snap)
	}
}

// CreateNewSnapShot data used for xDS service
func CreateNewSnapShot(configs []api.Configuration) cache.Snapshot {
	size := len(configs)
	clusters := make([]types.Resource, size)
	routes := make([]*route.VirtualHost, size)
	route := make([]types.Resource, 1)
	listeners := make([]types.Resource, 1)

	for i, c := range configs {
		if !strings.Contains(c.Proxy, "://") {
			c.Proxy = fmt.Sprintf("http://%s", c.Proxy)
		}
		u, err := url2.Parse(c.Proxy)
		if err != nil {
			log.Printf("Invalid proxy url parsing %s %v", c.Proxy, err)
			continue
		}
		remoteHost := u.Hostname()
		port, err := strconv.ParseUint(u.Port(), 10, 32)
		if err != nil {
			log.Printf("Invalid port parsing from url %s %v", c.Proxy, err)
			continue
		}
		domain := c.Domain
		slug := Clean(c.Proxy)
		clusterName := "cluster_" + slug
		routeName := "route_" + slug

		clusters[i] = makeCluster(clusterName, remoteHost, uint32(port))
		routes[i] = makeVHostRoute(routeName, clusterName, domain)
	}
	listenerName := "listener_https"
	routeName := "route_rds"
	route[0] = makeRoute(routeName, routes)
	listeners[0] = makeListener(listenerName, routeName)
	secret := makeSecret(tlsName)

	atomic.AddInt32(&version, 1)
	log.Printf(">>>>>>>>>>>>>>>>>>> creating snapshot Version %s", fmt.Sprint(version))
	out := cache.NewSnapshot(fmt.Sprint(version), []types.Resource{}, clusters, route, listeners, []types.Resource{}, secret)
	return out
}

func xdsSource() *core.ConfigSource {
	source := &core.ConfigSource{}
	source.ResourceApiVersion = resource.DefaultAPIVersion
	source.ConfigSourceSpecifier = &core.ConfigSource_ApiConfigSource{
		ApiConfigSource: &core.ApiConfigSource{
			TransportApiVersion:       resource.DefaultAPIVersion,
			ApiType:                   core.ApiConfigSource_GRPC,
			SetNodeOnFirstMessageOnly: true,
			GrpcServices: []*core.GrpcService{{
				TargetSpecifier: &core.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &core.GrpcService_EnvoyGrpc{ClusterName: config.GetXdsCluster()},
				},
			}},
		},
	}
	return source
}

func makeCluster(clusterName, hostname string, port uint32) *cluster.Cluster {
	log.Printf(">>>>>>>>>>>>>>>>>>> creating cluster %v with %s and %d", clusterName, hostname, port)
	// address config
	h := &core.Address{Address: &core.Address_SocketAddress{
		SocketAddress: &core.SocketAddress{
			Address:  hostname,
			Protocol: core.SocketAddress_TCP,
			PortSpecifier: &core.SocketAddress_PortValue{
				PortValue: port,
			},
		},
	}}
	// cluster
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_LOGICAL_DNS},
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       makeEndpoint(clusterName, h),
	}
}

func makeEndpoint(clusterName string, addr *core.Address) *endpoint.ClusterLoadAssignment {
	return &endpoint.ClusterLoadAssignment{
		ClusterName: clusterName,
		Endpoints: []*endpoint.LocalityLbEndpoints{{
			LbEndpoints: []*endpoint.LbEndpoint{{
				HostIdentifier: &endpoint.LbEndpoint_Endpoint{
					Endpoint: &endpoint.Endpoint{
						Address: addr,
					},
				},
			}},
		}},
	}
}

func makeVHostRoute(routeName, clusterName, domain string) *route.VirtualHost {
	log.Printf(">>>>>>>>>>>>>>>>>>> creating vhost route %s %s %s", routeName, clusterName, domain)
	return &route.VirtualHost{
		Name:    routeName,
		Domains: []string{domain},
		Routes: []*route.Route{{
			Match: &route.RouteMatch{
				PathSpecifier: &route.RouteMatch_SafeRegex{
					SafeRegex: &matcher.RegexMatcher{EngineType: &matcher.RegexMatcher_GoogleRe2{}, Regex: ".*"},
				},
			},
			Action: &route.Route_Route{
				Route: &route.RouteAction{
					ClusterSpecifier: &route.RouteAction_Cluster{
						Cluster: clusterName,
					},
				},
			},
		}},
	}
}

func makeRoute(routeName string, vhosts []*route.VirtualHost) *route.RouteConfiguration {
	log.Printf(">>>>>>>>>>>>>>>>>>> creating route %s", routeName)
	return &route.RouteConfiguration{
		Name:         routeName,
		VirtualHosts: vhosts,
	}
}

func makeListener(listenerName string, route string) *listener.Listener {
	log.Printf(">>>>>>>>>>>>>>>>>>> creating listener %s %s", listenerName, route)
	rdsSource := xdsSource()
	// HTTP filter configuration
	manager := &hcm.HttpConnectionManager{
		CodecType:  hcm.HttpConnectionManager_AUTO,
		StatPrefix: "ingress_http",
		RouteSpecifier: &hcm.HttpConnectionManager_Rds{
			Rds: &hcm.Rds{
				ConfigSource:    rdsSource,
				RouteConfigName: route,
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
	// tls
	tlsc := &auth.DownstreamTlsContext{
		CommonTlsContext: &auth.CommonTlsContext{
			TlsCertificateSdsSecretConfigs: []*auth.SdsSecretConfig{{
				Name:      tlsName,
				SdsConfig: xdsSource(),
			}},
			ValidationContextType: &auth.CommonTlsContext_ValidationContextSdsSecretConfig{
				ValidationContextSdsSecretConfig: &auth.SdsSecretConfig{
					Name:      tlsName,
					SdsConfig: xdsSource(),
				},
			},
		},
	}
	mt, _ := ptypes.MarshalAny(tlsc)
	// listener
	return &listener.Listener{
		Name: listenerName,
		Address: &core.Address{
			Address: &core.Address_SocketAddress{
				SocketAddress: &core.SocketAddress{
					Protocol: core.SocketAddress_TCP,
					Address:  config.GetEnvoyHost(),
					PortSpecifier: &core.SocketAddress_PortValue{
						PortValue: config.GetEnvoyPort(),
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
			TransportSocket: &core.TransportSocket{
				Name: wellknown.TransportSocketTls,
				ConfigType: &core.TransportSocket_TypedConfig{
					TypedConfig: mt,
				},
			},
		}},
	}
}

func makeSecret(tlsName string) []types.Resource {
	log.Printf(">>>>>>>>>>>>>>>>>>> creating secret")
	return []types.Resource{
		&auth.Secret{
			Name: tlsName,
			Type: &auth.Secret_TlsCertificate{
				TlsCertificate: &auth.TlsCertificate{
					PrivateKey: &core.DataSource{
						Specifier: &core.DataSource_Filename{Filename: "/etc/envoy/certs/key.pem"},
					},
					CertificateChain: &core.DataSource{
						Specifier: &core.DataSource_Filename{Filename: "/etc/envoy/certs/cert.pem"},
					},
				},
			},
		},
	}
}
