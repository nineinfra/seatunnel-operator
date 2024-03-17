package controller

import (
	"context"
	"errors"
	"fmt"
	seatunnelv1 "github.com/nineinfra/seatunnel-operator/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"strings"
)

func getReplicas(cluster *seatunnelv1.SeatunnelJob) int32 {
	if cluster.Spec.Resource.Replicas != 0 && cluster.Spec.Resource.Replicas%2 != 0 {
		return cluster.Spec.Resource.Replicas
	}
	return DefaultReplicas
}

func checkSourceTypeSupported(srcType string) bool {
	return strings.Contains(SourceTypeSupportedList, srcType)
}

func checkTransformTypeSupported(transformType string) bool {
	return strings.Contains(TransformTypeSupportedList, transformType)
}

func checkSinkTypeSupported(sinkType string) bool {
	return strings.Contains(SinkTypeSupportedList, sinkType)
}

func writeEnvConfig(conf map[string]string) string {
	var sb strings.Builder
	sb.WriteString("env{\n")
	sb.WriteString(map2String(conf))
	sb.WriteString("}\n")
	return sb.String()
}

func writeSourceConfig(source *seatunnelv1.SourceConfig) string {
	var sb strings.Builder
	sb.WriteString("source{\n")
	if !checkSourceTypeSupported(source.Type) {
		return ""
	}
	sb.WriteString(fmt.Sprintf("%s {\n", source.Type))
	if source.Conf != nil {
		sb.WriteString(map2String(source.Conf))
	}
	if source.TableList != nil {
		sb.WriteString(fmt.Sprintf("%s = [\n", "table_list"))
		for _, v := range source.TableList {
			sb.WriteString(map2String(v))
		}
		sb.WriteString("]\n")
	}
	if source.Scheme.Fields != nil {
		sb.WriteString("schema = {\n")
		sb.WriteString("fields {\n")
		sb.WriteString(map2String(source.Scheme.Fields))
		sb.WriteString("}\n")
		sb.WriteString("}\n")
	}
	if source.Properties != nil {
		sb.WriteString("properties {\n")
		sb.WriteString(map2String(source.Properties))
		sb.WriteString("}\n")
	}
	if source.ExtraConfig != nil {
		sb.WriteString(fmt.Sprintf("%s.config = {", strings.ToLower(source.Type)))
		sb.WriteString(map2String(source.ExtraConfig))
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n")
	sb.WriteString("}\n")
	return sb.String()
}

func writeTransformConfig(transform *seatunnelv1.TransformConfig) string {
	var sb strings.Builder
	sb.WriteString("transform{\n")
	if !checkTransformTypeSupported(transform.Type) {
		return ""
	}
	sb.WriteString(fmt.Sprintf("%s {\n", transform.Type))
	sb.WriteString(map2String(transform.Conf))
	if transform.Scheme.Fields != nil {
		sb.WriteString("schema = {\n")
		sb.WriteString("fields {\n")
		sb.WriteString(map2String(transform.Scheme.Fields))
		sb.WriteString("}\n")
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n")
	sb.WriteString("}\n")
	return sb.String()
}

func writeSinkConfig(sink *seatunnelv1.SinkConfig) string {
	var sb strings.Builder
	sb.WriteString("sink{\n")
	if !checkSinkTypeSupported(sink.Type) {
		return ""
	}
	sb.WriteString(fmt.Sprintf("%s {\n", sink.Type))
	if sink.Conf != nil {
		sb.WriteString(map2String(sink.Conf))
	}
	if sink.PartitionBy != nil {
		sb.WriteString("partition_by = [")
		sb.WriteString(list2String(sink.PartitionBy))
		sb.WriteString("]\n")
	}
	if sink.SinkColumns != nil {
		sb.WriteString("sink_columns = [")
		sb.WriteString(list2String(sink.SinkColumns))
		sb.WriteString("]\n")
	}
	if sink.ExtraConfig != nil {
		sb.WriteString(fmt.Sprintf("%s.config = {", strings.ToLower(sink.Type)))
		sb.WriteString(map2String(sink.ExtraConfig))
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n")
	sb.WriteString("}\n")
	return sb.String()
}

func constructJobConfig(job *seatunnelv1.SeatunnelJob) string {
	envString := ""
	if job.Spec.Conf.Env != nil {
		envString = writeEnvConfig(job.Spec.Conf.Env)
	}
	sourceString := writeSourceConfig(&job.Spec.Conf.Source)
	transformString := writeTransformConfig(&job.Spec.Conf.Transform)
	sinkString := writeSinkConfig(&job.Spec.Conf.Sink)

	return fmt.Sprintf("%s%s%s%s", envString, sourceString, transformString, sinkString)
}

func constructLogConfig() string {
	tmpConf := DefaultLogConfKeyValue
	return map2String(tmpConf)
}

func getImageConfig(cluster *seatunnelv1.SeatunnelJob) seatunnelv1.ImageConfig {
	ic := seatunnelv1.ImageConfig{
		Repository:  cluster.Spec.Image.Repository,
		PullSecrets: cluster.Spec.Image.PullSecrets,
	}
	ic.Tag = cluster.Spec.Image.Tag
	if ic.Tag == "" {
		ic.Tag = cluster.Spec.Version
	}
	ic.PullPolicy = cluster.Spec.Image.PullPolicy
	if ic.PullPolicy == "" {
		ic.PullPolicy = string(corev1.PullIfNotPresent)
	}
	return ic
}

func (r *SeatunnelJobReconciler) constructConfigMap(job *seatunnelv1.SeatunnelJob) (*corev1.ConfigMap, error) {
	conf := constructJobConfig(job)
	if conf == "" {
		return nil, errors.New("invalid parameters")
	}
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(job, DefaultConfigNameSuffix),
			Namespace: job.Namespace,
			Labels:    ClusterResourceLabels(job),
		},
		Data: map[string]string{
			DefaultConfigFileName:    constructJobConfig(job),
			DefaultLogConfigFileName: constructLogConfig(),
		},
	}
	if err := ctrl.SetControllerReference(job, cm, r.Scheme); err != nil {
		return cm, err
	}
	return cm, nil
}

func (r *SeatunnelJobReconciler) constructVolumeMounts(cluster *seatunnelv1.SeatunnelJob) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      ClusterResourceName(cluster, DefaultConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/%s", DefaultConfPath, DefaultConfigFileName),
			SubPath:   DefaultConfigFileName,
		},
		{
			Name:      ClusterResourceName(cluster, DefaultConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/%s", DefaultConfPath, DefaultLogConfigFileName),
			SubPath:   DefaultLogConfigFileName,
		},
	}
	return volumeMounts
}

