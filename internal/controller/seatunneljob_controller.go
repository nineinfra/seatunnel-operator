/*
Copyright 2024 nineinfra.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	seatunnelv1 "github.com/nineinfra/seatunnel-operator/api/v1"
)

type reconcileFun func(ctx context.Context, job *seatunnelv1.SeatunnelJob, logger logr.Logger) error

// SeatunnelJobReconciler reconciles a SeatunnelJob object
type SeatunnelJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=seatunnel.nineinfra.tech,resources=seatunneljobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=seatunnel.nineinfra.tech,resources=seatunneljobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=seatunnel.nineinfra.tech,resources=seatunneljobs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the SeatunnelJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.0/pkg/reconcile
func (r *SeatunnelJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var job seatunnelv1.SeatunnelJob
	err := r.Get(ctx, req.NamespacedName, &job)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Object not found, it could have been deleted")
		} else {
			logger.Error(err, "Error occurred during fetching the object")
		}
		return ctrl.Result{}, err
	}
	requestArray := strings.Split(fmt.Sprint(req), "/")
	requestName := requestArray[1]
	logger.Info(fmt.Sprintf("Reconcile requestName %s,job.Name %s", requestName, job.Name))
	if requestName == job.Name {
		logger.Info("Create or update clusters")
		err = r.reconcileJobs(ctx, &job, logger)
		if err != nil {
			logger.Error(err, "Error occurred during create or update clusters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *SeatunnelJobReconciler) reconcileJobStatus(ctx context.Context, job *seatunnelv1.SeatunnelJob, logger logr.Logger) (err error) {
	job.Status.Init()
	existsPods := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(ClusterResourceLabels(job))
	listOps := &client.ListOptions{
		Namespace:     job.Namespace,
		LabelSelector: labelSelector,
	}
	err = r.Client.List(context.TODO(), existsPods, listOps)
	if err != nil {
		return err
	}
	var (
		readyMembers   []string
		unreadyMembers []string
	)
	for _, p := range existsPods.Items {
		ready := true
		for _, c := range p.Status.ContainerStatuses {
			if !c.Ready {
				ready = false
			}
		}
		if ready {
			readyMembers = append(readyMembers, p.Name)
		} else {
			unreadyMembers = append(unreadyMembers, p.Name)
		}
	}
	job.Status.Members.Ready = readyMembers
	job.Status.Members.Unready = unreadyMembers

	logger.Info("Updating cluster status")
	if job.Status.ReadyReplicas == job.Spec.Resource.Replicas {
		job.Status.SetPodsReadyConditionTrue()
	} else {
		job.Status.SetPodsReadyConditionFalse()
	}
	if job.Status.CurrentVersion == "" && job.Status.IsClusterInReadyState() {
		job.Status.CurrentVersion = job.Spec.Image.Tag
	}
	return r.Client.Status().Update(context.TODO(), job)
}

func (r *SeatunnelJobReconciler) reconcileConfigMap(ctx context.Context, job *seatunnelv1.SeatunnelJob, logger logr.Logger) (err error) {
	desiredCm, err := r.constructConfigMap(job)
	if err != nil {
		return err
	}
	existsCm := &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredCm.Name, Namespace: desiredCm.Namespace}, existsCm)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new ConfigMap")
		err = r.Client.Create(context.TODO(), desiredCm)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	} else {
		logger.Info("Updating existing ConfigMap")
		existsCm.Data = desiredCm.Data
		err = r.Client.Update(context.TODO(), existsCm)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *SeatunnelJobReconciler) reconcileWorkload(ctx context.Context, job *seatunnelv1.SeatunnelJob, logger logr.Logger) (err error) {
	desiredJob, err := r.constructWorkload(job)
	if err != nil {
		return err
	}
	existsSts := &appsv1.StatefulSet{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredJob.Name, Namespace: desiredJob.Namespace}, existsSts)
	if err != nil && errors.IsNotFound(err) {
		logger.Info("Creating a new SeatunnelJob")
		err = r.Client.Create(context.TODO(), desiredJob)
		if err != nil {
			return err
		}
		return nil
	} else if err != nil {
		return err
	}
	logger.Info("Creating a new SeatunnelJob successfully")
	return nil
}

func (r *SeatunnelJobReconciler) reconcileJobs(ctx context.Context, job *seatunnelv1.SeatunnelJob, logger logr.Logger) error {
	for _, fun := range []reconcileFun{
		r.reconcileConfigMap,
		r.reconcileWorkload,
		r.reconcileJobStatus,
	} {
		if err := fun(ctx, job, logger); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *SeatunnelJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&seatunnelv1.SeatunnelJob{}).
		Complete(r)
}
