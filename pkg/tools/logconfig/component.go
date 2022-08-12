package logconfig

import (
	"fmt"
	essink "github.com/loggie-io/loggie/pkg/sink/elasticsearch"
	grpcsink "github.com/loggie-io/loggie/pkg/sink/grpc"
	filesource "github.com/loggie-io/loggie/pkg/source/file"
	. "github.com/onsi/gomega"
	"gopkg.in/yaml.v2"
	"loggie-test/pkg/resources/elasticsearch"
	"loggie-test/pkg/resources/genfilesdeploy"
	"loggie-test/pkg/resources/k8s_loggie_aggre"
	"path"
)

func makeFileSource(deploy *genfilesdeploy.GenFilesDeployment, options *filesource.Config) string {
	filesrc := options
	if filesrc == nil {
		filesrc = &filesource.Config{}
	}

	filesrc.CollectConfig.Paths = []string{path.Join(deploy.Conf.Dir, "*.log")}
	src := []fileSource{
		{
			NameType: NameType{
				Name: "common",
				Type: filesource.Type,
			},
			Config: *filesrc,
		},
	}
	srcout, err := yaml.Marshal(&src)
	Expect(err).ShouldNot(HaveOccurred())

	return string(srcout)
}

func makeESSink(esIns *elasticsearch.ES) string {
	essk := essink.Config{
		Hosts: []string{fmt.Sprintf("%s.%s.svc:9200", esIns.Conf.Service, esIns.Conf.Namespace)},
		Index: esIns.Conf.Index,
	}
	sk := elasticsearchSink{
		NameType: NameType{
			Type: elasticsearch.Name,
		},
		Config: essk,
	}
	skout, err := yaml.Marshal(&sk)
	Expect(err).ShouldNot(HaveOccurred())

	return string(skout)
}

func makeGrpcSink(loggieIns *k8s_loggie_aggre.Loggie) string {
	gcsk := grpcsink.Config{
		Host: "loggie-aggregator." + loggieIns.Conf.Namespace + ".svc:6066",
	}
	sk := struct {
		NameType `yaml:",inline"`
		grpcsink.Config `yaml:",inline"`
	}{
		NameType: NameType{
			Type: grpcsink.Type,
		},
		Config: gcsk,
	}

	out, err := yaml.Marshal(&sk)
	Expect(err).ShouldNot(HaveOccurred())

	return string(out)
}

func makeGrpcSource() string {
	src := []struct {
		NameType `yaml:",inline"`
	}{
		{
			NameType{
				Name: "g",
				Type: "grpc",
			},
		},
	}
	out, err := yaml.Marshal(&src)
	Expect(err).ShouldNot(HaveOccurred())
	return string(out)
}

