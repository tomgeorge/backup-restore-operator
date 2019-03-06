# Cluster Setup

I am using minishift/hostpath for local development.  I have tested the AWS functionality on an Openshift 4 cluster in AWS.  The hostpath functionality has been tested on a kubernetes 1.13 cluster in Digitalocean and in minishift/minikube.

All the commands use the `oc` command, but you can substitute `kubectl` with no issues.


Edit `example/snapshot-serviceaccount-clusterrole-rolebinding.yaml` and add your AWS access keys in the `awskeys` secret.

Get the base64 encoded values by echo-ing them:

`echo -n $AWS_ACCESS_KEY_ID | base64`

Regardless of storage strategy (hostPath or EBS), switch to the default project and create the Service Account, Cluster Role, and Cluster Rolebinding:

```shell
oc project default
oc apply -f example/snapshot-serviceaccount-clusterrole-rolebinding.yaml
oc apply -f example/storageclass.yaml
```

These commands:

+ creates the serviceaccount, role, and rolebinding for the snapshot controller.
+ creates the `snapshot-promoter` StorageClass object.

## Setting up AWS

```shell
oc apply -f example/aws/ebs-provisioner.yaml
oc apply -f example/aws/aws-snapshotter-deployment.yaml

<wait for snapshotter deployment pod to start>
```

These commands:

+ enables the AWS EBS dynamic storage provisioner
+ creates a snapshot provisioner and controller deployment in the default namespace

### Creating a Volumesnapshot and Restoring in AWS

```shell
oc apply -f example/aws/aws-pvc.yaml
oc apply -f example/aws/aws-snapshot.yaml
```

These commands:

+ create an example PVC, which is provisioned in EBS.  The PV is created under the hood by the dynamic storage provisioner.
+ creates a snapshot object which looks at that PVC.

After creating the snapshot, `oc get volumesnapshot` and `oc get volumesnapshotdata` should return results.  If you check the logs of the controller deployment, in the `snapshot-controller` container, you should see a message saying that your snapshot was created.

```shell
oc apply -f example/aws/aws-restore-claim.yaml
oc apply -f example/aws/aws-restore-pod.yaml
```

These commands:

+ Creates a PVC that is based on the snapshot you created in the above section
+ Creates a pod, which makes use of this snapshotted claim

## Hostpath

Per [this issue](https://github.com/kubernetes-incubator/external-storage/issues/1139), creating a Deployment object will not work for host path.  You can, however, run the snapshot controller locally, or on a cluster, as a binary.  

### Start the snapshot controller as a binary

```shell
cd $GOPATH/src/github.com
git clone https://github.com/kubernetes-incubator/external-storage/
cd snapshot
make all
_output/bin/snapshot-controller -kubecontig=/path/to/kubeconfig
```

### Create the pv and pvc

```shell
oc apply -f hostpath/pv.yaml
oc apply -f hostpath/pvc.yaml
```

### Add some test data to the hostpath directory

The PV we just created mounts to `/tmp/test` on the host.  The snapshot creation will fail if the directory is empty.

```shell
mkdir -p /tmp/test
echo "Hello world" > /tmp/test/data
```

### Create the snapshot

```shell
oc apply -f hostpath/snapshot.yaml
```

Check the controller logs, and wait for it to finish creating the snapshot.  Verify that the `VolumeSnapshot` and `VolumeSnapshotData` objects were created: 

```shell
oc get volumesnapshot,volumesnapshotdata -o yaml

items:
- apiVersion: volumesnapshot.external-storage.k8s.io/v1
  kind: VolumeSnapshot
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"volumesnapshot.external-storage.k8s.io/v1","kind":"VolumeSnapshot","metadata":{"annotations":{},"name":"snappyq","namespace":"default"},"spec":{"persistentVolumeClaimName":"hostpath-pvc"}}
    creationTimestamp: 2019-03-06T03:32:44Z
    generation: 1
    labels:
      SnapshotMetadata-PVName: hostpath-pv
      SnapshotMetadata-Timestamp: "1551843205421392891"
    name: snappyq
    namespace: default
    resourceVersion: "403070"
    selfLink: /apis/volumesnapshot.external-storage.k8s.io/v1/namespaces/default/volumesnapshots/snappyq
    uid: 813f1dc0-3fc0-11e9-b0a0-5254008a48f2
  spec:
    persistentVolumeClaimName: hostpath-pvc
    snapshotDataName: k8s-volume-snapshot-999305c3-3fc0-11e9-b28a-54e1add9c45d
  status:
    conditions:
    - lastTransitionTime: 2019-03-06T03:33:25Z
      message: Snapshot created successfully
      reason: ""
      status: "True"
      type: Ready
    creationTimestamp: null
- apiVersion: volumesnapshot.external-storage.k8s.io/v1
  kind: VolumeSnapshotData
  metadata:
    creationTimestamp: 2019-03-06T03:33:25Z
    generation: 1
    name: k8s-volume-snapshot-999305c3-3fc0-11e9-b28a-54e1add9c45d
    namespace: ""
    resourceVersion: "403069"
    selfLink: /apis/volumesnapshot.external-storage.k8s.io/v1/volumesnapshotdatas/k8s-volume-snapshot-999305c3-3fc0-11e9-b28a-54e1add9c45d
    uid: 99722028-3fc0-11e9-b0a0-5254008a48f2
  spec:
    hostPath:
      snapshot: /tmp/9991cadd-3fc0-11e9-b28a-54e1add9c45d.tgz
    persistentVolumeRef:
      kind: PersistentVolume
      name: hostpath-pv
    volumeSnapshotRef:
      kind: VolumeSnapshot
      name: default/snappyq-813f1dc0-3fc0-11e9-b0a0-5254008a48f2
  status:
    conditions:
    - lastTransitionTime: 2019-03-06T03:33:25Z
      message: Snapshot created successfully
      reason: ""
      status: "True"
      type: Ready
    creationTimestamp: null
kind: List
metadata:
  resourceVersion: ""
  selfLink: ""
```
