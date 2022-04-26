# Bhojpur VPN - Networking Framework

The Bhojpur VPN is a decentralized, virtual private networking system applied within the
[Bhojpur.NET Platform](https://github.com/bhojpur/platform) for delivery of distributed
`applications` or `services`. It uses libp2p to build private decentralized networks that
could be accessed via shared secrets.

It can:

- **Create a VPN** :  Secure VPN between P2P peers
  - Automatically assign IPs to nodes
  - Embedded tiny DNS server to resolve internal/external IPs
  - Create trusted zones to prevent network access if token is leaked

- **Act as a reverse Proxy** : Share a TCP service. The Bhojpur VPN let exposes TCP services
to the P2P network nodes without establishing a VPN connection: creates reverse proxy and
tunnels traffic into the P2P network.

- **Send files via P2P** : Send files over P2P between nodes without establishing a VPN connection.

- **Be used as a library**: Plug a distributed P2P ledger easily in your project!

## Installation

Download the precompiled static release in the [releases page](https://github.com/bhojpur/vpn/releases).
You can either install it in your system or just run it.

## Simple Usage

The Bhojpur VPN works by generating tokens (or, a configuration file) that can be shared
between different machines, hosts, or peers to access to a decentralized secured network
between them. Every token is unique and identifies the network, no central server setup,
or specifying hosts IP is required.

To generate a config run:

```bash
# Generate a new config file and use it later as BHOJPUR_VPN_CONFIG
$ vpnsvr -g > config.yaml
```

or, to generate a portable token:

```bash
$ BHOJPUR_VPN_TOKEN=$(vpnsvr -g -b)
```

*NOTE*: Tokens are config merely encoded in base64, so this is equivalent:

```bash
$ BHOJPUR_VPN_TOKEN=$(vpnsvr -g | tee config.yaml | base64 -w0)
```

All Bhojpur VPN commands implies that you either specify a `BHOJPUR_VPN_TOKEN`
(`--token` as parameter) or a `BHOJPUR_VPN_CONFIG` as this is the way for `vpnsvr`
to establish a network between the nodes.

The configuration file is the network definition and allows you to connect over to
your peers securely.

**Warning** Exposing this file or passing-it by is equivalent to give full control
to the network.

### As a VPN

To start the VPN, simply run `vpnsvr` without any argument.

An example of running Bhojpur VPN on multiple hosts:

```bash
# on Node A
$ BHOJPUR_VPN_TOKEN=.. vpnsvr --address 10.1.0.11/24
# on Node B
$ BHOJPUR_VPN_TOKEN=.. vpnsvr --address 10.1.0.12/24
# on Node C ...
$ BHOJPUR_VPN_TOKEN=.. vpnsvr --address 10.1.0.13/24
...
```

... and that's it! the `--address` is a _virtual_ unique IP for each node, and it
is actually the IP, where the node will be reachable to from the VPN. You can assign
IPs freely to the nodes of the network, while you can override the default
`bhojpurvpn0` interface with `IFACE` (or `--interface`)

*NOTE*: It might take up time to build the connection between nodes. Wait at least
5 mins, it depends on the network behind the hosts.

## Use Case: [Bhojpur DCP](https://github.com/bhojpur/dcp) test cluster

Let's say you are developing something for the Kubernetes and you would like to 
a multi-node setup, but you have machines available only behind NAT; and, you would
really like to leverage hardware.

1) Generate `Bhojpur VPN` config: `vpnsvr -g > vpn.yaml`
2) Start the VPN:

   on node A: `sudo IFACE=bhojpurvpn0 ADDRESS=10.1.0.3/24 BHOJPUR_VPN_CONFIG=vpn.yml vpnsvr`
   
   on node B: `sudo IFACE=bhojpurvpn0 ADDRESS=10.1.0.4/24 BHOJPUR_VPN_CONFIG=vpm.yml vpnsvr`
3) Start [Bhojpur DCP](https://gihub.com/bhojpur/dcp):
 
   on node A: `dcp server --flannel-iface=bhojpurvpn0`
   
   on node B: `DCP_URL=https://10.1.0.3:6443 DCP_TOKEN=xx dcp agent --flannel-iface=bhojpurvpn0 --node-ip 10.10.4`

We have used flannel here, but other CNI should work as well.

Don't miss out [Bhojpur OS](https://github.com/bhojpur/os), which is a Linux derivative
built on top of [Bhojpur DCP](https://gihub.com/bhojpur/dcp) and `Bhojpur VPN` for automatic
node discovery!

## As a library

The `Bhojpur VPN` could be used as a library. It is very portable and offers a functional
interface. To join a node in a network from a token, without starting the VPN:

```golang
import (
    node "github.com/bhojpur/vpn/pkg/node"
)

e := node.New(
    node.Logger(l),
    node.LogLevel(log.LevelInfo),
    node.MaxMessageSize(2 << 20),
    node.FromBase64( mDNSEnabled, DHTEnabled, token ),
    // ....
  )

e.Start(ctx)

```

or, to start a Bhojpur VPN:

```golang
import (
    vpn "github.com/bhojpur/vpn/pkg/engine"
    node "github.com/bhojpur/vpn/pkg/node"
)

opts, err := vpn.Register(vpnOpts...)
if err != nil {
	return err
}

e := vpn.New(append(o, opts...)...)

e.Start(ctx)
```

## Troubleshooting

If during bootstrap you see messages like:

```bash
vpnsvr[3679]:             * [/ip4/104.131.131.82/tcp/4001] failed to negotiate stream multiplexer: context deadline exceeded     
```

or,

```bash
vpnsvr[9971]: 2021/12/16 20:56:34 failed to sufficiently increase receive buffer size (was: 208 kiB, wanted: 2048 kiB, got: 416 kiB)
```

or, generally experiencing poor network performance, it is recommended to increase the
maximum buffer size by running:

```bash
$ sysctl -w net.core.rmem_max=2500000
```
