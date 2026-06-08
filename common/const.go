package common

const (
	udp     = "udp"
	tcp     = "tcp"
	keyPath = "key/id_ed25519"

	UDP = "udp"
	TCP = "tcp"

	MinKeyLen = 6

	// 所有端口写死，无需配置文件
	PortHTTP    = ":21114"
	PortNAT     = ":21115"
	PortSignal  = ":21116"
	PortRelay   = ":21117"
	PortWS      = ":21118"

	// must_key 默认关闭，按需在此改为 true
	MustKey = false
)
