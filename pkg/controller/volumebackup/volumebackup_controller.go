package volumebackup

import (
	"context"
	"fmt"

	"github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	snapclientset "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned"
	backupsv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	"github.com/tomgeorge/backup-restore-operator/pkg/util/executor"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_volumebackup")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new VolumeBackup Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	snapClientset, err := snapclientset.NewForConfig(mgr.GetConfig())
	if err != nil {
		panic(err)
	}
	return &ReconcileVolumeBackup{client: mgr.GetClient(), config: mgr.GetConfig(), snapClientset: snapClientset, scheme: mgr.GetScheme(), executor: executor.CreateNewRemotePodExecutor(mgr.GetConfig())}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("volumebackup-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource VolumeBackup
	err = c.Watch(&source.Kind{Type: &backupsv1alpha1.VolumeBackup{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner VolumeBackup
	err = c.Watch(&source.Kind{Type: &v1alpha1.VolumeSnapshot{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &backupsv1alpha1.VolumeBackup{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileVolumeBackup{}

// ReconcileVolumeBackup reconciles a VolumeBackup object
type ReconcileVolumeBackup struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client        client.Client
	config        *rest.Config
	snapClientset snapclientset.Interface
	scheme        *runtime.Scheme
	executor      executor.PodExecutor
}

// Reconcile reads that state of the cluster for a VolumeBackup object and makes changes based on the state read
// and what is in the VolumeBackup.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVolumeBackup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	//reqLogger := log
	reqLogger.Info("Reconciling VolumeBackup")

	// Fetch the VolumeBackup instance
	instance := &backupsv1alpha1.VolumeBackup{}
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
	reqLogger.Info(fmt.Sprintf("Deployment to grab hook from is %v in namespace %v", instance.Spec.ApplicationRef, instance.ObjectMeta.Namespace))

	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), client.ObjectKey{
		Namespace: instance.ObjectMeta.Namespace,
		Name:      instance.Spec.ApplicationRef,
	}, deployment)

	if err != nil {
		reqLogger.Info(fmt.Sprintf("Error getting deployment: %v", err))
		return reconcile.Result{}, err
	}

	pods := &corev1.PodList{}

	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		reqLogger.Error(err, "Could not create selectors from label %v", deployment.Spec.Selector)
		return reconcile.Result{}, err
	}

	//TODO: What happens if the pod is in a bad state (error, crashloop, etc.)?
	//TODO: maybe simulate this in the e2e stuff
	err = r.client.List(context.TODO(), &client.ListOptions{
		Namespace:     deployment.ObjectMeta.Namespace,
		LabelSelector: selector,
	}, pods)

	if err != nil {
		reqLogger.Error(err, "Could not list pods for deployment %v", deployment.Name)
		return reconcile.Result{}, err
	}

	for _, pod := range pods.Items {
		if err := r.freezePod(&pod); err != nil {
			reqLogger.Error(err, "Error freezing pod: %v", pod.Name)
			return reconcile.Result{}, err
		}
		if err = r.issueBackup(instance, &pod); err != nil {
			reqLogger.Error(err, "Error creating volumesnapshot")
			return reconcile.Result{}, err
		}
		// TODO: Do we want to defer unfreezing the pod? can we even defer it if we don't know the pod beforehand? Could probably make another function
		// TODO: Check VolumeSnapshot.Status.ReadyToUse before unfreezing
		if err = r.unfreezePod(&pod); err != nil {
			reqLogger.Error(err, "Error un-freezing pod: %v", pod.Name)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileVolumeBackup) issueBackup(instance *backupsv1alpha1.VolumeBackup, pod *corev1.Pod) error {
	snapshots, err := r.createVolumeSnapshotsFromPod(pod)
	if err != nil {
		return err
	}
	for i, _ := range snapshots {
		snapshot := &snapshots[i]
		if err := r.requestCreate(snapshot, instance); err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileVolumeBackup) requestCreate(snapshot *v1alpha1.VolumeSnapshot, instance *backupsv1alpha1.VolumeBackup) error {
	if err := controllerutil.SetControllerReference(instance, snapshot, r.scheme); err != nil {
		log.Error(err, "Unable to set owner reference of %v", snapshot.Name)
		return err
	}
	_, err := r.snapClientset.VolumesnapshotV1alpha1().VolumeSnapshots(instance.Namespace).Create(snapshot)
	if err != nil {
		log.Error(err, "Error creating VolumeSnapshot")
		return err
	}
	return nil
}

func (r *ReconcileVolumeBackup) createVolumeSnapshotsFromPod(pod *corev1.Pod) ([]v1alpha1.VolumeSnapshot, error) {
	snapshots := []v1alpha1.VolumeSnapshot{}
	for _, volume := range pod.Spec.Volumes {
		if volume.VolumeSource.PersistentVolumeClaim != nil {
			snapshot, err := r.createVolumeSnapshotFromPod(pod, volume)
			if err != nil {
				return nil, err
			}
			snapshots = append(snapshots, *snapshot)
		}
	}
	return snapshots, nil
}

func (r *ReconcileVolumeBackup) createVolumeSnapshotFromPod(pod *corev1.Pod, volume corev1.Volume) (*v1alpha1.VolumeSnapshot, error) {
	claim := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      volume.PersistentVolumeClaim.ClaimName,
	}, claim)
	if err != nil {
		return nil, err
	}
	pv := &corev1.PersistentVolume{}
	err = r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: "",
		Name:      claim.Spec.VolumeName,
	}, pv)
	if err != nil {
		return nil, err
	}

	snapshot := &v1alpha1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%v-%v", pod.ObjectMeta.OwnerReferences[0].Name, volume.Name),
			Namespace: pod.Namespace,
		},
		Spec: v1alpha1.VolumeSnapshotSpec{
			Source: &corev1.TypedLocalObjectReference{
				Name: volume.PersistentVolumeClaim.ClaimName,
				Kind: "PersistentVolumeClaim",
			},
			// TODO: look up volume from claim name and get storage class from it
			VolumeSnapshotClassName: &pv.Spec.StorageClassName,
		},
	}
	return snapshot, nil
}

func (r *ReconcileVolumeBackup) unfreezePod(pod *corev1.Pod) error {
	postHook := pod.Annotations["backups.example.com.post-hook"]
	command := []string{"/bin/sh", "-i", "-c"}
	command = append(command, postHook)
	_, err := r.executor.DoRemoteExec(pod, command)
	return err
}

func (r *ReconcileVolumeBackup) freezePod(pod *corev1.Pod) error {
	preHook := pod.Annotations["backups.example.com.pre-hook"]
	command := []string{"/bin/sh", "-i", "-c"}
	command = append(command, preHook)
	_, err := r.executor.DoRemoteExec(pod, command)
	return err
}
