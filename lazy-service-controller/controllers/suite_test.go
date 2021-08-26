/*
Copyright 2021.

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

package controllers

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	deployv1 "lazy-service-controller/api/v1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = deployv1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

}, 60)

var _ = Describe("Webhook server test", func() {

	var (
		ctx context.Context
		//ctxCancel    context.CancelFunc
		testHostPort string
		client       *http.Client
		server       *webhook.Server
		servingOpts  envtest.WebhookInstallOptions
	)

	BeforeEach(func() {
		//ctx, calcelFunc = context.WithCancel(context.Background())

		// closed in indivual tests differently

		servingOpts = envtest.WebhookInstallOptions{}
		Expect(servingOpts.PrepWithoutInstalling()).To(Succeed())

		testHostPort = net.JoinHostPort(servingOpts.LocalServingHost, fmt.Sprintf("%d", servingOpts.LocalServingPort))

		// bypass needing to set up the x509 cert pool, etc ourselves
		clientTransport, err := rest.TransportFor(&rest.Config{
			TLSClientConfig: rest.TLSClientConfig{CAData: servingOpts.LocalServingCAData},
		})
		Expect(err).NotTo(HaveOccurred())
		client = &http.Client{
			Transport: clientTransport,
		}

		server = &webhook.Server{
			Host:    servingOpts.LocalServingHost,
			Port:    servingOpts.LocalServingPort,
			CertDir: servingOpts.LocalServingCertDir,
		}
	})
	AfterEach(func() {
		Expect(servingOpts.Cleanup()).To(Succeed())
	})

	startServer := func() (done <-chan struct{}) {
		doneCh := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer close(doneCh)
			Expect(server.Start(ctx)).To(Succeed())
		}()
		// wait till we can ping the server to start the test
		Eventually(func() error {
			_, err := client.Get(fmt.Sprintf("https://%s/unservedpath", testHostPort))
			return err
		}).Should(Succeed())

		// this is normally called before Start by the manager
		Expect(server.InjectFunc(func(i interface{}) error {
			boolInj, canInj := i.(interface{ InjectBool(bool) error })
			if !canInj {
				return nil
			}
			return boolInj.InjectBool(true)
		})).To(Succeed())

		return doneCh
	}

	go startServer()

	Context("should panic if duplicate path register ", func() {
		server.Register("/somepath", &testHandler{})

	})

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

type testHandler struct {
	injectedField bool
}

func (t *testHandler) InjectBool(val bool) error {
	t.injectedField = val
	return nil
}
func (t *testHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	if _, err := resp.Write([]byte("gadzooks!")); err != nil {
		panic("unable to write http response!")
	}
}
