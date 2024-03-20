package controller

import (
	"context"
	"errors"
	"fmt"
	seatunnelv1 "github.com/nineinfra/seatunnel-operator/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
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
	sb.WriteString(map2String(conf, 2))
	sb.WriteString("}\n")
	return sb.String()
}

func writeSourceConfig(source *seatunnelv1.SourceConfig) string {
	var sb strings.Builder
	sb.WriteString("source{\n")
	if !checkSourceTypeSupported(source.Type) {
		return ""
	}
	if source.Type != "" {
		writeSpaces(&sb, 2)
		sb.WriteString(fmt.Sprintf("%s{\n", source.Type))
		if source.Conf != nil {
			sb.WriteString(map2String(source.Conf, 4))
		}
		if source.TableList != nil {
			writeSpaces(&sb, 4)
			sb.WriteString(fmt.Sprintf("%s=[\n", "table_list"))
			for _, v := range source.TableList {
				writeSpaces(&sb, 6)
				sb.WriteString("{\n")
				sb.WriteString(map2String(v, 8))
				writeSpaces(&sb, 6)
				sb.WriteString("},\n")
			}
			writeSpaces(&sb, 4)
			sb.WriteString("]\n")
		}
		if source.Scheme.Fields != nil {
			writeSpaces(&sb, 4)
			sb.WriteString("schema={\n")
			writeSpaces(&sb, 6)
			sb.WriteString("fields{\n")
			sb.WriteString(map2String(source.Scheme.Fields, 8))
			writeSpaces(&sb, 6)
			sb.WriteString("}\n")
			writeSpaces(&sb, 4)
			sb.WriteString("}\n")
		}
		if source.Properties != nil {
			writeSpaces(&sb, 4)
			sb.WriteString("properties{\n")
			sb.WriteString(map2String(source.Properties, 6))
			writeSpaces(&sb, 4)
			sb.WriteString("}\n")
		}
		if source.ExtraConfig != nil {
			writeSpaces(&sb, 4)
			sb.WriteString(fmt.Sprintf("%s.config={", strings.ToLower(source.Type)))
			sb.WriteString(map2String(source.ExtraConfig, 6))
			writeSpaces(&sb, 4)
			sb.WriteString("}\n")
		}
		writeSpaces(&sb, 2)
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

func writeTransformConfig(transform *seatunnelv1.TransformConfig) string {
	var sb strings.Builder
	sb.WriteString("transform{\n")
	if !checkTransformTypeSupported(transform.Type) {
		return ""
	}
	if transform.Type != "" {
		writeSpaces(&sb, 2)
		sb.WriteString(fmt.Sprintf("%s{\n", transform.Type))
		sb.WriteString(map2String(transform.Conf, 4))
		if transform.Scheme.Fields != nil {
			writeSpaces(&sb, 4)
			sb.WriteString("schema={\n")
			writeSpaces(&sb, 6)
			sb.WriteString("fields{\n")
			sb.WriteString(map2String(transform.Scheme.Fields, 8))
			writeSpaces(&sb, 6)
			sb.WriteString("}\n")
			writeSpaces(&sb, 4)
			sb.WriteString("}\n")
		}
		writeSpaces(&sb, 2)
		sb.WriteString("}\n")
	}
	sb.WriteString("}\n")
	return sb.String()
}

func writeSinkConfig(sink *seatunnelv1.SinkConfig) string {
	var sb strings.Builder
	sb.WriteString("sink{\n")
	if !checkSinkTypeSupported(sink.Type) {
		return ""
	}
	if sink.Type != "" {
		writeSpaces(&sb, 2)
		sb.WriteString(fmt.Sprintf("%s{\n", sink.Type))
		if sink.Conf != nil {
			sb.WriteString(map2String(sink.Conf, 4))
		}
		if sink.PartitionBy != nil {
			writeSpaces(&sb, 4)
			sb.WriteString("partition_by=[")
			sb.WriteString(list2String(sink.PartitionBy))
			sb.WriteString("]\n")
		}
		if sink.SinkColumns != nil {
			writeSpaces(&sb, 4)
			sb.WriteString("sink_columns=[")
			sb.WriteString(list2String(sink.SinkColumns))
			sb.WriteString("]\n")
		}
		if sink.ExtraConfig != nil {
			writeSpaces(&sb, 4)
			sb.WriteString(fmt.Sprintf("%s.config={", strings.ToLower(sink.Type)))
			sb.WriteString(map2String(sink.ExtraConfig, 6))
			sb.WriteString("}\n")
		}
		writeSpaces(&sb, 2)
		sb.WriteString("}\n")
	}
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
	return map2String(tmpConf, 0)
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

func (r *SeatunnelJobReconciler) getClusterRefsConfigMap(cluster *seatunnelv1.SeatunnelJob) *corev1.ConfigMap {
	clusterRefsCM := &corev1.ConfigMap{}
	_ = r.Client.Get(context.TODO(), types.NamespacedName{Name: NineResourceName(cluster, DefaultClusterRefsConfigNameSuffix), Namespace: cluster.Namespace}, clusterRefsCM)
	return clusterRefsCM
}

func (r *SeatunnelJobReconciler) constructClusterRefsVolumeMounts(cluster *seatunnelv1.SeatunnelJob) []corev1.VolumeMount {
	clusterRefsCM := r.getClusterRefsConfigMap(cluster)
	volumeMounts := make([]corev1.VolumeMount, 0)
	for _, confFile := range ClusterRefsConfFileList {
		if _, ok := clusterRefsCM.Data[confFile]; ok {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      NineResourceName(cluster, DefaultClusterRefsConfigNameSuffix),
				MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, confFile),
				SubPath:   confFile,
			})
		}
	}
	return volumeMounts
}

