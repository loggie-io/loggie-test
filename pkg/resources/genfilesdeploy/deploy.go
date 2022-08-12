package genfilesdeploy

import (
	"context"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"loggie-test/pkg/resources"
	"loggie-test/pkg/resources/genfiles"
	"loggie-test/pkg/tools/k8s"
)

const (
	Name = "genfilesDeployment"
)

var Label = map[string]string{
	"loggie-test": "genfiles",
}

func init() {
	resources.Register(Name, makeGenFilesDeployment)
}

type Config struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
	Replicas  *int32 `yaml:"replicas"`
	Image     string `yaml:"image"`

	genfiles.Config `yaml:",inline"`
}

var _ resources.Resource = (*GenFilesDeployment)(nil)

type GenFilesDeployment struct {
	Conf *Config
}

func makeGenFilesDeployment() interface{} {
	return &GenFilesDeployment{
		Conf: &Config{},
	}
}

func (r *GenFilesDeployment) Config() interface{} {
	return r.Conf
}

func (r *GenFilesDeployment) Name() string {
	return Name
}

func (r *GenFilesDeployment) Setup(ctx context.Context) error {

	// create deployment for log collection
	k8s.InitCluster()

	out, err := yaml.Marshal(r.Conf.Config)
	if err != nil {
		return err
	}

	cmName := r.Conf.Name + "-config"
	var cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: r.Conf.Namespace,
		},
		Data: map[string]string{
			"config.yml": string(out),
		},
	}

	err = k8s.Mgr.GetClient().Create(ctx, cm)
	if err != nil {
		return err
	}

	var deployment = &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Conf.Name,
			Namespace: r.Conf.Namespace,
			Labels:    Label,
		},
		Spec: v1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: Label,
			},
			Replicas: r.Conf.Replicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: Label,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "genfiles",
							Image: r.Conf.Image,
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "log",
									MountPath: r.Conf.Dir,
								},
								{
									Name:      "config",
									MountPath: "/config.yml",
									SubPath:   "config.yml",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "log",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: cmName,
									},
								},
							},
						},
					},
				},
			},
		},
	}
	return k8s.Mgr.GetClient().Create(ctx, deployment)
}

func (r *GenFilesDeployment) CleanUp(ctx context.Context) error {
	// delete cm
	var cm = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Conf.Name + "-config",
			Namespace: r.Conf.Namespace,
		},
	}
	err := k8s.Mgr.GetClient().Delete(context.Background(), cm)
	if err != nil {
		return err
	}

	// delete deployment
	deployment := &v1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      r.Conf.Name,
			Namespace: r.Conf.Namespace,
		},
	}

	return k8s.Mgr.GetClient().Delete(context.Background(), deployment)
}

func (r *GenFilesDeployment) Ready() (bool, error) {
	// check if deployment is running
	deployment := &v1.Deployment{}
	err := k8s.Mgr.GetClient().Get(context.Background(), types.NamespacedName{Namespace: r.Conf.Namespace, Name: r.Conf.Name}, deployment)
	if err != nil {
		return false, err
	}

	if deployment.Status.Replicas == *r.Conf.Replicas {
		return true, nil
	}

	return false, nil
}

func (r *GenFilesDeployment) AllCount() int64 {
	return int64(r.Conf.FileCount * r.Conf.LineCount)
}

func (r *GenFilesDeployment) GetDeployment() (*v1.Deployment, error) {
	deployment := &v1.Deployment{}
	if err := k8s.Mgr.GetClient().Get(context.Background(), types.NamespacedName{
		Name:      r.Conf.Name,
		Namespace: r.Conf.Namespace,
	}, deployment); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *GenFilesDeployment) UpdateDeployment(d *v1.Deployment) error {
	if err := k8s.Mgr.GetClient().Update(context.Background(), d); err != nil {
		return err
	}
	return nil
}
