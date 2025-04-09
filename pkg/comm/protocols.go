package comm

type Protocol interface{}

type TCP struct{}
type TLS struct{}
type UDP struct{}

func GetProtocolString(proto Protocol) string {
	switch proto.(type) {
	case TCP:
		return "tcp"
	case TLS:
		return "tls"
	case UDP:
		return "udp"
	}
	return "unknown"
}
