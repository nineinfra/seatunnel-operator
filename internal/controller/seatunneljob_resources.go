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
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"strconv"
	"strings"

	kyuubiv1 "github.com/nineinfra/kyuubi-operator/api/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
)

func getReplicas(job *seatunnelv1.SeatunnelJob) int32 {
	if job.Spec.Resource.Replicas != 0 {
		return job.Spec.Resource.Replicas
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
	sb.WriteString(map2String(conf, 2, seatunnelv1.NINEINFRA_NAMED_ENV_LIST))
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
			sb.WriteString(map2String(source.Conf, 4, nil))
		}
		if source.TableList != nil {
			writeSpaces(&sb, 4)
			sb.WriteString(fmt.Sprintf("%s=[\n", "table_list"))
			for _, v := range source.TableList {
				writeSpaces(&sb, 6)
				sb.WriteString("{\n")
				sb.WriteString(map2String(v, 8, nil))
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
			sb.WriteString(map2String(source.Scheme.Fields, 8, nil))
			writeSpaces(&sb, 6)
			sb.WriteString("}\n")
			writeSpaces(&sb, 4)
			sb.WriteString("}\n")
		}
		if source.Properties != nil {
			writeSpaces(&sb, 4)
			sb.WriteString("properties{\n")
			sb.WriteString(map2String(source.Properties, 6, nil))
			writeSpaces(&sb, 4)
			sb.WriteString("}\n")
		}
		if source.ExtraConfig != nil {
			writeSpaces(&sb, 4)
			sb.WriteString(fmt.Sprintf("%s.config={", strings.ToLower(source.Type)))
			sb.WriteString(map2String(source.ExtraConfig, 6, nil))
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
		sb.WriteString(map2String(transform.Conf, 4, nil))
		if transform.Scheme.Fields != nil {
			writeSpaces(&sb, 4)
			sb.WriteString("schema={\n")
			writeSpaces(&sb, 6)
			sb.WriteString("fields{\n")
			sb.WriteString(map2String(transform.Scheme.Fields, 8, nil))
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
			sb.WriteString(map2String(sink.Conf, 4, nil))
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
			sb.WriteString(fmt.Sprintf("%s.config={\n", strings.ToLower(sink.Type)))
			sb.WriteString(map2String(sink.ExtraConfig, 6, nil))
			writeSpaces(&sb, 4)
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
	return map2String(tmpConf, 0, nil)
}

func getImageConfig(job *seatunnelv1.SeatunnelJob) seatunnelv1.ImageConfig {
	ic := seatunnelv1.ImageConfig{
		Repository:  job.Spec.Image.Repository,
		PullSecrets: job.Spec.Image.PullSecrets,
	}
	ic.Tag = job.Spec.Image.Tag
	if ic.Tag == "" {
		ic.Tag = job.Spec.Version
	}
	ic.PullPolicy = job.Spec.Image.PullPolicy
	if ic.PullPolicy == "" {
		ic.PullPolicy = string(corev1.PullIfNotPresent)
	}
	return ic
}

func (r *SeatunnelJobReconciler) getNameSpacedKyuubiCluster(job *seatunnelv1.SeatunnelJob) (*kyuubiv1.KyuubiCluster, error) {
	metav1.AddToGroupVersion(r.Scheme, kyuubiv1.GroupVersion)
	utilruntime.Must(kyuubiv1.AddToScheme(r.Scheme))

	existsKCList := &kyuubiv1.KyuubiClusterList{}
	err := r.Client.List(context.TODO(), existsKCList, &ctrlclient.ListOptions{Namespace: job.Namespace})
	if err != nil {
		r.logger.Info(fmt.Sprintf("%s%s", "Get kyuubiclusterlist failed,err:", err.Error()))
		return nil, err
	}
	if len(existsKCList.Items) < 1 {
		return nil, err
	}
	return &existsKCList.Items[0], nil
}

func (r *SeatunnelJobReconciler) getClusterRefsConfigMapName(job *seatunnelv1.SeatunnelJob) string {
	kc, err := r.getNameSpacedKyuubiCluster(job)
	if err != nil {
		return NineResourceName(job, DefaultClusterRefsConfigNameSuffix)
	}
	return fmt.Sprintf("%s%s", kc.Name, DefaultClusterRefsConfigNameSuffix)
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

func (r *SeatunnelJobReconciler) getClusterRefsConfigMap(job *seatunnelv1.SeatunnelJob) *corev1.ConfigMap {
	clusterRefsCM := &corev1.ConfigMap{}
	_ = r.Client.Get(context.TODO(), types.NamespacedName{Name: r.getClusterRefsConfigMapName(job), Namespace: job.Namespace}, clusterRefsCM)
	return clusterRefsCM
}

func (r *SeatunnelJobReconciler) constructClusterRefsVolumeMounts(job *seatunnelv1.SeatunnelJob) []corev1.VolumeMount {
	clusterRefsCM := r.getClusterRefsConfigMap(job)
	volumeMounts := make([]corev1.VolumeMount, 0)
	for _, confFile := range ClusterRefsConfFileList {
		if _, ok := clusterRefsCM.Data[confFile]; ok {
			volumeMounts = append(volumeMounts, corev1.VolumeMount{
				Name:      r.getClusterRefsConfigMapName(job),
				MountPath: fmt.Sprintf("%s/%s", DefaultSparkConfDir, confFile),
				SubPath:   confFile,
			})
		}
	}
	return volumeMounts
}

func (r *SeatunnelJobReconciler) constructClusterRefsVolumes(job *seatunnelv1.SeatunnelJob) []corev1.Volume {
	clusterRefsCM := r.getClusterRefsConfigMap(job)
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
			Name: r.getClusterRefsConfigMapName(job),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: r.getClusterRefsConfigMapName(job),
					},
					Items: keyToPaths,
				},
			},
		},
	}
}

