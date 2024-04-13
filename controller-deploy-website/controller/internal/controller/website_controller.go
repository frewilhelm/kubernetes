/*
Copyright 2024.

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

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	webv1 "example.com/website/api/v1"
)

// WebSiteReconciler reconciles a WebSite object
type WebSiteReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=web.example.com,resources=websites,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=web.example.com,resources=websites/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=web.example.com,resources=websites/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the WebSite object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.17.2/pkg/reconcile
func (r *WebSiteReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	var log = ctrl.Log.WithName("controller_web_site")

	var website webv1.WebSite
	if err := r.Get(ctx, req.NamespacedName, &website); err != nil {
		log.Error(err, "unable to fetch WebSite")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if the Deployment already exists, if not create a new one
	var deployment appsv1.Deployment
	deploymentName := website.Name + "-deployment"
	depKey := types.NamespacedName{Name: deploymentName, Namespace: website.Namespace}
	if err := r.Get(ctx, depKey, &deployment); err != nil {
		if errors.IsNotFound(err) {
			// Define a new Deployment
			dep := r.deploymentForWebSite(&website)
			log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
			if err := r.Create(ctx, dep); err != nil {
				log.Error(err, "failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
				return ctrl.Result{}, err
			}
			// Deployment created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else {
			log.Error(err, "failed to get Deployment")
			return ctrl.Result{}, err
		}
	}

	// Ensure the deployment size is the same as the spec
	size := website.Spec.Replicas
	if deployment.Spec.Replicas != size {
		deployment.Spec.Replicas = size
		err := r.Update(ctx, &deployment)
		if err != nil {
			log.Error(err, "failed to update Deployment", "Deployment.Namespace", deployment.Namespace, "Deployment.Name", deployment.Name)
			return ctrl.Result{}, err
		}
	}

	// Check if Service already exsists, if not create a new one
	var service corev1.Service
	serviceName := website.Name + "-service"
	servKey := types.NamespacedName{Name: serviceName, Namespace: website.Namespace}
	if err := r.Get(ctx, servKey, &service); err != nil {
		if errors.IsNotFound(err) {
			// Define a new Service
			serv := r.serviceForWebSite(&website)
			log.Info("Create a new Service", "Service.Namespace", serv.Namespace, "Service.Name", serv.Name)
			if err := r.Create(ctx, serv); err != nil {
				log.Error(err, "failed to create new Service", "Service.Namespace", serv.Namespace, "Service.Name", serv.Name)
				return ctrl.Result{}, err
			}
			// Service created successfully - return and requeue
			return ctrl.Result{Requeue: true}, nil
		} else {
			log.Error(err, "failed to get Service")
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *WebSiteReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&webv1.WebSite{}).
		Complete(r)
}

func (r *WebSiteReconciler) deploymentForWebSite(ws *webv1.WebSite) *appsv1.Deployment {
	ls := labelsForWebSite(ws.Name)
	replicas := ws.Spec.Replicas

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ws.Name + "-deployment",
			Namespace: ws.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(ws, webv1.GroupVersion.WithKind("WebSite")),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "web",
							Image: ws.Spec.ImageName,
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
									Name:          "http",
								},
							},
							ImagePullPolicy: corev1.PullIfNotPresent, // Ensure to pull from local
						}},
				},
			},
		},
	}

	return dep
}

func labelsForWebSite(name string) map[string]string {
	return map[string]string{"type": "web", "app": name}
}

func (r *WebSiteReconciler) serviceForWebSite(ws *webv1.WebSite) *corev1.Service {
	ls := labelsForWebSite(ws.Name)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ws.Name + "-service",
			Namespace: ws.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: ls,
			Ports: []corev1.ServicePort{{
				Port:       80,
				TargetPort: intstr.FromInt(80),
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	return svc
}
