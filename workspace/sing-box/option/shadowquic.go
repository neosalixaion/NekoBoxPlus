package option

type ShadowQUICOutboundOptions struct {
	DialerOptions
	ServerOptions
	Username          string      `json:"username"`
	Password          string      `json:"password"`
	ServerName        string      `json:"server_name"`
	ALPN              []string    `json:"alpn,omitempty"`
	CongestionControl string      `json:"congestion_control,omitempty" enum:"cubic,new_reno,bbr"`
	UDPOverStream     bool        `json:"udp_over_stream,omitempty"`
	ZeroRTTHandshake  bool        `json:"zero_rtt_handshake,omitempty"`
	Network           NetworkList `json:"network,omitempty"`
}
