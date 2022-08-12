package k8s

import (
	"context"
	"fmt"
	"github.com/onsi/ginkgo/v2"
	"io"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
	"net/http"
	"os"
)

func PortForwardByService(namespace, svcName string, listenPort int) error {
	// get svc label select pod
	svc, err := Mgr.Kubeclient.CoreV1().Services(namespace).Get(context.Background(), svcName, v1.GetOptions{})
	if err != nil {
		return err
	}

	podList, err := Mgr.Kubeclient.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{LabelSelector: labels.FormatLabels(svc.Spec.Selector)})
	if err != nil {
		return err
	}
	if len(podList.Items) == 0 {
		return fmt.Errorf("select pods is null by service labelSelector: %s", labels.FormatLabels(svc.Spec.Selector))
	}
	podName := podList.Items[0]
	return PortForward(namespace, podName.Name, listenPort)
}

func PortForward(namespace, podName string, listenPort int) error {
	ginkgo.GinkgoWriter.Printf("port-forward start listen :%d\n", listenPort)
	config := Mgr.GetConfig()

	StopChannel := make(chan struct{}, 1)
	ReadyChannel := make(chan struct{})

	req := Mgr.Kubeclient.CoreV1().RESTClient().Post().Namespace(namespace).
		Resource("pods").Name(podName).SubResource("portforward")

	transport, upgrader, err := spdy.RoundTripperFor(config)
	if err != nil {
		return err
	}
	address := []string{"127.0.0.1"}
	ports := []string{fmt.Sprintf("%d:%d", listenPort, listenPort)}

	IOStreams := struct {
		// In think, os.Stdin
		In io.Reader
		// Out think, os.Stdout
		Out io.Writer
		// ErrOut think, os.Stderr
		ErrOut io.Writer
	}{os.Stdin, os.Stdout, os.Stderr}

	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, http.MethodPost, req.URL())
	fw, err := portforward.NewOnAddresses(dialer, address, ports, StopChannel, ReadyChannel, IOStreams.Out, IOStreams.ErrOut)
	if err != nil {
		return err
	}
	err = fw.ForwardPorts()
	if err != nil {
		return err
	}
	ginkgo.GinkgoWriter.Printf("port-forward stop listen :%d\n", listenPort)
	return nil
}

