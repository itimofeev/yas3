package store

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"github.com/quic-go/quic-go/http3"

	"github.com/itimofeev/yas3/internal/entity"
)

type Config struct {
	StoreAddr string `validate:"required"`
}

type Client struct {
	httpClient http.Client
	cfg        Config
}

func New(cfg Config) (*Client, error) {
	roundTripper := &http3.RoundTripper{
		TLSClientConfig: &tls.Config{
			RootCAs: getRootCA(),
		},
	}

	httpClient := http.Client{
		Transport: roundTripper,
	}

	return &Client{
		httpClient: httpClient,
		cfg:        cfg,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, fileName string, content io.Reader) error {
	slog.Debug("starting upload file to store server", "fileName", fileName, "serverId", c.GetID())
	url := c.cfg.StoreAddr + "/api/v1/uploadFile/" + fileName
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, content)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response code not 200: %d", resp.StatusCode)
	}

	slog.Debug("file uploaded to store server", "fileName", fileName, "serverId", c.GetID())
	return nil
}

func (c *Client) GetFile(ctx context.Context, fileName string) (io.ReadCloser, error) {
	url := c.cfg.StoreAddr + "/api/v1/getFile/" + fileName
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, err
	}

	return resp.Body, nil
}

func (c *Client) GetAvailableSpace(ctx context.Context) (entity.AvailableSpace, error) {
	url := c.cfg.StoreAddr + "/api/v1/getAvailableSpace"
	uploadReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return entity.AvailableSpace{}, err
	}
	resp, err := c.httpClient.Do(uploadReq)
	if err != nil {
		return entity.AvailableSpace{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return entity.AvailableSpace{}, fmt.Errorf("response code not 200: %d", resp.StatusCode)
	}

	m := make(map[string]int64)
	if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
		return entity.AvailableSpace{}, err
	}

	return entity.AvailableSpace{
		Total: m["total"],
		Used:  m["used"],
	}, nil
}

func (c *Client) GetID() string {
	return c.cfg.StoreAddr
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
