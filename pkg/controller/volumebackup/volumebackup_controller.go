package volumebackup

import (
	"bytes"
	"context"
	"fmt"

	"github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	snapclientset "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned"
	backupsv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
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
	return &ReconcileVolumeBackup{client: mgr.GetClient(), config: mgr.GetConfig(), snapClientset: snapClientset, scheme: mgr.GetScheme()}
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
}

// Reconcile reads that state of the cluster for a VolumeBackup object and makes changes based on the state read
// and what is in the VolumeBackup.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVolumeBackup) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	//reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger := log
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
	foundDeployment := r.client.Get(context.TODO(), client.ObjectKey{
		Namespace: instance.ObjectMeta.Namespace,
		Name:      instance.Spec.ApplicationRef,
	}, deployment)

	if foundDeployment != nil {
		reqLogger.Info(fmt.Sprintf("Err object is not null %v", foundDeployment))
	} else {
	}

	pods := &corev1.PodList{}

	selector, err := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	if err != nil {
		reqLogger.Error(err, "Could not create selectors from label %v", deployment.Spec.Selector)
		return reconcile.Result{}, err
	}

	err = r.client.List(context.TODO(), &client.ListOptions{
		Namespace:     deployment.ObjectMeta.Namespace,
		LabelSelector: selector,
	}, pods)

	if err != nil {
		reqLogger.Error(err, "Could not list pods for deployment %v", deployment.Name)
		return reconcile.Result{}, err
	}

	for _, pod := range pods.Items {
		r.freezePod(&pod)
		err := r.issueBackup(instance, &pod)
		if err != nil {
			reqLogger.Error(err, "Error creating volumesnapshot")
			return reconcile.Result{}, err
		}
		// TODO: Do we want to defer unfreezing the pod? can we even defer it if we don't know the pod beforehand? Could probably make another function
		r.unfreezePod(&pod)
		reqLogger.Info(fmt.Sprintf("post doRemoteExec"))
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileVolumeBackup) issueBackup(instance *backupsv1alpha1.VolumeBackup, pod *corev1.Pod) error {
	snapshots := r.createVolumeSnapshotsFromPod(pod, instance.Spec.StorageClass)
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

func (r *ReconcileVolumeBackup) createVolumeSnapshotsFromPod(pod *corev1.Pod, storageClass string) []v1alpha1.VolumeSnapshot {
	snapshots := []v1alpha1.VolumeSnapshot{}
	for _, volume := range pod.Spec.Volumes {
		snapshot := r.createVolumeSnapshotFromPod(pod, storageClass, volume)
		snapshots = append(snapshots, *snapshot)
	}
	return snapshots
}

func (r *ReconcileVolumeBackup) createVolumeSnapshotFromPod(pod *corev1.Pod, storageClass string, volume corev1.Volume) *v1alpha1.VolumeSnapshot {
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
			VolumeSnapshotClassName: &storageClass,
		},
	}
	return snapshot
}

func (r *ReconcileVolumeBackup) unfreezePod(pod *corev1.Pod) int {
	postHook := pod.Annotations["backups.example.com.post-hook"]
	command := []string{"/bin/sh", "-i", "-c"}
	command = append(command, postHook)
	err := r.doRemoteExec(pod, command)
	return err
}

func (r *ReconcileVolumeBackup) freezePod(pod *corev1.Pod) int {
	preHook := pod.Annotations["backups.example.com.pre-hook"]
	command := []string{"/bin/sh", "-i", "-c"}
	command = append(command, preHook)
	err := r.doRemoteExec(pod, command)
	return err
}

func (r *ReconcileVolumeBackup) doRemoteExec(pod *corev1.Pod, command []string) int {
	cfg, err := kubernetes.NewForConfig(r.config)
	execRequest := cfg.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(pod.Namespace).
		Name(pod.Name).
		SubResource("exec").
		Param("container", pod.Spec.Containers[0].Name)
	execRequest.VersionedParams(&corev1.PodExecOptions{
		Container: pod.Spec.Containers[0].Name,
		Command:   command,
		Stdin:     false,
		Stdout:    true,
		Stderr:    true,
		TTY:       true,
	}, scheme.ParameterCodec)

	log.Info(fmt.Sprintf("Command is %v", command))
	log.Info(fmt.Sprintf("Running command %s on pod %s in container %s", command, pod.Name, pod.Spec.Containers[0].Name))

	exec, err := remotecommand.NewSPDYExecutor(r.config, "POST", execRequest.URL())
	if err != nil {
		log.Error(err, "Error exec-ing")
	}

	var stdOut, stdErr bytes.Buffer

	err = exec.Stream(remotecommand.StreamOptions{
		Stdout: &stdOut,
		Stderr: &stdErr,
		Tty:    true,
	})

	var exitCode int
	if err == nil {
		exitCode = 0
		fmt.Println(stdOut.String())
		fmt.Println(stdErr.String())
	} else {
		log.Error(nil, fmt.Sprintf("exit code is %d", exitCode))
		fmt.Println(stdOut.String())
		fmt.Println(stdErr.String())
		log.Error(err, "Error")
	}

	log.Info(fmt.Sprintf("Exit Code: %v", exitCode))
	if exitCode != 0 {
		exitCode = 2
	}
	return exitCode
}
