package controller

import (
	"fmt"
	seatunnelv1 "github.com/nineinfra/seatunnel-operator/api/v1"
	corev1 "k8s.io/api/core/v1"
	"strconv"
	"strings"
)

func NineResourceName(job *seatunnelv1.SeatunnelJob, suffixs ...string) string {
	return job.Name + strings.Join(suffixs, "-")
}

func ClusterResourceName(job *seatunnelv1.SeatunnelJob, suffixs ...string) string {
	return job.Name + DefaultNameSuffix + strings.Join(suffixs, "-")
}

func ClusterResourceLabels(job *seatunnelv1.SeatunnelJob) map[string]string {
	return map[string]string{
		"cluster": job.Name,
		"app":     DefaultClusterSign,
	}
}

func GetStorageClassName(job *seatunnelv1.SeatunnelJob) string {
	if job.Spec.Resource.StorageClass != "" {
		return job.Spec.Resource.StorageClass
	}
	return DefaultStorageClass
}

func GetFullSvcName(job *seatunnelv1.SeatunnelJob) string {
	return fmt.Sprintf("%s.%s.svc.%s", ClusterResourceName(job), job.Namespace, GetClusterDomain(job))
}

func GetClusterDomain(job *seatunnelv1.SeatunnelJob) string {
	if job.Spec.K8sConf != nil {
		if value, ok := job.Spec.K8sConf[DefaultClusterDomainName]; ok {
			return value
		}
	}
	return DefaultClusterDomain
}

func GetSeatunnelEngine(job *seatunnelv1.SeatunnelJob) string {
	if job.Spec.Conf.Env != nil {
		if value, ok := job.Spec.Conf.Env[seatunnelv1.NINEINFRA_ENV_SEATUNNEL_ENGINE]; ok {
			if value == seatunnelv1.NINEINFRA_ENV_SEATUNNEL_ENGINE_FLINK || value == seatunnelv1.NINEINFRA_ENV_SEATUNNEL_ENGINE_SPARK {
				return value
			}
		}
	}
	return seatunnelv1.NINEINFRA_ENV_SEATUNNEL_ENGINE_DEFAULT
}

func GetSeatunnelDebugMode(job *seatunnelv1.SeatunnelJob) string {
	if job.Spec.Conf.Env != nil {
		if value, ok := job.Spec.Conf.Env[seatunnelv1.NINEINFRA_ENV_SEATUNNEL_DEBUG_MODE]; ok {
			if value == seatunnelv1.NINEINFRA_ENV_SEATUNNEL_DEBUG_MODE_ON ||
				value == seatunnelv1.NINEINFRA_ENV_SEATUNNEL_DEBUG_MODE_OFF {
				return value
			}
		}
	}
	return seatunnelv1.NINEINFRA_ENV_SEATUNNEL_DEBUG_MODE_DEFAULT
}

func GetSeatunnelDebugTime(job *seatunnelv1.SeatunnelJob) int {
	if job.Spec.Conf.Env != nil {
		if value, ok := job.Spec.Conf.Env[seatunnelv1.NINEINFRA_ENV_SEATUNNEL_DEBUG_TIME]; ok {
			atoi, err := strconv.Atoi(value)
			if err == nil {
				return atoi
			}
		}
	}
	return seatunnelv1.NINEINFRA_ENV_SEATUNNEL_DEBUG_TIME_DEFAULT
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
