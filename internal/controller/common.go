package controller

import (
	"fmt"
	seatunnelv1 "github.com/nineinfra/seatunnel-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"strings"
)

func ClusterResourceName(cluster *seatunnelv1.SeatunnelJob, suffixs ...string) string {
	return cluster.Name + DefaultNameSuffix + strings.Join(suffixs, "-")
}

func ClusterResourceLabels(cluster *seatunnelv1.SeatunnelJob) map[string]string {
	return map[string]string{
		"cluster": cluster.Name,
		"app":     DefaultClusterSign,
	}
}

func GetStorageClassName(cluster *seatunnelv1.SeatunnelJob) string {
	if cluster.Spec.Resource.StorageClass != "" {
		return cluster.Spec.Resource.StorageClass
	}
	return DefaultStorageClass
}

func GetFullSvcName(cluster *seatunnelv1.SeatunnelJob) string {
	return fmt.Sprintf("%s.%s.svc.%s", ClusterResourceName(cluster), cluster.Namespace, GetClusterDomain(cluster))
}

func GetClusterDomain(cluster *seatunnelv1.SeatunnelJob) string {
	if cluster.Spec.K8sConf != nil {
		if value, ok := cluster.Spec.K8sConf[DefaultClusterDomainName]; ok {
			return value
		}
	}
	return DefaultClusterDomain
}

func ConstructClusterRefsVolumeMounts(cluster *seatunnelv1.SeatunnelJob) []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      ClusterResourceName(cluster, DefaultClusterRefsConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, DefaultSparkConfFile),
			SubPath:   DefaultSparkConfFile,
		},
		{
			Name:      ClusterResourceName(cluster, DefaultClusterRefsConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, DefaultHiveSiteFile),
			SubPath:   DefaultHiveSiteFile,
		},
		{
			Name:      ClusterResourceName(cluster, DefaultClusterRefsConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, DefaultHdfsSiteFile),
			SubPath:   DefaultHdfsSiteFile,
		},
		{
			Name:      ClusterResourceName(cluster, DefaultClusterRefsConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, DefaultCoreSiteFile),
			SubPath:   DefaultCoreSiteFile,
		},
	}
}

func ConstructClusterRefsVolumes(cluster *seatunnelv1.SeatunnelJob) []corev1.Volume {
	return []corev1.Volume{
		{
			Name: ClusterResourceName(cluster, DefaultClusterRefsConfigNameSuffix),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: fmt.Sprintf("%s%s", cluster.Name, DefaultClusterRefsConfigNameSuffix),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  DefaultSparkConfFile,
							Path: DefaultSparkConfFile,
						},
						{
							Key:  DefaultHiveSiteFile,
							Path: DefaultHiveSiteFile,
						},
						{
							Key:  DefaultHdfsSiteFile,
							Path: DefaultHdfsSiteFile,
						},
						{
							Key:  DefaultCoreSiteFile,
							Path: DefaultCoreSiteFile,
						},
					},
				},
			},
		},
	}
}

func DefaultDownwardAPI() []corev1.EnvVar {
	return []corev1.EnvVar{
		{
			Name: "POD_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.podIP",
				},
			},
		},
		{
			Name: "POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.name",
				},
			},
		},
		{
			Name: "NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.namespace",
				},
			},
		},
		{
			Name: "POD_UID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "metadata.uid",
				},
			},
		},
		{
			Name: "HOST_IP",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: "status.hostIP",
				},
			},
		},
	}
}

func DefaultEnvs() []corev1.EnvVar {
	envs := DefaultDownwardAPI()
	return envs
}
