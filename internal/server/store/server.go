package store

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/quic-go/quic-go"
)

type Config struct {
	Port int
}

type Server struct {
	ln *quic.Listener
}

func New(cfg Config) (*Server, error) {
	cert, err := tls.X509KeyPair([]byte(certPEM), []byte(keyPEM))
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"quic-echo"},
	}
	quicConf := &quic.Config{}

	ln, err := quic.ListenAddr(":"+fmt.Sprint(cfg.Port), tlsConfig, quicConf)
	if err != nil {
		return nil, err
	}
	slog.Info("listening for incoming connections", "port", cfg.Port)

	return &Server{
		ln: ln,
	}, nil
}

func (s Server) Run(ctx context.Context) error {
	conn, err := s.ln.Accept(ctx)
	if err != nil {
		return err
	}
	slog.Info("new connection accepted", "remote", conn.RemoteAddr().String(), "local", conn.LocalAddr().String())

	for {
		receiveStream, err := conn.AcceptUniStream(ctx)
		var netErr net.Error
		if errors.As(err, &netErr) {
			slog.Error("error on accepting new stream", "err", netErr.Error())
		}
		slog.Info("stream accepted", "streamID", receiveStream.StreamID())
		_ = receiveStream
	}

	return nil
}

const certPEM = `-----BEGIN CERTIFICATE-----
MIIDIDCCAgigAwIBAgIUIAyUM++OWOEC9kuauODr7gnVFgMwDQYJKoZIhvcNAQEL
BQAwKDEmMCQGA1UECgwdcXVpYy1nbyBDZXJ0aWZpY2F0ZSBBdXRob3JpdHkwHhcN
MjQwOTIwMDgyNjA2WhcNMzQwOTE4MDgyNjA2WjASMRAwDgYDVQQKDAdxdWljLWdv
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAmbUBMWdiYduhtK8ImkO7
AUZ1AJapdf51hPXP71V5bcPnjbKyjVQXNX3+eBf//ozo5M373kAmtmC71/7X0K3u
ZaNiCS8SSnznEyBS+82rCWuszL5kHzM40i5zasYrREGcCNPCTCP9KcJtVZcZ8Txt
hqwDcksk8UQ4npQ3mW4ft7xDNw0H45dzdmGsW3JJu2oJ+DLntUk08V4szVruKrrx
nGz1RwF5r2j0Oqvaimh4UaMPGQF5+QIV6+RQXZ86vMLN8HTkRE3V+zf0PGyCHG5I
obqJENZItU6sVCRLNW8fsINLv/l1FJrlpdK3nk/mi/jMZ+Nei7c8KT1nYassQiNV
BQIDAQABo1gwVjAUBgNVHREEDTALgglsb2NhbGhvc3QwHQYDVR0OBBYEFGsSxJ0k
/0x/BdB3jF64WoDYJMV0MB8GA1UdIwQYMBaAFAX/6ghLHNt6p9cS671yS7Hl/98n
MA0GCSqGSIb3DQEBCwUAA4IBAQB9j0W6Jf8hzsy9nKCmy8dqKaMnPSHqEc5fg+fq
YlaFn9IfmshUJSAAblqmyeOZ9xyf9XWOSz/OcqfmXjfdk+Ataisdt7LXVBPoQv1l
3nEMvwK2FVUS7xOt+JAVRR7dE2gLVNuXgdoFI/YM6dTgHeVB/WPIAvYd+vEV17z3
XYxD9St5rKieZdWBtkpd8zh+sqOAdt/RBAN8qSWsEVs5eGl99FWjc3OhIvIO5cfD
t+7IQ5zpLNNiE65C7OlBh2BxYHpLCdOVLauwmfOCRcDkSco2VPzZQo3UKGT/bv7n
pIqNEEumAfOomh2izV6FDBzc2xR0Lo9bUT2wstcm+OPCbT4a
-----END CERTIFICATE-----`

const keyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQCZtQExZ2Jh26G0
rwiaQ7sBRnUAlql1/nWE9c/vVXltw+eNsrKNVBc1ff54F//+jOjkzfveQCa2YLvX
/tfQre5lo2IJLxJKfOcTIFL7zasJa6zMvmQfMzjSLnNqxitEQZwI08JMI/0pwm1V
lxnxPG2GrANySyTxRDielDeZbh+3vEM3DQfjl3N2Yaxbckm7agn4Mue1STTxXizN
Wu4quvGcbPVHAXmvaPQ6q9qKaHhRow8ZAXn5AhXr5FBdnzq8ws3wdORETdX7N/Q8
bIIcbkihuokQ1ki1TqxUJEs1bx+wg0u/+XUUmuWl0reeT+aL+Mxn416LtzwpPWdh
qyxCI1UFAgMBAAECggEALn2bf0RscvnZ9sssWHCdCv0zHXr0ha5yAEXTX2okgMlM
68R9kha5sGHMCqoDsYwQU0hkYqxXqTkoB+RahZFeNprM58c3ipUt1VClGOlzzrUl
PoZlxTQafyQyn7yR0KLhnZ/jOGF4TN20cTtzSs6CuEWmAzdsVJdUYs6k5ID1Ef7k
m9Zpp8DmWSWiP9/sfW1SiNCfzZwzWsf+hrFra1BlX/FuFMXYER+63mK737UUgcq7
/pZCqqf0zUj3LKGBYNd1Bsh/mPn4+V9vtPJyTzPHKN1kO6/4ECiu+wh3X0Q0iw7y
6IItcYmTFZU+k2I0m4dCXJE5CfclIPuLM3Z+LabWXwKBgQDWs/IGsBgM98Er8jan
Q4Ghcb0JtHQQ1JsbUX47nO4/UhHqS6Eo2bGJ8QsDP1n5RfIxtnGh5I+5rtV0zmDf
JMe60fIPln8/KGMgcUnPYfHgNZhNnxrUG9F4xLJw1ejrPGcK/UdNhO00/r8a+QwX
xc71oBDF48NKRWYCki0yY2D0pwKBgQC3RZlus7nm+0dxmQ9bzthS2kzRClgaKXCL
34z8rKERhLnMENW32UOfcVOVF/eM+cnBBt1kn9Avf8jPLfzcdtMZZIBijyqJEKX4
2NepMg+wH1yzradJOwqRyFc8ulUWK/0hEZB69GFFztfKEjIvkpPlLr6dtAg7hwqK
LcrIrtbicwKBgDlwaLaLU9PcUGyuXxq+f6auZBF9mnOKPXjAg5H1OPPtw+c3loT0
QIAT4Ytb3nlG0jWkhp/2ItFdSbP3JolsMJb1ZdnvvFksN+DNDh8SKACAth9GCopm
atLxZH+1apvMTBDvk6zUfBVqdbwElsyhWe3yhao7ddqf2Fulubu6RI0PAoGAcgXl
TdCXyrNvYae+vHnWcMXMoQn0gmJh2UQ+bT3iAAo5plKbBQUxY1OYktwUcis+cM+a
km4zkjnIb83G0ktDTzsN/UPhTOGEwWv30RaKWCNLA6b4u0D2dHjWfbvYEDFSDW7m
GvlMQ6hK7Teg7aQeS6pngapurMp5rjqLPYw5uS0CgYBIWi56OQiMv0NApuYi73M4
asfAxkNcImzvgh56kFcA5/JF2bcXjMdDIQT1lFjBUsUDXvWDE3OKenljSipL4x3B
uv4lpMutgwbsC/sCbQv5qMcffeEV4+1CR2RLWt2VH5MKf5oW+yMrs7r+mc8ejPts
Tcvn605B+xRcNGJ3Qhowyg==
-----END PRIVATE KEY-----`
