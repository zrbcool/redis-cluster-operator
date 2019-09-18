package rediscluster

import (
	"context"
	"k8s.io/apimachinery/pkg/labels"
	"strconv"
	"time"

	redisoperatorv1alpha1 "github.com/zrbcool/redis-cluster-operator/pkg/apis/redisoperator/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_rediscluster")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new RedisCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRedisCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rediscluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource RedisCluster
	err = c.Watch(&source.Kind{Type: &redisoperatorv1alpha1.RedisCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner RedisCluster
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &redisoperatorv1alpha1.RedisCluster{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRedisCluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRedisCluster{}

// ReconcileRedisCluster reconciles a RedisCluster object
type ReconcileRedisCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a RedisCluster object and makes changes based on the state read
// and what is in the RedisCluster.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRedisCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling RedisCluster")

	// Fetch the RedisCluster instance
	instance := &redisoperatorv1alpha1.RedisCluster{}
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

	var instances []string
	for rsIndex := 1; rsIndex <= instance.Spec.Size; rsIndex++ {
		replicaSet := &appsv1.ReplicaSet{}
		rsName := instance.Name + "-rs-" + strconv.Itoa(rsIndex)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: rsName, Namespace: instance.Namespace}, replicaSet)

		if err != nil && errors.IsNotFound(err) {
			var rs = newReplicaSetForCR(instance, rsName)
			if err := controllerutil.SetControllerReference(instance, rs, r.scheme); err != nil {
				return reconcile.Result{}, err
			}
			err := r.client.Create(context.TODO(), rs)
			if err != nil {
				return reconcile.Result{}, err
			}
		} else {
			//reqLogger.Info("ReplicaSet " + rsName + " already exist, skip ...")
			podList := &corev1.PodList{}
			labelSelector := labels.SelectorFromSet(labelsForPod(instance, replicaSet))
			listOps := &client.ListOptions{
				Namespace:     replicaSet.Namespace,
				LabelSelector: labelSelector,
			}
			err = r.client.List(context.TODO(), listOps, podList)
			if err != nil {
				continue
			}
			podIps := getPodIps(podList.Items)
			for _, temp := range podIps {
				instances = append(instances, temp)
			}
			continue
		}
	}

	available := int32(0)
	current := int32(0)
	for rsIndex := 1; rsIndex <= instance.Spec.Size; rsIndex++ {
		found := &appsv1.ReplicaSet{}
		rsName := instance.Name + "-rs-" + strconv.Itoa(rsIndex)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: rsName, Namespace: instance.Namespace}, found)

		if err != nil && errors.IsNotFound(err) {
			continue
		}

		available += found.Status.AvailableReplicas
		current++
	}
	instance.Status.Desired = int32(instance.Spec.Size)
	instance.Status.Current = current
	instance.Status.Ready = available
	instance.Status.Instances = instances

	if instance.Status.Desired == instance.Status.Ready {
		instance.Status.ClusterStatus = "READY"
		// 判断redis集群是否OK，如果OK则设置为CLUSTER-OK，否则为CLUSTER-FAIL，或者CLUSTER-CREATING
	} else {
		instance.Status.ClusterStatus = "FAIL"
	}

	_ = r.client.Status().Update(context.TODO(), instance)

	return reconcile.Result{true, time.Duration(3) * time.Second}, nil
}

func labelsForPod(cr *redisoperatorv1alpha1.RedisCluster, rs *appsv1.ReplicaSet) map[string]string {
	return map[string]string{"app": cr.Name, "rsName": rs.Name}
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *redisoperatorv1alpha1.RedisCluster) *corev1.Pod {
	labels := map[string]string{
		"app": cr.Name,
	}
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name + "-pod",
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "busybox",
					Image:   "busybox",
					Command: []string{"sleep", "3600"},
				},
			},
		},
	}
}

func getPodIps(pods []corev1.Pod) []string {
	var podIps []string
	for _, pod := range pods {
		podIps = append(podIps, pod.Status.PodIP)
	}
	return podIps
}

func newReplicaSetForCR(cr *redisoperatorv1alpha1.RedisCluster, rsName string) *appsv1.ReplicaSet {
	labels := map[string]string{
		"app":    cr.Name,
		"rsName": rsName,
	}
	return &appsv1.ReplicaSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      rsName,
			Namespace: cr.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.ReplicaSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Name:      cr.Name + "-pod",
					Namespace: cr.Namespace,
					Labels:    labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "busybox",
							Image:   "busybox",
							Command: []string{"sleep", "3600"},
						},
					},
				},
			},
		},
	}
}
