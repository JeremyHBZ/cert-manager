/*
Copyright 2021.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	certv1alpha1 "github.com/JeremyHBZ/cert-manager/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
)

// CertReconciler reconciles a Cert object
type CertReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=test.redhat.com,resources=certs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=test.redhat.com,resources=certs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=test.redhat.com,resources=certs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cert object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.7.0/pkg/reconcile
func (r *CertReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("cert", req.NamespacedName)

	// your logic here
	// Fetch the cert instance
	cert := &certv1alpha1.Cert{}
	err := r.Get(ctx, req.NamespacedName, cert)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("cert resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get cert")
		return ctrl.Result{}, err
	}

	// Check if the secret already exists, if not create a new one
	found := &corev1.Secret{}
	err = r.Get(ctx, types.NamespacedName{Name: secretName(cert), Namespace: cert.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// Define a new secret
		secret := secretForCert(cert)
		log.Info("Creating a new secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
		err = r.Create(ctx, secret)
		if err != nil {
			log.Error(err, "Failed to create new secret", "secret.Namespace", secret.Namespace, "secret.Name", secret.Name)
			return ctrl.Result{}, err
		}
		// Secret created successfully
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get secret")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}
func secretForCert(c *certv1alpha1.Cert) *corev1.Secret{
	key, cert:=secretData(c)
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: c.Namespace,
			Name:      secretName(c),
		},
		Data: map[string][]byte{
			"tls.key":[]byte(key),
			"tls.cert":[]byte(cert)},
	}
	return secret

}
func secretData(c *certv1alpha1.Cert) (string, string){
	key:="mock-tls.key"
	cert:="mock-tls.cert"
	return key, cert
}
func secretName(cert *certv1alpha1.Cert)string{
	return cert.Name+"-"+cert.Spec.CN
}

// SetupWithManager sets up the controller with the Manager.
func (r *CertReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&certv1alpha1.Cert{}).
		Owns(&corev1.Secret{}).
		Complete(r)
}
