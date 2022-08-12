package k8s

import (
	"context"
	logconfigv1beta1 "github.com/loggie-io/loggie/pkg/discovery/kubernetes/apis/loggie/v1beta1"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sync"

	kubeclientset "k8s.io/client-go/kubernetes"
)

var (
	Mgr *ClusterManager

	scheme = runtime.NewScheme()

	once sync.Once
)

type ClusterManager struct {
	manager.Manager                          // from controller-runtime
	Kubeclient      *kubeclientset.Clientset // from client-go
}

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(logconfigv1beta1.AddToScheme(scheme))

	//+kubebuilder:scaffold:scheme
}

func InitCluster() {
	once.Do(func() {
		cfg := ctrl.GetConfigOrDie()
		mgr, err := ctrl.NewManager(cfg, ctrl.Options{
			Scheme:                 scheme,
			LeaderElection:         false,
			MetricsBindAddress:     "0",
			HealthProbeBindAddress: "0",
		})
		Expect(err).ShouldNot(HaveOccurred())

		go func() {
			if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
				Expect(err).ShouldNot(HaveOccurred())
			}
		}()

		mgr.GetCache().WaitForCacheSync(context.Background())

		// make kube clientset
		kubeClient, err := kubeclientset.NewForConfig(cfg)
		Expect(err).ShouldNot(HaveOccurred())

		Mgr = &ClusterManager{
			Manager:    mgr,
			Kubeclient: kubeClient,
		}
	})
}
