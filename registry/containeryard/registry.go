package containeryard

import (
	// "crypto/tls"
	// "crypto/x509"
	"fmt"
	"github.com/ngageoint/seed-silo/registry/v2"
	// "io/ioutil"

	"net/http"
	"os"
	"strings"

	"github.com/ngageoint/seed-common/util"
)

//ContainerYardRegistry type representing a Container Yard registry
type ContainerYardRegistry struct {
	URL      string
	Hostname string
	Client   *http.Client
	Org      string
	Username string
	Password string
	v2Base   *v2.V2registry
	Print    util.PrintCallback
}

func (r *ContainerYardRegistry) Name() string {
	return "ContainerYardRegistry"
}

//New creates a new docker hub registry from the given URL
func New(registryUrl, org, username, password string) (*ContainerYardRegistry, error) {
	if util.PrintUtil == nil {
		util.InitPrinter(util.PrintErr, os.Stderr, os.Stdout)
	}
	url := strings.TrimSuffix(registryUrl, "/")
	reg, err := v2.New(url, org, username, password)

	host := strings.Replace(url, "https://", "", 1)
	host = strings.Replace(host, "http://", "", 1)

	// if _, err := os.Stat("cert.pem"); err == nil {
	// 	caCertPool := x509.NewCertPool()
	// 	caCert, err := ioutil.ReadFile("cert.pem")
	// 	cert, err := tls.LoadX509KeyPair("cert.pem", "key.pem")

	// 	caCertPool.AppendCertsFromPEM(caCert)
	// 	// if err != nil {
	// 	// log.(err)
	// 	client := &http.Client{
	// 		Transport: &http.Transport{
	// 			TLSClientConfig: &tls.Config{
	// 				RootCAs:      caCertPool,
	// 				Certificates: []tls.Certificate{cert},
	// 			},
	// 		},
	// 	}

	// 	registry := &ContainerYardRegistry{
	// 		URL:      url,
	// 		Hostname: host,
	// 		Client:   client,
	// 		Org:      org,
	// 		Username: username,
	// 		Password: password,
	// 		v2Base:   reg,
	// 		Print:    util.PrintUtil,
	// 	}

	// 	return registry, err
	// }

	client := &http.Client{}

	registry := &ContainerYardRegistry{
		URL:      url,
		Hostname: host,
		Client:   client,
		Org:      org,
		Username: username,
		Password: password,
		v2Base:   reg,
		Print:    util.PrintUtil,
	}

	return registry, err

	// Create a CA certificate pool and add cert.pem to it
	// caCert, err := ioutil.ReadFile("cert.pem")
	// if err != nil {
	// 	log.Fatal(err)
	// }
}

func (r *ContainerYardRegistry) url(pathTemplate string, args ...interface{}) string {
	pathSuffix := fmt.Sprintf(pathTemplate, args...)
	url := fmt.Sprintf("%s%s", r.URL, pathSuffix)
	return url
}

func (r *ContainerYardRegistry) Ping() error {
	//query that should quickly return an empty json response
	url := r.url("/search?q=NoImagesWithThisName&t=json")
	var response Response
	err := r.getContainerYardJson(url, &response)
	return err
}
