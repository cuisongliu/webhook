/*
Copyright © 2021 cuisongliu@qq.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package webhook

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
)

type CertConfig struct {
	CommonName   string
	Organization []string
	// AltNames contains the domain names and IP addresses that will be added
	// to the API Server's x509 certificate SubAltNames field. The values will
	// be passed directly to the x509.Certificate object.
	AltNames struct {
		DNSNames []string
		IPs      []net.IP
	}
}

// NewPrivateKey creates an RSA private key
func NewPrivateKey(keyType x509.PublicKeyAlgorithm) (crypto.Signer, error) {
	if keyType == x509.ECDSA {
		return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	}

	return rsa.GenerateKey(rand.Reader, 2048)
}

func NewSigned(cfg CertConfig) (csr, keyPEM []byte, err error) {
	key, err := NewPrivateKey(x509.RSA)
	if err != nil {
		return nil, nil, fmt.Errorf("new signed private failed %s", err)
	}
	pk := x509.MarshalPKCS1PrivateKey(key.(*rsa.PrivateKey))
	keyPEM = pem.EncodeToMemory(&pem.Block{
		Type: "RSA PRIVATE KEY", Bytes: pk,
	})
	_, csr, err = GenerateCSR(cfg, key)

	csr = pem.EncodeToMemory(&pem.Block{
		Type: "CERTIFICATE REQUEST", Bytes: csr,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("new signed csr failed %s", err)
	}

	return
}

// GenerateCSR will generate a new *x509.CertificateRequest template to be used
// by issuers that utilise CSRs to obtain Certificates.
// The CSR will not be signed, and should be passed to either EncodeCSR or
// to the x509.CreateCertificateRequest function.
func GenerateCSR(cfg CertConfig, key crypto.Signer) (*x509.CertificateRequest, []byte, error) {
	if len(cfg.CommonName) == 0 {
		return nil, nil, errors.New("must specify a CommonName")
	}

	var dnsNames []string
	var ips []net.IP

	for _, v := range cfg.AltNames.DNSNames {
		dnsNames = append(dnsNames, v)
	}
	for _, v := range cfg.AltNames.IPs {
		ips = append(ips, v)
	}
	certTmpl := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},

		DNSNames:    dnsNames,
		IPAddresses: ips,
	}
	certDERBytes, err := x509.CreateCertificateRequest(rand.Reader, &certTmpl, key)
	if err != nil {
		return nil, nil, err
	}
	r1, r3 := x509.ParseCertificateRequest(certDERBytes)
	return r1, certDERBytes, r3
}

func (c *CertWebHook) generateTLS() (csr []byte, key []byte, err error) {
	host := fmt.Sprintf("%s.%s", c.ServiceName, c.Namespace)
	dnsNames := []string{
		host,
		fmt.Sprintf("%s.svc", host),
		fmt.Sprintf("%s.svc.cluster.local", host),
	}
	cfg := CertConfig{
		CommonName:   host,
		Organization: c.Subject,
		AltNames: struct {
			DNSNames []string
			IPs      []net.IP
		}{
			DNSNames: dnsNames,
		},
	}
	csr, key, err = NewSigned(cfg)
	return
}
