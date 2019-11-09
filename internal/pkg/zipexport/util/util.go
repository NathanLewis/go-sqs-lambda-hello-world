package util

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
	//pkcs12 "github.com/develar/go-pkcs12"
)

// Configuration for aws s3 bucket
type Configuration struct {
	AccessKey           string
	SecretKey           string
	Bucket              string
	Region              string
	Token               string
	Port                int
	CertKeyFileLocation string
	CertCrtFileLocation string
	AwsAccountID        string
}

//Config used to load the config valiues.
func (conf *Configuration) Config() Configuration {

	if (Configuration{}) == config {
		configFile, _ := os.Open("conf.json")
		decoder := json.NewDecoder(configFile)
		err := decoder.Decode(&config)
		if err != nil {
			panic("Error reading conf")
		}
	}
	return config
}

var config = Configuration{}

//UgcCert holder for certs
type UgcCert struct{}

//Pcks12Cert used to get pcks12Cert
func (ugcCert *UgcCert) Pcks12Cert() tls.Certificate {

	config = config.Config()
	//data, err := ioutil.ReadFile(WormholePcks12())
	pemData, err := ioutil.ReadFile(config.CertKeyFileLocation)
	certData, err := ioutil.ReadFile(config.CertCrtFileLocation)
	if err != nil {
		log.WithFields(log.Fields{
			"Error": fmt.Sprintf("%+v", err),
		}).Fatal("Problems reading cert")
	}
	// then use PEM data for tls to construct tls certificate:
	cert, err := tls.X509KeyPair(certData, pemData)
	if err != nil {
		panic(err)
	}
	return cert
}

//CertPool used for wormhole
func (ugcCert *UgcCert) CertPool() *x509.CertPool {

	config = config.Config()
	certPool, _ := x509.SystemCertPool()
	pemData, _ := ioutil.ReadFile(config.CertKeyFileLocation)
	certPool.AppendCertsFromPEM(pemData)
	return certPool
}
