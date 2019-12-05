package gslb

import (
	"context"
	"fmt"
	"reflect"

	ohmyglbv1beta1 "github.com/AbsaOSS/ohmyglb/pkg/apis/ohmyglb/v1beta1"
	v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_gslb")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Gslb Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGslb{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("gslb-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Gslb
	err = c.Watch(&source.Kind{Type: &ohmyglbv1beta1.Gslb{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource GslbResolvers and requeue the owner Gslb
	err = c.Watch(&source.Kind{Type: &ohmyglbv1beta1.GslbResolver{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ohmyglbv1beta1.Gslb{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileGslb implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileGslb{}

// ReconcileGslb reconciles a Gslb object
type ReconcileGslb struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Gslb object and makes changes based on the state read
// and what is in the Gslb.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a GslbResolver as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGslb) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Gslb")

	// Fetch the Gslb instance
	instance := &ohmyglbv1beta1.Gslb{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new GslbResolver object
	gslbResolver := newGslbResolverForCR(instance)

	// Set Gslb instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, gslbResolver, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this GslbResolver already exists
	found := &ohmyglbv1beta1.GslbResolver{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: gslbResolver.Name, Namespace: gslbResolver.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new GslbResolver", "GslbResolver.Namespace", gslbResolver.Namespace, "GslbResolver.Name", gslbResolver.Name)
		err = r.client.Create(context.TODO(), gslbResolver)
		if err != nil {
			return reconcile.Result{}, err
		}

		// GslbResolver created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// GslbResolver already exists - don't requeue
	reqLogger.Info("Skip reconcile: GslbResolver already exists", "GslbResolver.Namespace", found.Namespace, "GslbResolver.Name", found.Name)

	//Update Gslb status with managed host list that are getting retrieved from Ingress objects with special annotation
	ingressList := &v1beta1.IngressList{}
	listOpts := []client.ListOption{
		client.InNamespace(instance.Namespace),
		client.MatchingLabels{"gslb": "true"},
	}
	if err = r.client.List(context.TODO(), ingressList, listOpts...); err != nil {
		reqLogger.Error(err, "Failed to list ingresses", "Gslb.Namespace", instance.Namespace, "Gslb.Name", instance.Name)
		return reconcile.Result{}, err
	}

	gslbHosts := getIngressHosts(ingressList.Items)

	// Update status.Nodes if needed
	if !reflect.DeepEqual(gslbHosts, instance.Status.Hosts) {
		instance.Status.Hosts = gslbHosts
		err := r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update Gslb status")
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

// newGslbResolverForCR returns a busybox gslbResolver with the same name/namespace as the cr
func newGslbResolverForCR(cr *ohmyglbv1beta1.Gslb) *ohmyglbv1beta1.GslbResolver {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &ohmyglbv1beta1.GslbResolver{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-gslbresolver",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: ohmyglbv1beta1.GslbResolverSpec{
			Size: 3,
		},
	}
}

func getIngressHosts(ingresses []v1beta1.Ingress) []string {
	var ingressHosts []string
	for _, ingress := range ingresses {
		ingressHosts = append(ingressHosts, fmt.Sprintf("%#v", ingress.Spec))
	}
	return ingressHosts
}