/*
Copyright 2023.

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

package controllers

import (
	"context"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	v1 "aor.dev/example/api/v1"
)

// AorJoaPainKillerReconciler reconciles a AorJoaPainKiller object
type AorJoaPainKillerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=example.aor.dev,resources=aorjoapainkillers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=example.aor.dev,resources=aorjoapainkillers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=example.aor.dev,resources=aorjoapainkillers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the AorJoaPainKiller object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *AorJoaPainKillerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	// TODO(user): your logic here
	var pkl v1.AorJoaPainKiller

	if err := r.Get(ctx, req.NamespacedName, &pkl); err != nil {
		log.Log.Error(err, "unable to fetch")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var podList corev1.PodList
	var pillFound bool
	if err := r.List(ctx, &podList); err != nil {
		log.Log.Error(err, "unable to list pods")
	} else {
		for _, item := range podList.Items {
			if strings.HasPrefix(item.GetName(), "pill") {
				log.Log.Info("pod name with prefix pill was found", "name", item.GetName())
				pillFound = true
			}
		}
	}

	pkl.Status.Cure = pillFound

	if err := r.Status().Update(ctx, &pkl); err != nil {
		log.Log.Error(err, "unable to update aorjoa's pain status", "status", pillFound)
		return ctrl.Result{}, err
	}

	log.Log.Info("custom resource reconciled")

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *AorJoaPainKillerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.AorJoaPainKiller{}).
		Watches(&source.Kind{Type: &corev1.Pod{}},
			handler.EnqueueRequestsFromMapFunc(r.mapPodsReqToFooReq)).
		Complete(r)
}

func (r *AorJoaPainKillerReconciler) mapPodsReqToFooReq(obj client.Object) []reconcile.Request {
	ctx := context.Background()
	log := log.FromContext(ctx)

	// List all the Foo custom resource
	req := []reconcile.Request{}
	var list v1.AorJoaPainKillerList
	if err := r.Client.List(context.TODO(), &list); err != nil {
		log.Error(err, "unable to list foo custom resources")
	} else {
		// Only keep Foo custom resources related to the Pod that triggered the reconciliation request
		for _, item := range list.Items {
			if strings.HasPrefix(item.GetName(), "pill") {
				req = append(req, reconcile.Request{
					NamespacedName: types.NamespacedName{Name: item.Name, Namespace: item.Namespace},
				})
				log.Info("pod linked to a foo custom resource issued an event", "name", obj.GetName())
			}
		}
	}
	return req
}
