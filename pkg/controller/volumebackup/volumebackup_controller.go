package volumebackup

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kubernetes-csi/external-snapshotter/pkg/apis/volumesnapshot/v1alpha1"
	snapclientset "github.com/kubernetes-csi/external-snapshotter/pkg/client/clientset/versioned"
	backupsv1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	"github.com/tomgeorge/backup-restore-operator/pkg/util"
	"github.com/tomgeorge/backup-restore-operator/pkg/util/executor"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	return &ReconcileVolumeBackup{
		client:        mgr.GetClient(),
		snapClientset: snapClientset,
		config:        mgr.GetConfig(),
		scheme:        mgr.GetScheme(),
		executor:      executor.CreateNewRemotePodExecutor(mgr.GetConfig()),
	}
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
	reqLogger.Info("Reconciling VolumeBackup")
	defer reqLogger.Info("End Reconcile Loop")

	// Fetch the VolumeBackup instance
	instance := &backupsv1alpha1.VolumeBackup{}

	util := &util.BackupRestoreUtils{
		Client: r.client,
		Cfg:    r.config,
	}

	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if kubeErrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	reqLogger.Info(fmt.Sprintf("Deployment to grab hook from is %v in namespace %v",
		instance.Spec.ApplicationName,
		instance.ObjectMeta.Namespace))

	// TODO: Change ApplicationName to just a pod selector
	deployment := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), client.ObjectKey{
		Namespace: instance.ObjectMeta.Namespace,
		Name:      instance.Spec.ApplicationName,
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
	err = r.client.List(context.TODO(), pods, &client.ListOptions{
		Namespace:     deployment.ObjectMeta.Namespace,
		LabelSelector: selector,
	})

	if err != nil {
		reqLogger.Error(err, "Could not list pods for deployment %v", deployment.Name)
		return reconcile.Result{}, err
	}

	if len(pods.Items) == 0 {
		reqLogger.Info("Deployment to back up has no replicas", "deployment.Name", deployment.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	// Use the first pod in the list to do the freeze
	// TODO: Document that we are performing the backup on pod 0 of the returned list of pods
	podToExec := pods.Items[0]
	reqLogger.Info("got a pod", "Pod.Name", podToExec.Name)
	if podToExec.Status.Phase != corev1.PodRunning {
		reqLogger.Info("Pod is not in the running phase", "Pod.Name", podToExec.Name)
		return reconcile.Result{Requeue: true}, nil
	}

	// Get the container referenced by the backup request
	var containerToExec corev1.Container
	var containerStatus corev1.ContainerStatus
	for idx, container := range podToExec.Spec.Containers {
		if container.Name == instance.Spec.ContainerName {
			containerToExec = container
			containerStatus = podToExec.Status.ContainerStatuses[idx]
		}
	}

	// If we can't match a container in the deployment to one specified by the VolumeBackupSpec, return an error
	if &containerToExec == nil || &containerStatus == nil {
		err = errors.New("Could not locate container to exec")
		reqLogger.Error(err, "Could not locate container to run exec", "Container.Name", instance.Spec.ContainerName)
		return reconcile.Result{}, err
	}

	if !containerStatus.Ready {
		reqLogger.Info("Container is not yet ready", "Container.Name", instance.Spec.ContainerName)
		return reconcile.Result{Requeue: true}, nil
	}

	if !instance.IsFrozen() && !instance.IsSnapshotIssued() {
		if err := r.freeze(&podToExec); err != nil {
			reqLogger.Error(err, "Error freezing pod", "Pod.Name", podToExec.Name)
			instance.UpdateStatus(backupsv1alpha1.PodFrozen, backupsv1alpha1.ConditionFalse, "ErrorFreezing", err.Error())
			updateErr := r.client.Status().Update(context.TODO(), instance)
			if updateErr != nil {
				reqLogger.Error(updateErr, "Unable to update the status of VolumeBackup", "Name", instance.Name)
				return reconcile.Result{}, updateErr
			}
			return reconcile.Result{}, err
		}
		instance.UpdateStatus(backupsv1alpha1.PodFrozen, backupsv1alpha1.ConditionTrue, "FreezeSuccessful", fmt.Sprintf("Froze Pod %v", podToExec.Name))
		instance.Status.PodPhase = backupsv1alpha1.PhaseFrozen
		reqLogger.Info("Updating status", "Name", instance.Name)
		updateerr := r.client.Status().Update(context.TODO(), instance)
		if updateerr != nil {
			reqLogger.Error(updateerr, "Unable to update the status of VolumeBackup", "Name", instance.Name)
			return reconcile.Result{}, err
		}

		// VolumeBackup has been updated with the frozen status, so return
		return reconcile.Result{}, nil
	}

	if !instance.IsSnapshotIssued() {
		if err = r.issueSnapshot(instance, deployment, &podToExec); err != nil {
			if !kubeErrors.IsAlreadyExists(err) {
				reqLogger.Error(err, "Error creating volumesnapshot")
				return reconcile.Result{}, err
			} else {
				reqLogger.Info("Backup already exists")
			}
		}
		instance.UpdateStatus(backupsv1alpha1.SnapshotIssued, backupsv1alpha1.ConditionTrue, "Created snapshot", "Created snapshot")
		updateErr := r.client.Status().Update(context.TODO(), instance)
		if updateErr != nil {
			reqLogger.Error(updateErr, "Unable to update the status of the VolumeBackup", "Name", instance.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if !instance.IsSnapshotCreated() {
		snapshotName := fmt.Sprintf("%v-%v", deployment.Name, instance.Spec.VolumeName)
		snapshot, err := r.snapClientset.SnapshotV1alpha1().VolumeSnapshots(instance.Namespace).Get(snapshotName, metav1.GetOptions{})
		if err != nil {
			reqLogger.Error(err, "Error getting volumesnapshot")
			return reconcile.Result{
				Requeue: true,
			}, err
		}
		if snapshot.Status.CreationTime != nil {
			instance.UpdateStatus(backupsv1alpha1.SnapshotCreated, backupsv1alpha1.ConditionTrue, "SnapshotReadyToUse", "Finished creating Volume Snapshot")
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				reqLogger.Error(err, "Error updating the status of the VolumeBackup", "Name", instance.Name)
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		} else {
			return reconcile.Result{RequeueAfter: 60 * time.Second}, nil
		}
	}

	if instance.IsPodFrozen() {
		reqLogger.Info("Unfreezing pod", "Pod.Name", podToExec.Name)
		if err = r.unfreeze(&podToExec); err != nil {
			reqLogger.Error(err, "Error un-freezing pod", "Pod.Name", podToExec.Name)
			return reconcile.Result{}, err
		}
		instance.Status.PodPhase = backupsv1alpha1.PhaseUnfrozen
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Error updating the status of the VolumeBackup", "Name", instance.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	storageAPIGroup := v1alpha1.GroupName

	if !instance.IsSnapshotUploading() {
		restorePVC := &corev1.PersistentVolumeClaim{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name + "-restore",
				Namespace: instance.Namespace,
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				Resources: corev1.ResourceRequirements{
					Requests: corev1.ResourceList{"storage": *resource.NewQuantity(1111111110, resource.BinarySI)},
				},
				DataSource: &corev1.TypedLocalObjectReference{
					APIGroup: &storageAPIGroup,
					Kind:     "VolumeSnapshot",
					Name:     util.GetVolumeSnapshotFromCR(instance, &podToExec),
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					"ReadWriteOnce",
				},
			},
		}

		isController := true
		restorePVC.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: v1alpha1.GroupName,
				Kind:       instance.Kind,
				Name:       instance.Name,
				UID:        instance.GetUID(),
				Controller: &isController,
			},
		})

		err = r.client.Create(context.TODO(), restorePVC)
		if err != nil {
			reqLogger.Error(err, "Error creating PVC based off of VolumeSnapshot")
			return reconcile.Result{}, err
		}

		provider, err := r.GetBackupProviderFromBackup(instance)
		if err != nil {
			reqLogger.Error(err, "Could not get secret from provider name")
			return reconcile.Result{}, err
		}

		uploaderPod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      instance.Name + "-uploader",
				Namespace: instance.Namespace,
			},
			Spec: corev1.PodSpec{
				Volumes: []corev1.Volume{
					{
						Name: "volume-to-backup",
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: restorePVC.GetName()},
						},
					},
				},
				RestartPolicy: corev1.RestartPolicyOnFailure,
				Containers: []corev1.Container{
					{
						Name:            "s3-uploader",
						ImagePullPolicy: corev1.PullAlways,
						Image:           "docker.io/tomgeorge/do-s3-uploader:latest",
						Env: []corev1.EnvVar{
							{
								Name:  "SPACES_BUCKET_NAME",
								Value: instance.Spec.BucketName,
							},
							{
								Name:  "SPACES_ENDPOINT",
								Value: instance.Spec.BucketEndpoint,
							},
							{
								Name:  "TAR_DIR",
								Value: "/etc/backup",
							},
							{
								Name: "SPACES_KEY",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: provider.Spec.SecretName,
										},
										Key: "s3Key",
									},
								},
							},
							{
								Name: "SPACES_SECRET",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: provider.Spec.SecretName,
										},
										Key: "s3SecretKey",
									},
								},
							},
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "volume-to-backup",
								MountPath: "/etc/backup",
							},
						},
					},
				},
			},
		}

		uploaderPod.SetOwnerReferences([]metav1.OwnerReference{
			{
				APIVersion: instance.APIVersion,
				UID:        instance.UID,
				Kind:       "VolumeBackup",
				Name:       instance.Name,
				Controller: &isController,
			},
		})

		if err = r.client.Create(context.TODO(), uploaderPod); err != nil {
			reqLogger.Error(err, "Could not create uploader pod")
			return reconcile.Result{}, err
		}

		instance.UpdateStatus(backupsv1alpha1.SnapshotUploading, backupsv1alpha1.ConditionTrue, "SnapshotUploading", "Uploader pod created")
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Error updating the status of the VolumeBackup", "Name", instance.Name)
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	if instance.Status.Phase == backupsv1alpha1.SnapshotUploading && !instance.IsPodFrozen() {
		uploaderPod := &corev1.Pod{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name + "-uploader"}, uploaderPod); err != nil {
			reqLogger.Error(err, "Could not get uploader pod")
			return reconcile.Result{}, err
		}
		if uploaderPod.Status.Phase == corev1.PodSucceeded {
			instance.UpdateStatus(backupsv1alpha1.SnapshotUploaded, backupsv1alpha1.ConditionTrue, "SnapshotUploadedToObjectStore", "Upload to object storesucceeded")
			if err = r.client.Status().Update(context.TODO(), instance); err != nil {
				reqLogger.Error(err, "Could not update volumebackup status")
				return reconcile.Result{}, err
			}
			return reconcile.Result{}, nil
		}
	}

	if instance.Status.Phase == backupsv1alpha1.SnapshotUploaded && !instance.IsPodFrozen() {
		reqLogger.Info("Deleting uploader pod")
		uploaderPod := &corev1.Pod{}
		if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name + "-uploader"}, uploaderPod); err != nil {
			reqLogger.Error(err, "Could not get uploader pod")
			return reconcile.Result{}, err
		}
		if err := r.client.Delete(context.TODO(), uploaderPod); err != nil {
			reqLogger.Error(err, "Could not delete uploader pod")
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	// TODO: Check VolumeSnapshot.Status Creation date before unfreezing
	// TODO: Check if the pod is unfrozen before unfreezing.  If the creation date exists, and the pod is still frozen, then unfreeze

	// TODO: update the status of the VolumeBackup, that the pod was unfrozen and return
	// TODO: check if the VolumeSnapshot is ready to use

	// TODO: Once the VolumeSnapshot is ready to use, upload it to an object store
	// - Create a PVC from the snapshot
	// - Create a `Job` that mounts the new PVC and does the upload

	return reconcile.Result{}, nil
}

