package network

import (
	"encoding/json"
	"fmt"
	"net/http"

	"golang.org/x/net/context"

	"github.com/docker/docker/api/server/httputils"
	"github.com/docker/docker/errors"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/docker/engine-api/types/network"
	"github.com/docker/libnetwork"
)

func (n *networkRouter) getNetworksList(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	filter := r.Form.Get("filters")
	netFilters, err := filters.FromParam(filter)
	if err != nil {
		return err
	}

	list := []types.NetworkResource{}

	if nr, err := n.clusterProvider.GetNetworks(); err == nil {
		for _, nw := range nr {
			list = append(list, nw)
		}
	}

	// Combine the network list returned by Docker daemon if it is not already
	// returned by the cluster manager
SKIP:
	for _, nw := range n.backend.GetNetworks() {
		for _, nl := range list {
			if nl.ID == nw.ID() {
				continue SKIP
			}
		}
		list = append(list, *n.buildNetworkResource(nw))
	}

	list, err = filterNetworks(list, netFilters)
	if err != nil {
		return err
	}
	return httputils.WriteJSON(w, http.StatusOK, list)
}

func (n *networkRouter) getNetwork(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	nw, err := n.backend.FindNetwork(vars["id"])
	if err != nil {
		if nr, err := n.clusterProvider.GetNetwork(vars["id"]); err == nil {
			return httputils.WriteJSON(w, http.StatusOK, nr)
		}
		return err
	}
	return httputils.WriteJSON(w, http.StatusOK, n.buildNetworkResource(nw))
}

func (n *networkRouter) postNetworkCreate(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	var create types.NetworkCreateRequest

	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	if err := json.NewDecoder(r.Body).Decode(&create); err != nil {
		return err
	}

	if _, err := n.clusterProvider.GetNetwork(create.Name); err == nil {
		return libnetwork.NetworkNameError(create.Name)
	}

	nw, err := n.backend.CreateNetwork(create)
	if err != nil {
		if _, ok := err.(libnetwork.ManagerRedirectError); !ok {
			return err
		}
		id, err := n.clusterProvider.CreateNetwork(create)
		if err != nil {
			return err
		}
		nw = &types.NetworkCreateResponse{ID: id}
	}

	return httputils.WriteJSON(w, http.StatusCreated, nw)
}

func (n *networkRouter) postNetworkConnect(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	var connect types.NetworkConnect
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	if err := json.NewDecoder(r.Body).Decode(&connect); err != nil {
		return err
	}

	nw, err := n.backend.FindNetwork(vars["id"])
	if err != nil {
		return err
	}

	if nw.Info().Dynamic() {
		err := fmt.Errorf("operation not supported for swarm scoped networks")
		return errors.NewRequestForbiddenError(err)
	}

	return n.backend.ConnectContainerToNetwork(connect.Container, nw.Name(), connect.EndpointConfig)
}

func (n *networkRouter) postNetworkDisconnect(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	var disconnect types.NetworkDisconnect
	if err := httputils.ParseForm(r); err != nil {
		return err
	}

	if err := httputils.CheckForJSON(r); err != nil {
		return err
	}

	if err := json.NewDecoder(r.Body).Decode(&disconnect); err != nil {
		return err
	}

	nw, err := n.backend.FindNetwork(vars["id"])
	if err != nil {
		return err
	}

	if nw.Info().Dynamic() {
		err := fmt.Errorf("operation not supported for swarm scoped networks")
		return errors.NewRequestForbiddenError(err)
	}

	return n.backend.DisconnectContainerFromNetwork(disconnect.Container, nw, disconnect.Force)
}

