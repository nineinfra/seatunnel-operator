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
	seatunnelv1 "github.com/nineinfra/seatunnel-operator/api/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type reconcileJobFun func(seatunnelJob *seatunnelv1.SeatunnelJob) error

// JobReconciler reconciles a Job object
type JobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	logger logr.Logger
	ctx    context.Context
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
func (r *JobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx)
	r.ctx = ctx
	var job batchv1.Job
	err := r.Get(ctx, req.NamespacedName, &job)
	if err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info("Object not found, it could have been deleted")
		} else {
			r.logger.Error(err, "Error occurred during fetching the object")
		}
		return ctrl.Result{}, err
	}

	if len(job.OwnerReferences) < 1 {
		r.logger.Info(fmt.Sprintf("OwnerReference for job %s not found", job.Name))
		return ctrl.Result{}, nil
	}
	seatunneljobName := job.OwnerReferences[0].Name
	var seatunnelJob seatunnelv1.SeatunnelJob
	err = r.Get(ctx, types.NamespacedName{Name: seatunneljobName, Namespace: req.Namespace}, &seatunnelJob)
	if err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info(fmt.Sprintf("SeatunnelJob %s not found, it could have been deleted", seatunneljobName))
		} else {
			r.logger.Error(err, "Error occurred during fetching the object")
		}
		return ctrl.Result{}, err
	}
	requestArray := strings.Split(fmt.Sprint(req), "/")
	requestName := requestArray[1]
	r.logger.Info(fmt.Sprintf("Reconcile requestName %s,job.Name %s", requestName, job.Name))
	if requestName == job.Name {
		r.logger.Info("Update clusters")
		err = r.reconcileJobs(&seatunnelJob)
		if err != nil {
			r.logger.Error(err, "Error occurred during update clusters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *JobReconciler) notifySeatunnelJob(job *seatunnelv1.SeatunnelJob) (err error) {
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
		succeededMembers   []string
		unsucceededMembers []string
		failedMembers      []string
	)
	for _, p := range existsPods.Items {
		switch p.Status.Phase {
		case corev1.PodSucceeded:
			succeededMembers = append(succeededMembers, p.Name)
		case corev1.PodFailed:
			failedMembers = append(failedMembers, p.Name)
		default:
			unsucceededMembers = append(unsucceededMembers, p.Name)
		}
	}
	job.Status.Members.Succeeded = succeededMembers
	job.Status.Members.Unsucceeded = unsucceededMembers
	job.Status.Members.Failed = failedMembers

	job.Status.SucceededReplicas = int32(len(succeededMembers))
	if job.Status.SucceededReplicas == 1 {
		job.Status.SetJobSucceededConditionTrue()
		job.Status.Completed = true
	} else {
		job.Status.SetJobSucceededConditionFalse()
		job.Status.Completed = false
	}
	r.logger.Info("Updating SeatunnelJob by JobReconciler")
	return r.Client.Status().Update(context.TODO(), job)
}

func (r *JobReconciler) reconcileJobs(job *seatunnelv1.SeatunnelJob) error {
	for _, fun := range []reconcileJobFun{
		r.notifySeatunnelJob,
	} {
		if err := fun(job); err != nil {
			return err
		}
	}

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *JobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.Job{}).WithEventFilter(predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			name := e.ObjectNew.GetName()
			return strings.HasSuffix(name, DefaultNameSuffix)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
	}).Complete(r)
}
