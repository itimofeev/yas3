package store

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"log"
	"log/slog"
	"testing"
	"time"

	"github.com/quic-go/quic-go"
)

func TestClientConnect(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second) // 3s handshake timeout
	defer cancel()

	quicConfig := &quic.Config{Versions: []quic.Version{quic.Version2}}
	tlsConfig := &tls.Config{
		RootCAs:    getRootCA(),
		NextProtos: []string{"quic-echo"},
	}

	sess, err := quic.DialAddr(ctx, "localhost:9090", tlsConfig, quicConfig)
	if err != nil {
		log.Fatal(err)
	}

	// Открытие потока для передачи данных
	stream, err := sess.OpenUniStream()
	if err != nil {
		log.Fatal(err)
	}

	// Отправка сообщения на сервер
	message := "Привет от QUIC-клиента!"
	_, err = stream.Write([]byte(message))
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Отправлено сообщение: %s", message)

	<-stream.Context().Done()
	slog.Info("stream closed")
}

func getRootCA() *x509.CertPool {
	p, _ := pem.Decode([]byte(rootCA))
	if p.Type != "CERTIFICATE" {
		panic("expected a certificate")
	}

	caCert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		panic(err)
	}

	certPool := x509.NewCertPool()
	certPool.AddCert(caCert)

	return certPool
}

const rootCA = `-----BEGIN CERTIFICATE-----
MIIDMTCCAhmgAwIBAgIUEHLr7ydHz4+E2R79VRGGvV8EcMEwDQYJKoZIhvcNAQEL
BQAwKDEmMCQGA1UECgwdcXVpYy1nbyBDZXJ0aWZpY2F0ZSBBdXRob3JpdHkwHhcN
MjQwOTIwMDgyNjA1WhcNMzQwOTE4MDgyNjA1WjAoMSYwJAYDVQQKDB1xdWljLWdv
IENlcnRpZmljYXRlIEF1dGhvcml0eTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCC
AQoCggEBANyMVzoNzegwhBKr2Ml3Re/KZZO0Mhv4RFuLdNu2pf+fuDTJU4DQx3Iu
hAu4F20Yo2w6AZyCXJuFeCGF2gcme0Sk9tlS55xRp4WM1kFX+u+Qg43ccdiFS4cE
QKbqCffeDGN680/XAwawPNrdbVCTmJiyFO6Yw6ppo+3ggquJs5CKwXWxlQzyMANq
bmrPDXLb3c85fR98vfUNQypgxnnbCWxpItyprB/gDx+/dxn2BggQ90gBHqqlRndz
+Rt7oO0EJHvn0VHFkLYapPjiCPwC5H6T5MZ0tmudUmjF3hBFFePCsQzufZPweBkD
1mvbdtnm8+DEYJc5XCstZTQVKU3ZHEMCAwEAAaNTMFEwHQYDVR0OBBYEFAX/6ghL
HNt6p9cS671yS7Hl/98nMB8GA1UdIwQYMBaAFAX/6ghLHNt6p9cS671yS7Hl/98n
MA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEBAMdngL/b3l17JTVK
UETXgzbPiAS9Jym10+h+iKXJtEpGHamWGiX0kQuboW2OMlW6zdyNbU5QPu0TCsLS
5djuM3tbkxCWmhCgL65eon5WCEZnqsny5GVZefvcPTJ7on2JwMzouLAudwiRfYF3
jqkrgovm7Bn1YwpWZmmYv+7dUZVC7M+O6lfYuqs/2AY2TW2B1VaOgvschPgpOnX6
G3qUejLtO99TH8ip2KHllbo8IMkCTU3+nYuJYQo0kvb/0uFA7whiZe5TVu5w/+Tk
RyPYtm1UvFehBAzPx3xHrZFZc1WW1U9YtPC1ppWkLLRdZHuePfcsfQVpUKG4DDMz
eW8vj+Y=
-----END CERTIFICATE-----
`