func (n *networkRouter) deleteNetwork(ctx context.Context, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := httputils.ParseForm(r); err != nil {
		return err
	}
	if _, err := n.clusterProvider.GetNetwork(vars["id"]); err == nil {
		return n.clusterProvider.RemoveNetwork(vars["id"])
	}
	if err := n.backend.DeleteNetwork(vars["id"]); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (n *networkRouter) buildNetworkResource(nw libnetwork.Network) *types.NetworkResource {
	r := &types.NetworkResource{}
	if nw == nil {
		return r
	}

	info := nw.Info()
	r.Name = nw.Name()
	r.ID = nw.ID()
	r.Scope = info.Scope()
	if n.clusterProvider.IsManager() {
		if _, err := n.clusterProvider.GetNetwork(nw.Name()); err == nil {
			r.Scope = "swarm"
		}
	} else if info.Dynamic() {
		r.Scope = "swarm"
	}
	r.Driver = nw.Type()
	r.EnableIPv6 = info.IPv6Enabled()
	r.Internal = info.Internal()
	r.Options = info.DriverOptions()
	r.Containers = make(map[string]types.EndpointResource)
	buildIpamResources(r, info)
	r.Internal = info.Internal()
	r.Labels = info.Labels()

	epl := nw.Endpoints()
	for _, e := range epl {
		ei := e.Info()
		if ei == nil {
			continue
		}
		sb := ei.Sandbox()
		tmpID := e.ID()
		key := "ep-" + tmpID
		if sb != nil {
			key = sb.ContainerID()
		}

		r.Containers[key] = buildEndpointResource(tmpID, e.Name(), ei)
	}
	return r
}

func buildIpamResources(r *types.NetworkResource, nwInfo libnetwork.NetworkInfo) {
	id, opts, ipv4conf, ipv6conf := nwInfo.IpamConfig()

	ipv4Info, ipv6Info := nwInfo.IpamInfo()

	r.IPAM.Driver = id

	r.IPAM.Options = opts

	r.IPAM.Config = []network.IPAMConfig{}
	for _, ip4 := range ipv4conf {
		if ip4.PreferredPool == "" {
			continue
		}
		iData := network.IPAMConfig{}
		iData.Subnet = ip4.PreferredPool
		iData.IPRange = ip4.SubPool
		iData.Gateway = ip4.Gateway
		iData.AuxAddress = ip4.AuxAddresses
		r.IPAM.Config = append(r.IPAM.Config, iData)
	}

	if len(r.IPAM.Config) == 0 {
		for _, ip4Info := range ipv4Info {
			iData := network.IPAMConfig{}
			iData.Subnet = ip4Info.IPAMData.Pool.String()
			iData.Gateway = ip4Info.IPAMData.Gateway.String()
			r.IPAM.Config = append(r.IPAM.Config, iData)
		}
	}

	hasIpv6Conf := false
	for _, ip6 := range ipv6conf {
		if ip6.PreferredPool == "" {
			continue
		}
		hasIpv6Conf = true
		iData := network.IPAMConfig{}
		iData.Subnet = ip6.PreferredPool
		iData.IPRange = ip6.SubPool
		iData.Gateway = ip6.Gateway
		iData.AuxAddress = ip6.AuxAddresses
		r.IPAM.Config = append(r.IPAM.Config, iData)
	}

	if !hasIpv6Conf {
		for _, ip6Info := range ipv6Info {
			iData := network.IPAMConfig{}
			iData.Subnet = ip6Info.IPAMData.Pool.String()
			iData.Gateway = ip6Info.IPAMData.Gateway.String()
			r.IPAM.Config = append(r.IPAM.Config, iData)
		}
	}
}

func buildEndpointResource(id string, name string, info libnetwork.EndpointInfo) types.EndpointResource {
	er := types.EndpointResource{}

	er.EndpointID = id
	er.Name = name
	ei := info
	if ei == nil {
		return er
	}

	if iface := ei.Iface(); iface != nil {
		if mac := iface.MacAddress(); mac != nil {
			er.MacAddress = mac.String()
		}
		if ip := iface.Address(); ip != nil && len(ip.IP) > 0 {
			er.IPv4Address = ip.String()
		}

		if ipv6 := iface.AddressIPv6(); ipv6 != nil && len(ipv6.IP) > 0 {
			er.IPv6Address = ipv6.String()
		}
	}
	return er
}
