package store

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/quic-go/quic-go/http3"
)

func TestClientConnect(t *testing.T) {
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs: getRootCA(),
		},
	}
	defer roundTripper.Close()
	client := &http.Client{
		Transport: roundTripper,
	}

	addr := "https://localhost:9090/api/v1/hello"
	rsp, err := client.Get(addr)
	if err != nil {
		panic(err)
	}
	defer rsp.Body.Close()

	body := &bytes.Buffer{}
	_, err = io.Copy(body, rsp.Body)
	if err != nil {
		panic(err)
	}

	log.Printf("Body length: %d bytes \n", body.Len())
	log.Printf("Response body %s \n", body.Bytes())
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