// issueSnapshot finds the correct volume in the pod to backup according to the VolumeBackup spec, then calls
// createVolumeSnapshotFromPod and requestCreate
func (r *ReconcileVolumeBackup) issueSnapshot(instance *backupsv1alpha1.VolumeBackup, deployment *appsv1.Deployment, pod *corev1.Pod) error {
	volumeToBackup := &corev1.Volume{}
	volumes := pod.Spec.Volumes
	for _, vol := range pod.Spec.Volumes {
		if vol.VolumeSource.PersistentVolumeClaim != nil && vol.VolumeSource.PersistentVolumeClaim.ClaimName == instance.Spec.VolumeName {
			volumeToBackup = vol.DeepCopy()
		}
	}
	for i := 0; i < len(volumes); i++ {
		if volumes[i].VolumeSource.PersistentVolumeClaim != nil &&
			volumes[i].VolumeSource.PersistentVolumeClaim.ClaimName == instance.Spec.VolumeName {
			volumeToBackup = volumes[i].DeepCopy()
		}
	}
	snapshot, err := r.createVolumeSnapshotFromPod(deployment, pod, volumeToBackup)
	if err != nil {
		return err
	}
	if err := r.requestCreate(snapshot, instance); err != nil {
		return err
	}
	return nil
}

// createVolumeSnapshotFromPod takes a deployment, a pod, and a volume and creates a VolumeSnapshot of the form deploymentName-claimName
func (r *ReconcileVolumeBackup) createVolumeSnapshotFromPod(parentDeployment *appsv1.Deployment, pod *corev1.Pod, volume *corev1.Volume) (*v1alpha1.VolumeSnapshot, error) {
	claim := &corev1.PersistentVolumeClaim{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Namespace: pod.Namespace,
		Name:      volume.VolumeSource.PersistentVolumeClaim.ClaimName,
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
			Name:      fmt.Sprintf("%v-%v", parentDeployment.Name, volume.PersistentVolumeClaim.ClaimName),
			Namespace: pod.Namespace,
		},
		Spec: v1alpha1.VolumeSnapshotSpec{
			Source: &corev1.TypedLocalObjectReference{
				Name: volume.PersistentVolumeClaim.ClaimName,
				Kind: "PersistentVolumeClaim",
			},
			// TODO: Add a new field to VolumeBackup specifying the VolumeSnapshotClass
			VolumeSnapshotClassName: &pv.Spec.StorageClassName,
		},
	}
	return snapshot, nil
}

