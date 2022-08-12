package logconfig

import (
	"context"
	logconfigv1beta1 "github.com/loggie-io/loggie/pkg/discovery/kubernetes/apis/loggie/v1beta1"
	essink "github.com/loggie-io/loggie/pkg/sink/elasticsearch"
	filesource "github.com/loggie-io/loggie/pkg/source/file"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/genfilesdeploy"
	"loggie-test/pkg/resources/k8s_loggie_aggre"
	"loggie-test/pkg/tools/k8s"
)

type NameType struct {
	Name              string `yaml:"name,omitempty"`
	Type              string `yaml:"type"`
}

type fileSource struct {
	NameType `yaml:",inline"`
	filesource.Config `yaml:",inline"`
}

type elasticsearchSink struct {
	NameType `yaml:",inline"`
	essink.Config `yaml:",inline"`
}

func GenLogConfigWithESSink(esIns *elasticsearch.ES, deploy *genfilesdeploy.GenFilesDeployment, filesourceOpts *filesource.Config, ) *logconfigv1beta1.LogConfig {
	srcout := makeFileSource(deploy, filesourceOpts)
	skout := makeESSink(esIns)

	return &logconfigv1beta1.LogConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "logconfig-" + deploy.Conf.Name,
			Namespace: deploy.Conf.Namespace,
			Labels:    genfilesdeploy.Label,
		},
		Spec: logconfigv1beta1.Spec{
			Selector: &logconfigv1beta1.Selector{
				Type: logconfigv1beta1.SelectorTypePod,
				PodSelector: logconfigv1beta1.PodSelector{
					LabelSelector: genfilesdeploy.Label,
				},
			},
			Pipeline: &logconfigv1beta1.Pipeline{
				Sources: srcout,
				Sink:    skout,
			},
		},
	}
}


func GenLogConfigWithGrpcSink(lgc *filesource.Config, loggieIns *k8s_loggie_aggre.Loggie, deploy *genfilesdeploy.GenFilesDeployment) *logconfigv1beta1.LogConfig {
	srcout := makeFileSource(deploy, lgc)
	gsink := makeGrpcSink(loggieIns)

	return &logconfigv1beta1.LogConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "logconfig-" + deploy.Conf.Name,
			Namespace: deploy.Conf.Namespace,
			Labels:    genfilesdeploy.Label,
		},
		Spec: logconfigv1beta1.Spec{
			Selector: &logconfigv1beta1.Selector{
				Type: logconfigv1beta1.SelectorTypePod,
				PodSelector: logconfigv1beta1.PodSelector{
					LabelSelector: genfilesdeploy.Label,
				},
			},
			Pipeline: &logconfigv1beta1.Pipeline{
				Sources: srcout,
				Sink:    gsink,
			},
		},
	}
}

func GenClusterLogConfigWithGrpcSourceAndESSink(esIns *elasticsearch.ES) *logconfigv1beta1.ClusterLogConfig {
	gsrc := makeGrpcSource()
	esink := makeESSink(esIns)

	return &logconfigv1beta1.ClusterLogConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "aggregator",
		},
		Spec: logconfigv1beta1.Spec{
			Selector: &logconfigv1beta1.Selector{
				Type: logconfigv1beta1.SelectorTypeCluster,
				Cluster: "aggregator",
			},
			Pipeline: &logconfigv1beta1.Pipeline{
				Sources: gsrc,
				Sink:    esink,
			},
		},
	}
}

func CreateLogConfig(lgc *logconfigv1beta1.LogConfig) error {
	return k8s.Mgr.GetClient().Create(context.Background(), lgc)
}

func CreateClusterLogConfig(clgc *logconfigv1beta1.ClusterLogConfig) error {
	return k8s.Mgr.GetClient().Create(context.Background(), clgc)
}

func DeleteClusterLogConfig(clgc *logconfigv1beta1.ClusterLogConfig) error {
	return k8s.Mgr.GetClient().Delete(context.Background(), clgc)
}

func DeleteLogConfig(lgc *logconfigv1beta1.LogConfig) error {
	return k8s.Mgr.GetClient().Delete(context.Background(), lgc)
}