func (r *SeatunnelJobReconciler) constructClusterRefsVolumes(cluster *seatunnelv1.SeatunnelJob) []corev1.Volume {
	clusterRefsCM := r.getClusterRefsConfigMap(cluster)
	keyToPaths := make([]corev1.KeyToPath, 0)
	for _, confFile := range ClusterRefsConfFileList {
		if _, ok := clusterRefsCM.Data[confFile]; ok {
			keyToPaths = append(keyToPaths, corev1.KeyToPath{
				Key:  confFile,
				Path: confFile,
			})
		}
	}
	return []corev1.Volume{
		{
			Name: NineResourceName(cluster, DefaultClusterRefsConfigNameSuffix),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: NineResourceName(cluster, DefaultClusterRefsConfigNameSuffix),
					},
					Items: keyToPaths,
				},
			},
		},
	}
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
	clusterRefsVolumeMounts := r.constructClusterRefsVolumeMounts(cluster)
	for _, v := range clusterRefsVolumeMounts {
		volumeMounts = append(volumeMounts, v)
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

	clusterRefsVolumes := r.constructClusterRefsVolumes(cluster)
	for _, v := range clusterRefsVolumes {
		volumes = append(volumes, v)
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
		RestartPolicy:                 corev1.RestartPolicyOnFailure,
		TerminationGracePeriodSeconds: &tgp,
		Volumes:                       r.constructVolumes(job),
		ServiceAccountName:            ClusterResourceName(job),
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

func (r *SeatunnelJobReconciler) constructServiceAccount(job *seatunnelv1.SeatunnelJob) (*corev1.ServiceAccount, error) {
	saDesired := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(job),
			Namespace: job.Namespace,
			Labels:    ClusterResourceLabels(job),
		},
	}

	if err := ctrl.SetControllerReference(job, saDesired, r.Scheme); err != nil {
		return saDesired, err
	}
	return saDesired, nil
}

func (r *SeatunnelJobReconciler) constructRole(job *seatunnelv1.SeatunnelJob) (*rbacv1.Role, error) {
	roleDesired := &rbacv1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(job),
			Namespace: job.Namespace,
			Labels:    ClusterResourceLabels(job),
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{
					"",
				},
				Resources: []string{
					"pods",
					"configmaps",
					"services",
					"persistentvolumeclaims",
				},
				Verbs: []string{
					"get",
					"create",
					"list",
					"delete",
					"watch",
					"deletecollection",
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(job, roleDesired, r.Scheme); err != nil {
		return roleDesired, err
	}
	return roleDesired, nil
}

func (r *SeatunnelJobReconciler) constructRoleBinding(job *seatunnelv1.SeatunnelJob) (*rbacv1.RoleBinding, error) {
	roleBindingDesired := &rbacv1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ClusterResourceName(job),
			Namespace: job.Namespace,
			Labels:    ClusterResourceLabels(job),
		},
		Subjects: []rbacv1.Subject{
			{
				Kind: "ServiceAccount",
				Name: ClusterResourceName(job),
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     ClusterResourceName(job),
		},
	}

	if err := ctrl.SetControllerReference(job, roleBindingDesired, r.Scheme); err != nil {
		return roleBindingDesired, err
	}
	return roleBindingDesired, nil
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
