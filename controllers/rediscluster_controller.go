/*
Copyright 2022.

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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"fmt"
	redisv1 "operator-redis/api/v1"
)

// RedisClusterReconciler reconciles a RedisCluster object
type RedisClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redis.jiang.operator,resources=redisclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redis.jiang.operator,resources=redisclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redis.jiang.operator,resources=redisclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the RedisCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *RedisClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	mylog := log.FromContext(ctx)

	mylog.Info("Start Reconcile Loop")

	var redisApp redisv1.RedisCluster
	err := r.Get(ctx, req.NamespacedName, &redisApp)
	if err != nil {

		if client.IgnoreNotFound(err) != nil {
			mylog.Info("not found RedisCluster resource")
			return ctrl.Result{}, err
		}
		return ctrl.Result{Requeue: true}, nil

	}

	var deployment appsv1.Deployment
	deployment.Name = redisApp.Name
	deployment.Namespace = redisApp.Namespace
	mutateDeployment, err := ctrl.CreateOrUpdate(ctx, r.Client, &deployment, func() error {
		MutateDeployment(&redisApp, &deployment)

		err := controllerutil.SetOwnerReference(&redisApp, &deployment, r.Scheme)
		return err
	})

	if err != nil {
		return ctrl.Result{}, err
	}
	mylog.Info("CreateOrUpdate", "RedisDeployment", mutateDeployment)

	var service corev1.Service
	service.Name = redisApp.Name
	service.Namespace = redisApp.Namespace

	// 判断是否需要对外服务
	if redisApp.Spec.Service {
		//
		mutateService, err := ctrl.CreateOrUpdate(ctx, r.Client, &service, func() error {
			MutateService(&redisApp, &service)

			err := controllerutil.SetOwnerReference(&redisApp, &service, r.Scheme)
			return err
		})
		if err != nil {
			return ctrl.Result{}, err
		}
		mylog.Info("CreateOrUpdate", "RedisService", mutateService)
	} else {
		err := r.Get(ctx, req.NamespacedName, &service)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				mylog.Info("not found Service resource")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
		err = r.Delete(ctx, &service)
		if err != nil {
			if client.IgnoreNotFound(err) != nil {
				mylog.Info("not found Service resource")
				return ctrl.Result{}, err
			}
			return ctrl.Result{Requeue: true}, nil
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RedisClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&redisv1.RedisCluster{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Watches(&source.Kind{ // 加入监听。
			Type: &appsv1.Deployment{},
		}, handler.Funcs{
			DeleteFunc: r.deploymentDeleteHandler,
		}).
		Watches(&source.Kind{ // 加入监听。
			Type: &corev1.Service{},
		}, handler.Funcs{
			DeleteFunc: r.serviceDeleteHandler,
		}).
		Complete(r)
}

func (r *RedisClusterReconciler) deploymentDeleteHandler(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	fmt.Println("被删除的对象名称是", event.Object.GetName())
	for _, ref := range event.Object.GetOwnerReferences() {
		if ref.Kind == redisv1.Kind && ref.APIVersion == redisv1.ApiVersion {
			// 重新入列，这样删除pod后，就会进入调和loop，发现owerReference还在，会立即创建出新的pod。
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ref.Name,
					Namespace: event.Object.GetNamespace()}})
		}
	}
}

func (r *RedisClusterReconciler) serviceDeleteHandler(event event.DeleteEvent, limitingInterface workqueue.RateLimitingInterface) {
	fmt.Println("被删除的对象名称是", event.Object.GetName())
	for _, ref := range event.Object.GetOwnerReferences() {
		if ref.Kind == redisv1.Kind && ref.APIVersion == redisv1.ApiVersion {
			// 重新入列，这样删除pod后，就会进入调和loop，发现owerReference还在，会立即创建出新的pod。
			limitingInterface.Add(reconcile.Request{
				NamespacedName: types.NamespacedName{Name: ref.Name,
					Namespace: event.Object.GetNamespace()}})
		}
	}
}
