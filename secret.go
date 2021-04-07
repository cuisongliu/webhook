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
	"io/ioutil"
	"k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"os"
	"path"
	"time"
)

const (
	certKey     = "tls.crt"
	keyKey      = "tls.key"
	csrKey      = "tls.csr"
	caBundleKey = "caBundle"
)

func (c *CertWebHook) generateSecret() (*corev1.Secret, error) {
	secret, err := c.client.CoreV1().Secrets(c.Namespace).Get(c.SecretName, v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			return nil, err
		}
		csr, key, err := c.generateTLS()
		if err != nil {
			return nil, err
		}

		secret = &corev1.Secret{
			ObjectMeta: v1.ObjectMeta{
				Namespace: c.Namespace,
				Name:      c.SecretName,
			},
			Data: map[string][]byte{
				csrKey: csr,
				keyKey: key,
			},
		}
		secret, err = c.client.CoreV1().Secrets(c.Namespace).Create(secret)
		if err != nil {
			return nil, err
		}
	}
	//csr
	err = c.pathCsr(secret)
	if err != nil {
		return nil, err
	}
	//ca
	caConfigMap, err := c.client.CoreV1().ConfigMaps("kube-system").Get("extension-apiserver-authentication", v1.GetOptions{})
	if err != nil {
		return nil, err
	}
	var caData string
	if caConfigMap != nil {
		caData = caConfigMap.Data["client-ca-file"]
	} else {
		return nil, errors.NewUnauthorized("ca configmap [extension-apiserver-authentication] data [client-ca-file] is not found.")
	}
	secret.Data[caBundleKey] = []byte(caData)
	secret, err = c.client.CoreV1().Secrets(c.Namespace).Update(secret)
	if err != nil {
		return nil, err
	}
	return secret, nil
}
func (c *CertWebHook) pathCsr(secret *corev1.Secret) error {
	dPolicy := v1.DeletePropagationBackground
	label := map[string]string{
		"csr-name": c.CsrName,
	}
	_ = c.client.CertificatesV1beta1().CertificateSigningRequests().Delete(c.CsrName, &v1.DeleteOptions{PropagationPolicy: &dPolicy})
	csrResource := &v1beta1.CertificateSigningRequest{}
	csrResource.Name = c.CsrName
	csrResource.Labels = label
	csrResource.Spec.Groups = []string{"system:authenticated"}
	csrResource.Spec.Usages = []v1beta1.KeyUsage{
		"digital signature",
		"key encipherment",
		"server auth",
	}
	csrResource.Spec.Request = secret.Data[csrKey]
	csrResource, err := c.client.CertificatesV1beta1().CertificateSigningRequests().Create(csrResource)

	if err != nil {
		return err
	}
	csrResource.Status.Conditions = []v1beta1.CertificateSigningRequestCondition{
		{Type: v1beta1.CertificateApproved, Reason: "PodSelfApprove", Message: "This CSR was approved by pod certificate approve.", LastUpdateTime: v1.NewTime(time.Now())},
	}
	csrResource, err = c.client.CertificatesV1beta1().CertificateSigningRequests().UpdateApproval(csrResource)
	if err != nil {
		return err
	}
	w, err := c.client.CertificatesV1beta1().CertificateSigningRequests().Watch(v1.ListOptions{LabelSelector: "csr-name=" + c.CsrName})
	if err != nil {
		return err
	}
	for {
		select {
		case <-time.After(time.Second * 10):
			return errors.NewBadRequest("The CSR is not ready.")
		case event := <-w.ResultChan():
			if event.Type == watch.Modified || event.Type == watch.Added {
				csr := event.Object.(*v1beta1.CertificateSigningRequest)
				if csr.Status.Certificate != nil {
					secret.Data[certKey] = csr.Status.Certificate
					return nil
				}
			}
		}
	}
}
func (c *CertWebHook) patchWebHook(caBundle string) error {
	for _, wk := range c.WebHook {

		if wk.ValidatingName != "" {
			vwebhook, err := c.client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Get(wk.ValidatingName, v1.GetOptions{})
			if err != nil {
				return err
			}
			for i := range vwebhook.Webhooks {
				vwebhook.Webhooks[i].ClientConfig.Service.Name = c.ServiceName
				vwebhook.Webhooks[i].ClientConfig.Service.Namespace = c.Namespace

				vwebhook.Webhooks[i].ClientConfig.CABundle = []byte(caBundle)
				if wk.NamespaceSelect != nil {
					if v, ok := wk.NamespaceSelect[vwebhook.Webhooks[i].Name]; ok {
						vwebhook.Webhooks[i].NamespaceSelector = v
					}
				}
				if wk.ObjectSelect != nil {
					if v, ok := wk.ObjectSelect[vwebhook.Webhooks[i].Name]; ok {
						vwebhook.Webhooks[i].ObjectSelector = v
					}
				}
			}
			_, err = c.client.AdmissionregistrationV1().ValidatingWebhookConfigurations().Update(vwebhook)
			if err != nil {
				return err
			}
		}

		if wk.MutatingName != "" {
			mwebhook, err := c.client.AdmissionregistrationV1().MutatingWebhookConfigurations().Get(wk.MutatingName, v1.GetOptions{})
			if err != nil {
				return err
			}
			for i := range mwebhook.Webhooks {
				mwebhook.Webhooks[i].ClientConfig.Service.Name = c.ServiceName
				mwebhook.Webhooks[i].ClientConfig.Service.Namespace = c.Namespace

				mwebhook.Webhooks[i].ClientConfig.CABundle = []byte(caBundle)
				if wk.NamespaceSelect != nil {
					if v, ok := wk.NamespaceSelect[mwebhook.Webhooks[i].Name]; ok {
						mwebhook.Webhooks[i].NamespaceSelector = v
					}
				}
				if wk.ObjectSelect != nil {
					if v, ok := wk.ObjectSelect[mwebhook.Webhooks[i].Name]; ok {
						mwebhook.Webhooks[i].ObjectSelector = v
					}
				}
			}
			_, err = c.client.AdmissionregistrationV1().MutatingWebhookConfigurations().Update(mwebhook)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *CertWebHook) writeTLSFiles(certData []byte, keyData []byte) error {
	if _, err := os.Stat(c.CertDir); os.IsNotExist(err) {
		if err := os.MkdirAll(c.CertDir, 0700); err != nil {
			return err
		}
	}
	if err := ioutil.WriteFile(path.Join(c.CertDir, "tls.crt"), certData, 0600); err != nil {
		return err
	}
	if err := ioutil.WriteFile(path.Join(c.CertDir, "tls.key"), keyData, 0600); err != nil {
		return err
	}
	return nil
}
