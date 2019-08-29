package http

import (
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/NathanLewis/go-sqs-lambda-hello-world/internal/pkg/zipexport/util"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

//UgcHttp wraper around http client
type UgcHttp struct {
	Timeout time.Duration
	Client  *http.Client
	UgcCert *util.UgcCert
}

func (ugcHttp *UgcHttp) dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, ugcHttp.Timeout)
}

func (ugcHttp *UgcHttp) transportWithProxy(certificates []tls.Certificate) *http.Transport {
	proxyString := "http://www-cache.reith.bbc.co.uk:80"
	proxyURL, _ := url.Parse(proxyString)
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			//			Certificates:       []tls.Certificate{ugcHttp.UgcCert.ClientCert()},
			Certificates: certificates,
			//			RootCAs:      ugcHttp.UgcCert.RootCA(),
		},
		Dial:               ugcHttp.dialTimeout,
		MaxIdleConns:       10,
		IdleConnTimeout:    6000 * time.Second,
		DisableCompression: true,
	}
	return tr
}

func (ugcHttp *UgcHttp) transportWithoutProxy(certificates []tls.Certificate) *http.Transport {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
			//			Certificates:       []tls.Certificate{ugcHttp.UgcCert.ClientCert()},
			//Certificates: certificates,
			RootCAs: ugcHttp.UgcCert.CertPool(),
		},
		Dial:               ugcHttp.dialTimeout,
		MaxIdleConns:       10,
		IdleConnTimeout:    6000 * time.Second,
		DisableCompression: true,
	}
	return tr
}

func (ugcHttp *UgcHttp) wormholeTransport() *http.Transport {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateFreelyAsClient,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{ugcHttp.UgcCert.Pcks12Cert()},
			//RootCAs: ugcHttp.UgcCert.RootCA(),
		},
		Dial:               ugcHttp.dialTimeout,
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	return tr
}

func (ugcHttp *UgcHttp) wormholeTransportWithProxy() *http.Transport {
	proxyString := "http://www-cache.reith.bbc.co.uk:80"
	proxyURL, _ := url.Parse(proxyString)
	tr := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateFreelyAsClient,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{ugcHttp.UgcCert.Pcks12Cert()},
			//RootCAs: ugcHttp.UgcCert.RootCA(),
		},
		Dial:               ugcHttp.dialTimeout,
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	return tr
}

//InitClient used to initialize http client
func (ugcHttp *UgcHttp) InitClient(update bool) *http.Client {

	deployment, ok := os.LookupEnv("DEPLOYMENT")
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			Renegotiation:      tls.RenegotiateFreelyAsClient,
			InsecureSkipVerify: true,
			Certificates:       []tls.Certificate{ugcHttp.UgcCert.Pcks12Cert()},
		},
		Dial:               ugcHttp.dialTimeout,
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	log.WithFields(log.Fields{
		"Env": deployment,
		"Ok":  ok,
	}).Debug("Value of Deployment environment variable")

	if len(deployment) > 0 {
		tr = ugcHttp.wormholeTransport()
	} else {
		tr = ugcHttp.wormholeTransportWithProxy()
	}

	client := &http.Client{
		Transport: tr,
	}

	if update {
		ugcHttp.Client = client
	}

	return client
}

//FetchItemWormHole fetches worm hole details
func (ugcHttp *UgcHttp) FetchItemWormHole(url string, header map[string]string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)

	for k, v := range header {
		req.Header.Add(k, v)
	}

	client := ugcHttp.InitClient(false)
	resp, err := client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "FetchItem failed for [url=%v] [header=%v]", url, header)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read body for %s ", url)
	}

	log.WithFields(log.Fields{
		"URL":        url,
		"Body":       string(body),
		"StatusCode": resp.StatusCode,
	}).Debug("Response from http")

	defer resp.Body.Close()
	return body, nil
}

//FetchItem fetches items from url
func (ugcHttp *UgcHttp) FetchItem(url string, header map[string]string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)

	for k, v := range header {
		req.Header.Add(k, v)
	}

	resp, err := ugcHttp.Client.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "FetchItem failed for [url=%v] [header=%v]", url, header)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to read body for %s ", url)
	}

	log.WithFields(log.Fields{
		"URL":        url,
		"Body":       string(body),
		"StatusCode": resp.StatusCode,
	}).Debug("Response from http")

	defer resp.Body.Close()
	return body, nil
}