// requestCreate sends the create request to the API server
func (r *ReconcileVolumeBackup) requestCreate(snapshot *v1alpha1.VolumeSnapshot, instance *backupsv1alpha1.VolumeBackup) error {
	if err := controllerutil.SetControllerReference(instance, snapshot, r.scheme); err != nil {
		log.Error(err, "Unable to set owner reference of %v", snapshot.Name)
		return err
	}

	_, err := r.snapClientset.SnapshotV1alpha1().VolumeSnapshots(instance.Namespace).Create(snapshot)
	if err != nil {
		log.Error(err, "Error creating VolumeSnapshot")
		return err
	}
	return nil
}

// TODO: PodExecutor assumes the zeroth container, should pass in the container name instead
func (r *ReconcileVolumeBackup) unfreeze(pod *corev1.Pod) error {
	postHook := pod.Annotations["backups.example.com.post-hook"]
	command := []string{"/bin/sh", "-i", "-c"}
	command = append(command, postHook)
	_, err := r.executor.DoRemoteExec(pod, command)
	return err
}

func (r *ReconcileVolumeBackup) freeze(pod *corev1.Pod) error {
	preHook := pod.Annotations["backups.example.com.pre-hook"]
	command := []string{"/bin/sh", "-i", "-c"}
	command = append(command, preHook)
	_, err := r.executor.DoRemoteExec(pod, command)
	return err
}

func (r *ReconcileVolumeBackup) GetSecretFromProviderName(cr *backupsv1alpha1.VolumeBackup) (*corev1.Secret, error) {
	provider := &backupsv1alpha1.VolumeBackupProvider{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: cr.Namespace, Name: cr.Spec.BackupProviderName}, provider); err != nil {
		log.Error(err, "Could not get VolumeBackupProvider")
		return nil, err
	}

	secret := &corev1.Secret{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: cr.Namespace, Name: provider.Spec.SecretName}, secret); err != nil {
		log.Error(err, "Could not get S3 secret")
		return nil, err
	}
	return secret, nil
}

func (r *ReconcileVolumeBackup) GetBackupProviderFromBackup(cr *backupsv1alpha1.VolumeBackup) (*backupsv1alpha1.VolumeBackupProvider, error) {
	provider := &backupsv1alpha1.VolumeBackupProvider{}
	if err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: cr.Namespace, Name: cr.Spec.BackupProviderName}, provider); err != nil {
		log.Error(err, "Could not get VolumeBackupProvider")
		return nil, err
	}
	return provider, nil
}