func (r *SeatunnelJobReconciler) constructVolumeMounts(job *seatunnelv1.SeatunnelJob) []corev1.VolumeMount {
	volumeMounts := []corev1.VolumeMount{
		{
			Name:      ClusterResourceName(job, DefaultConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/%s", DefaultConfPath, DefaultConfigFileName),
			SubPath:   DefaultConfigFileName,
		},
		{
			Name:      ClusterResourceName(job, DefaultConfigNameSuffix),
			MountPath: fmt.Sprintf("%s/%s", DefaultConfPath, DefaultLogConfigFileName),
			SubPath:   DefaultLogConfigFileName,
		},
	}
	clusterRefsVolumeMounts := r.constructClusterRefsVolumeMounts(job)
	for _, v := range clusterRefsVolumeMounts {
		volumeMounts = append(volumeMounts, v)
	}
	return volumeMounts
}

func (r *SeatunnelJobReconciler) constructVolumes(job *seatunnelv1.SeatunnelJob) []corev1.Volume {
	volumes := []corev1.Volume{
		{
			Name: ClusterResourceName(job, DefaultConfigNameSuffix),
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: ClusterResourceName(job, DefaultConfigNameSuffix),
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

	clusterRefsVolumes := r.constructClusterRefsVolumes(job)
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
	debugMode := GetSeatunnelDebugMode(job)
	cmdString := make([]string, 0)
	if debugMode == seatunnelv1.NINEINFRA_ENV_SEATUNNEL_DEBUG_MODE_ON {
		cmdString = []string{
			"sleep",
			strconv.Itoa(GetSeatunnelDebugTime(job)),
		}
	} else {
		seatunnelEngine := GetSeatunnelEngine(job)
		switch seatunnelEngine {
		case seatunnelv1.NINEINFRA_ENV_SEATUNNEL_ENGINE_FLINK:
			//todo
		case seatunnelv1.NINEINFRA_ENV_SEATUNNEL_ENGINE_SPARK:
			cmdString = []string{
				"/opt/seatunnel/bin/start-seatunnel-spark-3-connector-v2.sh",
				"--config",
				DefaultConfFile,
				"--master",
				fmt.Sprintf("k8s://https://%s:%s", os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")),
			}
		}
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