func (r *SeatunnelJobReconciler) constructVolumes(cluster *seatunnelv1.SeatunnelJob) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: ClusterResourceName(cluster, DefaultConfigNameSuffix),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ClusterResourceName(cluster, DefaultConfigNameSuffix),
					},
					Items: []corev1.KeyToPath{
						{
							Key:  DefaultConfigFileName,
							Path: DefaultConfigFileName,
						},
						{
							Key:  DefaultLogConfigFileName,
							Path: DefaultLogConfigFileName,
						},
					},
				},
			},
		},
	}

	return volumes
}

func (r *SeatunnelJobReconciler) constructPodSpec(job *seatunnelv1.SeatunnelJob) (*corev1.PodSpec, error) {
	tgp := int64(DefaultTerminationGracePeriod)
	ic := getImageConfig(job)
	var tmpPullSecrets []corev1.LocalObjectReference
	if ic.PullSecrets != "" {
		tmpPullSecrets = make([]corev1.LocalObjectReference, 0)
		tmpPullSecrets = append(tmpPullSecrets, corev1.LocalObjectReference{Name: ic.PullSecrets})
	}
	apiSvc := &corev1.Service{}
	err := r.Client.Get(context.TODO(), types.NamespacedName{Name: "kubernetes", Namespace: "default"}, apiSvc)
	if err != nil {
		return nil, err
	}
	cmdString := []string{
		"/opt/seatunnel/bin/start-seatunnel-spark-3-connector-v2.sh",
		"--master",
		fmt.Sprintf("k8s://https://%s:443", apiSvc.Spec.ClusterIP),
		"--config",
		DefaultConfFile,
	}
	return &corev1.PodSpec{
		Containers: []corev1.Container{
			{
				Name:            job.Name,
				Image:           ic.Repository + ":" + ic.Tag,
				ImagePullPolicy: corev1.PullPolicy(ic.PullPolicy),
				Command:         cmdString,
				Env:             DefaultEnvs(),
				VolumeMounts:    r.constructVolumeMounts(job),
			},
		},
		ImagePullSecrets:              tmpPullSecrets,
		RestartPolicy:                 corev1.RestartPolicyAlways,
		TerminationGracePeriodSeconds: &tgp,
		Volumes:                       r.constructVolumes(job),
		Affinity: &corev1.Affinity{
			PodAntiAffinity: &corev1.PodAntiAffinity{
				RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
					{
						TopologyKey: "kubernetes.io/hostname",
						LabelSelector: &metav1.LabelSelector{
							MatchLabels: ClusterResourceLabels(job),
						},
					},
				},
			},
		},
	}, nil
}

func (r *SeatunnelJobReconciler) constructWorkload(job *seatunnelv1.SeatunnelJob) (*batchv1.Job, error) {
	podSpec, err := r.constructPodSpec(job)
	if err != nil {
		return nil, err
	}
	stsDesired := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(job),
			Namespace: job.Namespace,
			Labels:    ClusterResourceLabels(job),
		},
		Spec: batchv1.JobSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: ClusterResourceLabels(job),
			},
			BackoffLimit: int32Ptr(getReplicas(job)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ClusterResourceLabels(job),
				},
				Spec: *podSpec,
			},
		},
	}

	if err := ctrl.SetControllerReference(job, stsDesired, r.Scheme); err != nil {
		return stsDesired, err
	}
	return stsDesired, nil
}
