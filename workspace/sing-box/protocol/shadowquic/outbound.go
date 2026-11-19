package shadowquic

import (
	"context"
	"net"

	shadowquic "github.com/exclavenetwork/sing-shadowquic"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/bufio"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/uot"
)

const outboundType = "shadowquic"

func RegisterOutbound(registry *outbound.Registry) {
	outbound.Register[option.ShadowQUICOutboundOptions](registry, outboundType, NewOutbound)
}

var (
	_ adapter.InterfaceUpdateListener = (*Outbound)(nil)
	_ adapter.IdleConnectionKeeper    = (*Outbound)(nil)
	_ adapter.OutboundWithMultiplex   = (*Outbound)(nil)
)

type Outbound struct {
	outbound.Adapter
	client    *shadowquic.Client
	udpStream bool
}

func NewOutbound(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.ShadowQUICOutboundOptions) (adapter.Outbound, error) {
	if options.Username == "" || options.Password == "" || options.ServerName == "" {
		return nil, E.New("shadowquic: username, password and server_name are required")
	}
	outboundDialer, err := dialer.New(ctx, options.DialerOptions, options.ServerIsDomain())
	if err != nil {
		return nil, err
	}
	client, err := shadowquic.NewClient(shadowquic.ClientOptions{
		Context:           ctx,
		Dialer:            outboundDialer,
		ServerAddress:     options.ServerOptions.Build(),
		Username:          options.Username,
		Password:          options.Password,
		ServerName:        options.ServerName,
		NextProtos:        options.ALPN,
		CongestionControl: options.CongestionControl,
		UDPOverStream:     options.UDPOverStream,
		ZeroRTTHandshake:  options.ZeroRTTHandshake,
	})
	if err != nil {
		return nil, err
	}
	return &Outbound{
		Adapter:   outbound.NewAdapterWithDialerOptions(outboundType, tag, options.Network.Build(), options.DialerOptions),
		client:    client,
		udpStream: options.UDPOverStream,
	}, nil
}

func (h *Outbound) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	switch N.NetworkName(network) {
	case N.NetworkTCP:
		return h.client.DialConn(ctx, destination)
	case N.NetworkUDP:
		if h.udpStream {
			streamConn, err := h.client.DialConn(ctx, uot.RequestDestination(uot.Version))
			if err != nil {
				return nil, err
			}
			return uot.NewLazyConn(streamConn, uot.Request{
				IsConnect: true,
				Destination: destination,
			}), nil
		}
		packetConn, err := h.ListenPacket(ctx, destination)
		if err != nil {
			return nil, err
		}
		return bufio.NewBindPacketConn(packetConn, destination), nil
	default:
		return nil, E.New("unsupported network: ", network)
	}
}

func (h *Outbound) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	if h.udpStream {
		streamConn, err := h.client.DialConn(ctx, uot.RequestDestination(uot.Version))
		if err != nil {
			return nil, err
		}
		return uot.NewLazyConn(streamConn, uot.Request{
			IsConnect: false,
			Destination: destination,
		}), nil
	}
	return h.client.ListenPacket(ctx)
}

func (h *Outbound) InterfaceUpdated(ctx context.Context) {
	_ = h.client.Close()
}

func (h *Outbound) MultiplexEnabled() bool {
	return true
}

func (h *Outbound) SetKeepIdleConnections(keep bool) {
	h.client.SetKeepIdleConnections(keep)
}

func (h *Outbound) CloseIdleConnections() {
	h.client.CloseIdleConnections()
}

func (h *Outbound) Start(stage adapter.StartStage, scope *adapter.Scope) error {
	if stage != adapter.StartStateInitialize {
		return nil
	}
	scope.Add(func() error {
		return h.client.Close()
	})
	return nil
}
