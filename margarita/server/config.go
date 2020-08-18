package server

import (
	"fmt"
	"github.com/countableset/poolside/margarita/api"
	"log"
	url2 "net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/countableset/poolside/margarita/config"

	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	auth "github.com/envoyproxy/go-control-plane/envoy/api/v2/auth"
	core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	listener "github.com/envoyproxy/go-control-plane/envoy/api/v2/listener"
	route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher"
	cache "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	v2cache "github.com/envoyproxy/go-control-plane/pkg/cache/v2"

	hcm "github.com/envoyproxy/go-control-plane/envoy/config/filter/network/http_connection_manager/v2"

	"github.com/envoyproxy/go-control-plane/pkg/resource/v2"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/golang/protobuf/ptypes"
)

var (
	version  int32
	tlsName  = "poolside.dev"
)

// ListenForConfigurationChanges listens and applies changes to cache
func ListenForConfigurationChanges(cache v2cache.SnapshotCache) {
	for configs := range api.ConfigUpdateChan {
		snap := CreateNewSnapShot(configs)
		cache.SetSnapshot("id_1", snap)
	}
}

// CreateNewSnapShot data used for xDS service
func CreateNewSnapShot(configs []api.Configuration) v2cache.Snapshot {
	size := len(configs)
	clusters := make([]cache.Resource, size)
	routes := make([]cache.Resource, size)
	listeners := make([]cache.Resource, size)

	for i, c := range configs {
		u, err := url2.Parse(c.Proxy)
		if err != nil {
			log.Printf("Invalid proxy url parsing %s %v", c.Proxy, err)
			continue
		}
		remoteHost := u.Hostname()
		port, err := strconv.ParseUint(u.Port(), 10, 32)
		if err != nil {
			log.Printf("Invaild port parsing from url %s %v", c.Proxy, err)
			continue
		}
		domain := c.Domain
		slug := Clean(c.Proxy)
		clusterName := "cluster_" + slug
		listenerName := "listener_" + slug
		routeName := "route_" + slug

		clusters[i] = makeCluster(clusterName, remoteHost, uint32(port))
		routes[i] = makeRoute(routeName, clusterName, domain)
		listeners[i] = makeListener(listenerName, routeName)
	}

	atomic.AddInt32(&version, 1)
	log.Printf(">>>>>>>>>>>>>>>>>>> creating snapshot Version %s", fmt.Sprint(version))
	out := v2cache.NewSnapshot(fmt.Sprint(version), nil, clusters, routes, listeners, nil)
	out.Resources[cache.Secret] = v2cache.NewResources(fmt.Sprint(version), makeSecret(tlsName))
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

func makeCluster(clusterName, hostname string, port uint32) *v2.Cluster {
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
	return &v2.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       ptypes.DurationProto(5 * time.Second),
		ClusterDiscoveryType: &v2.Cluster_Type{Type: v2.Cluster_LOGICAL_DNS},
		DnsLookupFamily:      v2.Cluster_V4_ONLY,
		LbPolicy:             v2.Cluster_ROUND_ROBIN,
		Hosts:                []*core.Address{h}, // TODO need to fix deprecation warning
	}
}

func makeRoute(routeName, clusterName, domain string) *v2.RouteConfiguration {
	log.Printf(">>>>>>>>>>>>>>>>>>> creating route %s %s %s", routeName, clusterName, domain)
	return &v2.RouteConfiguration{
		Name: routeName,
		VirtualHosts: []*route.VirtualHost{{
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
		}},
	}
}

func makeListener(listenerName string, route string) *v2.Listener {
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
	return &v2.Listener{
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

func makeSecret(tlsName string) []cache.Resource {
	log.Printf(">>>>>>>>>>>>>>>>>>> creating secret")
	return []cache.Resource{
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
