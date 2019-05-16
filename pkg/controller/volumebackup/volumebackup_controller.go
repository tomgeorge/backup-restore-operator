package volumebackup

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	backupsv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
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

type Writer struct {
	Str []string
}

func (w *Writer) Write(p []byte) (n int, err error) {
	str := string(p)
	if len(str) > 0 {
		w.Str = append(w.Str, str)
	}
	return len(str), nil
}

//func parseCommand(command string) (command string, args []string) {
//	parts := strings.Split(command, " ")
//	outCommand := parts[0]
//	outArgs := []string{}
//	append(parts[1:], outArgs)
//	append
//	return outCommand, outArgs
//}

func newStringReader(ss []string) io.Reader {
	formattedString := strings.Join(ss, " ")
	reader := strings.NewReader(formattedString)
	return reader
}

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
	return &ReconcileVolumeBackup{client: mgr.GetClient(), scheme: mgr.GetScheme()}
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
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
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
	client client.Client
	scheme *runtime.Scheme
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

	selector, _ := metav1.LabelSelectorAsSelector(deployment.Spec.Selector)
	err = r.client.List(context.TODO(), &client.ListOptions{
		Namespace:     deployment.ObjectMeta.Namespace,
		LabelSelector: selector,
	}, pods)

	if err != nil {
		reqLogger.Info("err is not nil")
	}

	kubeClient, kubeConfig, err := newClient()
	if err != nil {
		reqLogger.Error(err, "Error constructing kube client")
	}

	for _, pod := range pods.Items {
		freezePod(kubeClient, kubeConfig, pod, deployment)
		issueBackup()
		unfreezePod(kubeClient, kubeConfig, pod, deployment)
		reqLogger.Info(fmt.Sprintf("post doRemoteExec"))
	}

	return reconcile.Result{}, err
	// Define a new Pod object
	pod := newPodForCR(instance)

	// Set VolumeBackup instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Pod already exists

	found := &corev1.Pod{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.client.Create(context.TODO(), pod)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Pod created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Pod already exists - don't requeue
	reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	return reconcile.Result{}, nil
}

func issueBackup() {
	log.Info(fmt.Sprintf("issued a backup"))
}

func unfreezePod(clientSet *kubernetes.Clientset, config *rest.Config, pod corev1.Pod, deployment *appsv1.Deployment) int {
	fmt.Printf("unfreeze the pod")
	postHook := pod.Annotations["backups.example.com.post-hook"]
	command := []string{"/bin/sh", "-i", "-c"}
	command = append(command, postHook)
	err := doRemoteExec(clientSet, config, command, pod)
	return err
}

func freezePod(clientSet *kubernetes.Clientset, config *rest.Config, pod corev1.Pod, deployment *appsv1.Deployment) int {
	preHook := pod.Annotations["backups.example.com.pre-hook"]
	command := []string{"/bin/sh", "-i", "-c"}
	command = append(command, preHook)
	err := doRemoteExec(clientSet, config, command, pod)
	return err
}

func doRemoteExec(clientSet *kubernetes.Clientset, config *rest.Config, command []string, pod corev1.Pod) int {

	execRequest := clientSet.CoreV1().RESTClient().Post().
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

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", execRequest.URL())
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
		log.Info("err is nil")
		fmt.Println(stdOut.String())
		fmt.Println(stdErr.String())
	} else {
		log.Error(nil, fmt.Sprintf("exit code is %d", exitCode))
		fmt.Println(stdOut.String())
		fmt.Println(stdErr.String())
		log.Error(err, "Error")
	}

	fmt.Sprintf("Exit Code: %v", exitCode)
	if exitCode != 0 {
		exitCode = 2
	}
	return exitCode
}

// newPodForCR returns a busybox pod with the same name/namespace as the cr
func newPodForCR(cr *backupsv1alpha1.VolumeBackup) *corev1.Pod {
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

func newClient() (*kubernetes.Clientset, *rest.Config, error) {
	kubeConfig, err := clientcmd.BuildConfigFromFlags("", filepath.Join(os.Getenv("HOME"), clientcmd.RecommendedHomeDir, clientcmd.RecommendedFileName))
	if err != nil {
		return nil, nil, err
	}

	clientSet := kubernetes.NewForConfigOrDie(kubeConfig)
	return clientSet, kubeConfig, err
}
