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
package v1

import (
	"net"
	"testing"
)

func TestCert_GenerateTLS(t *testing.T) {
	type fields struct {
		Subject []string
	}
	type args struct {
		namespace string
		service   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{name: "generator", fields: fields{
			Subject: nil,
		}, args: args{
			namespace: "default",
			service:   "svc",
		}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CertWebHook{
				Subject:     tt.fields.Subject,
				Namespace:   tt.args.namespace,
				ServiceName: tt.args.service,
			}
			csr, key, err := c.generateTLS()
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateTLS() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			t.Logf("GenerateTLS() csr = \n%s", string(csr))
			t.Logf("GenerateTLS() key = \n%s", string(key))
		})
	}
}

func TestNewSignedCa(t *testing.T) {
	type args struct {
		cfg CertConfig
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "certs", args: args{
			cfg: CertConfig{
				CommonName: "dffff",
				Organization: []string{
					"DFFFFERR",
				},
				AltNames: struct {
					DNSNames []string
					IPs      []net.IP
				}{
					DNSNames: []string{
						"default.namespace",
					},
				},
			}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			csr, key, err := NewSigned(tt.args.cfg)
			if err != nil {
				t.Error(err)
			}
			t.Log(string(csr))
			t.Log(string(key))

		})
	}
}
