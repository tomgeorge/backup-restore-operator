# Backup and Restore Kubernetes Operator

An operator to easily back up the data of a stateful application in Kubernetes.  Uses [CSI](https://kubernetes-csi.github.io/docs/) for kubernetes-native backup management. 



### Contents

1. [Prerequesites](#prereqs)
2. [Installing the Operator](#install)
3. [Creating a Backup](#create)

### Prerequisites <a name="prereqs"></a>

- Kubernetes 1.12+, with the `VolumeSnapshotDataSource=true` feature gate set on the API server.  See [the documentation](https://kubernetes.io/docs/reference/command-line-tools-reference/feature-gates/) to learn how to enable a feature gate on a kubernetes component.
- A CSI storage driver intsalled in the cluster.  See the [list of drivers](https://kubernetes-csi.github.io/docs/drivers.html) for a list of options.



### Installing the Operator <a name="install"></a>

`kubectl apply -f deploy/`

This will create:

- A backup-restore-operator `ServiceAccount` in the kube-system namespace
- `ClusterRole` and `CluserRoleBinding` for the service account to perform necessary API operations
- A backup-restore-operator `Deployment` that runs the controllers and watches for `VolumeBackup` requests.
- `VolumeBackup` and `VolumeBackupProvider` CRDs.


Verify that the operator is running:

```
λ ~/go/src/github.com/tomgeorge/backup-restore-operator/ master* kubectl get pods -n kube-system | grep backup
backup-restore-operator-7c7c89d976-msx5z      1/1     Running   0          25s
```

Also check the logs to make sure you don't see any errors.

### Creating a Backup <a name="create"></a>

+ Create a deployment with a `pre-hook` and `post-hook` to quiesce the application.  The following is an example mysql deployment with a volume and hooks to freeze/unfreeze the application:

  ```yaml
  apiVersion: apps/v1
  kind: Deployment
  metadata:
    name: mysql
    labels:
      app: mysql-persistent
      template: mysql-persistent-template
  spec:
    replicas: 1
    selector:
      matchLabels:
        name: mysql
    template:
      metadata:
        labels:
          name: mysql
        annotations:
          backups.example.com.pre-hook: "mysql -h 127.0.0.1 --user=root --password=$MYSQL_ROOT_PASSWORD --database=$MYSQL_DATABASE -e 'flush tables with read lock;'"
          backups.example.com.post-hook: "mysql -h 127.0.0.1 --user=root --password=$MYSQL_ROOT_PASSWORD --database=$MYSQL_DATABASE -e 'unlock tables;'"
      spec:
        containers:
        - env:
          - name: MYSQL_USER
            valueFrom:
              secretKeyRef:
                key: database-user
                name: mysql
          - name: MYSQL_PASSWORD
            valueFrom:
              secretKeyRef:
                key: database-password
                name: mysql
          - name: MYSQL_ROOT_PASSWORD
            valueFrom:
              secretKeyRef:
                key: database-root-password
                name: mysql
          - name: MYSQL_DATABASE
            valueFrom:
              secretKeyRef:
                key: database-name
                name: mysql
          image: mysql
          imagePullPolicy: IfNotPresent
          livenessProbe:
            failureThreshold: 3
            initialDelaySeconds: 30
            periodSeconds: 10
            successThreshold: 1
            tcpSocket:
              port: 3306
            timeoutSeconds: 1
          name: mysql
          ports:
          - containerPort: 3306
            protocol: TCP
          readinessProbe:
            exec:
              command:
              - /bin/sh
              - -i
              - -c
              - MYSQL_PWD="$MYSQL_PASSWORD" mysql -h 127.0.0.1 -u $MYSQL_USER -D $MYSQL_DATABASE
                -e 'SELECT 1'
            failureThreshold: 3
            initialDelaySeconds: 5
            periodSeconds: 10
            successThreshold: 1
            timeoutSeconds: 1
          resources:
            limits:
              memory: 512Mi
          terminationMessagePath: /dev/termination-log
          terminationMessagePolicy: File
          volumeMounts:
          - mountPath: /mysql-data
            name: mysql-data
            subPath: mysql-data
        dnsPolicy: ClusterFirst
        restartPolicy: Always
        schedulerName: default-scheduler
        securityContext: {}
        terminationGracePeriodSeconds: 30
        volumes:
        - name: mysql-data
          persistentVolumeClaim:
            claimName: mysql
  ```

  + Create a VolumeBackup object with a reference to the application:
  ```yaml
  apiVersion: backups.example.com/v1alpha1
  kind: VolumeBackup
  metadata:
    name: example-volumebackup
  spec:
    applicationRef: mysql
  ```

You should now see a `VolumeBackup` and it's accompanying `VolumeSnapshot` and `VolumeSnapshotContent` objects:

`kubectl get volumebackups,volumesnapshot,volumesnapshotcontents -o yaml`



Backup Flow

Given a request for backup of a pvc:
Find all the pods that use that pvc
Call freeze
Use snapshot.external-storage.k8s.io api to create snapshot
Watch VolumeSnapshot object until it has an id
Unfreeze the pods
Watch the VolumeSnapshot to wait for the upload to complete (get some ID that corresponds to the uploaded data)


### Restore Flow
Given a Restore object that references a Deployment, a PersistentVolumeClaim, and a VolumeSnapshotData object:
Create a PVC with the snapshot-promoter storage class
Update the deployment with the newly created PVC
Application restarts
Optional: Delete the old PVC


Subsequent deployments of an application that has been restored could remove the previously restored data
