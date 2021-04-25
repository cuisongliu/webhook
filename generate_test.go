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
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"testing"
)

func TestCertWebHook_Generator(t *testing.T) {
	cli, _ := newK8sClient()
	type fields struct {
		Subject     []string
		CertDir     string
		Namespace   string
		ServiceName string
		SecretName  string
		CsrName     string
		WebHook     []WebHook
		client      *kubernetes.Clientset
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{name: "test", fields: fields{
			Subject:     []string{"www.cuisongliu.com"},
			CertDir:     "/Users/cuisongliu",
			Namespace:   "default",
			ServiceName: "service",
			SecretName:  "webhook-cert",
			CsrName:     "webhook-csr",
			WebHook: []WebHook{
				{
					ValidatingName: "validating-cfg",
				},
				{
					MutatingName: "mutating-cfg",
				},
			},
			client: cli,
		}, wantErr: false},
	}
	t.Log("before delete secrets and csr")
	_ = cli.CoreV1().Secrets("default").Delete("webhook-cert", &v1.DeleteOptions{})
	_ = cli.CertificatesV1beta1().CertificateSigningRequests().Delete("webhook-csr", &v1.DeleteOptions{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &CertWebHook{
				Subject:     tt.fields.Subject,
				CertDir:     tt.fields.CertDir,
				Namespace:   tt.fields.Namespace,
				ServiceName: tt.fields.ServiceName,
				SecretName:  tt.fields.SecretName,
				CsrName:     tt.fields.CsrName,
				WebHook:     tt.fields.WebHook,
				client:      tt.fields.client,
			}
			if err := c.Generator(); (err != nil) != tt.wantErr {
				t.Errorf("Generator() error = %v, wantErr %v", err, tt.wantErr)
			}

		})
	}
	t.Log("after delete secrets and csr")
	_ = cli.CoreV1().Secrets("default").Delete("webhook-cert", &v1.DeleteOptions{})
	_ = cli.CertificatesV1beta1().CertificateSigningRequests().Delete("webhook-csr", &v1.DeleteOptions{})
}
