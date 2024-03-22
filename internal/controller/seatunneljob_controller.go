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
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"strings"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	seatunnelv1 "github.com/nineinfra/seatunnel-operator/api/v1"
)

type reconcileFun func(job *seatunnelv1.SeatunnelJob) error

// SeatunnelJobReconciler reconciles a SeatunnelJob object
type SeatunnelJobReconciler struct {
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
func (r *SeatunnelJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	r.logger = log.FromContext(ctx)
	r.ctx = ctx
	var job seatunnelv1.SeatunnelJob
	err := r.Get(ctx, req.NamespacedName, &job)
	if err != nil {
		if errors.IsNotFound(err) {
			r.logger.Info("Object not found, it could have been deleted")
		} else {
			r.logger.Error(err, "Error occurred during fetching the object")
		}
		return ctrl.Result{}, err
	}
	requestArray := strings.Split(fmt.Sprint(req), "/")
	requestName := requestArray[1]
	r.logger.Info(fmt.Sprintf("Reconcile requestName %s,job.Name %s", requestName, job.Name))
	if requestName == job.Name {
		r.logger.Info("Create or update clusters")
		err = r.reconcileJobs(&job)
		if err != nil {
			r.logger.Error(err, "Error occurred during create or update clusters")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *SeatunnelJobReconciler) reconcileServiceAccount(job *seatunnelv1.SeatunnelJob) (err error) {
	desiredSA, err := r.constructServiceAccount(job)
	if err != nil {
		return err
	}
	existsSA := &corev1.ServiceAccount{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredSA.Name, Namespace: desiredSA.Namespace}, existsSA)
	if err != nil && errors.IsNotFound(err) {
		r.logger.Info("Creating a new ServiceAccount")
		err = r.Client.Create(context.TODO(), desiredSA)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		//Todo
	}
	return nil
}

func (r *SeatunnelJobReconciler) reconcileRole(job *seatunnelv1.SeatunnelJob) (err error) {
	desiredRole, err := r.constructRole(job)
	if err != nil {
		return err
	}
	existsRole := &rbacv1.Role{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredRole.Name, Namespace: desiredRole.Namespace}, existsRole)
	if err != nil && errors.IsNotFound(err) {
		r.logger.Info("Creating a new Role")
		err = r.Client.Create(context.TODO(), desiredRole)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		r.logger.Info("Updating existing Role")
		existsRole.Rules = desiredRole.Rules
		err = r.Client.Update(context.TODO(), existsRole)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SeatunnelJobReconciler) reconcileRoleBinding(job *seatunnelv1.SeatunnelJob) (err error) {
	desiredRoleBinding, err := r.constructRoleBinding(job)
	if err != nil {
		return err
	}
	existsRoleBinding := &rbacv1.RoleBinding{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredRoleBinding.Name, Namespace: desiredRoleBinding.Namespace}, existsRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		r.logger.Info("Creating a new RoleBinding")
		err = r.Client.Create(context.TODO(), desiredRoleBinding)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		//Todo
	}
	return nil
}

func (r *SeatunnelJobReconciler) reconcileConfigMap(job *seatunnelv1.SeatunnelJob) (err error) {
	desiredCm, err := r.constructConfigMap(job)
	if err != nil {
		return err
	}
	existsCm := &corev1.ConfigMap{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredCm.Name, Namespace: desiredCm.Namespace}, existsCm)
	if err != nil && errors.IsNotFound(err) {
		r.logger.Info("Creating a new ConfigMap")
		err = r.Client.Create(context.TODO(), desiredCm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		r.logger.Info("Updating existing ConfigMap")
		existsCm.Data = desiredCm.Data
		err = r.Client.Update(context.TODO(), existsCm)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *SeatunnelJobReconciler) reconcileWorkload(job *seatunnelv1.SeatunnelJob) (err error) {
	desiredJob, err := r.constructWorkload(job)
	if err != nil {
		return err
	}
	existsJob := &batchv1.Job{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: desiredJob.Name, Namespace: desiredJob.Namespace}, existsJob)
	if err != nil && errors.IsNotFound(err) {
		r.logger.Info("Creating a new SeatunnelJob")
		err = r.Client.Create(context.TODO(), desiredJob)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	} else {
		//Todo
	}
	r.logger.Info(fmt.Sprintf("Reconcile the SeatunnelJob %s successfully", desiredJob.Name))
	return nil
}

func (r *SeatunnelJobReconciler) reconcileJobs(job *seatunnelv1.SeatunnelJob) error {
	for _, fun := range []reconcileFun{
		r.reconcileConfigMap,
		r.reconcileServiceAccount,
		r.reconcileRole,
		r.reconcileRoleBinding,
		r.reconcileWorkload,
	} {
		if err := fun(job); err != nil {
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
