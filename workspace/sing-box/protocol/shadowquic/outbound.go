package shadowquic

import (
	"context"
	"net"
	"sync"

	shadowquic "github.com/exclavenetwork/sing-shadowquic"
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/outbound"
	"github.com/sagernet/sing-box/common/dialer"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common/bufio"
	E "github.com/sagernet/sing/common/exceptions"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/service"
	"github.com/sagernet/sing/service/pause"
)

const outboundType = "shadowquic"

func RegisterOutbound(registry *outbound.Registry) {
	outbound.Register[option.ShadowQUICOutboundOptions](registry, outboundType, NewOutbound)
}

var (
	_ adapter.InterfaceUpdateListener = (*Outbound)(nil)
	_ adapter.IdleConnectionKeeper    = (*Outbound)(nil)
	_ adapter.OutboundWithMultiplex    = (*Outbound)(nil)
)

type Outbound struct {
	outbound.Adapter
	client *shadowquic.Client

	idleAccess sync.Mutex
	keepIdle   bool
	paused     bool
	unregisterPauseCallback func()
}

func (h *Outbound) applyKeepIdleLocked() {
	h.client.SetKeepIdleConnections(h.keepIdle && !h.paused)
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

	h := &Outbound{
		Adapter: outbound.NewAdapterWithDialerOptions(outboundType, tag, options.Network.Build(), options.DialerOptions),
		client:  client,
		keepIdle: true,
	}
	if pauseManager := service.FromContext[pause.Manager](ctx); pauseManager != nil {
		h.paused = pauseManager.IsPaused()
		callback := pauseManager.RegisterCallback(func(_ int) {
			h.idleAccess.Lock()
			h.paused = pauseManager.IsPaused()
			h.applyKeepIdleLocked()
			h.idleAccess.Unlock()
		})
		h.unregisterPauseCallback = func() {
			pauseManager.UnregisterCallback(callback)
		}
		h.idleAccess.Lock()
		h.applyKeepIdleLocked()
		h.idleAccess.Unlock()
	}
	return h, nil
}

func (h *Outbound) DialContext(ctx context.Context, network string, destination M.Socksaddr) (net.Conn, error) {
	switch N.NetworkName(network) {
	case N.NetworkTCP:
		return h.client.DialConn(ctx, destination)
	case N.NetworkUDP:
		packetConn, err := h.client.ListenPacket(ctx)
		if err != nil {
			return nil, err
		}
		return bufio.NewBindPacketConn(packetConn, destination), nil
	default:
		return nil, E.New("unsupported network: ", network)
	}
}

func (h *Outbound) ListenPacket(ctx context.Context, destination M.Socksaddr) (net.PacketConn, error) {
	return h.client.ListenPacket(ctx)
}

func (h *Outbound) InterfaceUpdated(ctx context.Context) {
	_ = h.client.Close()
}

func (h *Outbound) MultiplexEnabled() bool {
	return true
}

func (h *Outbound) SetKeepIdleConnections(keep bool) {
	h.idleAccess.Lock()
	h.keepIdle = keep
	h.applyKeepIdleLocked()
	h.idleAccess.Unlock()
}

func (h *Outbound) CloseIdleConnections() {
	h.client.CloseIdleConnections()
}

func (h *Outbound) Start(stage adapter.StartStage, scope *adapter.Scope) error {
	if stage != adapter.StartStateInitialize {
		return nil
	}
	scope.Add(func() error {
		if h.unregisterPauseCallback != nil {
			h.unregisterPauseCallback()
		}
		return h.client.Close()
	})
	return nil
}
