kubectl apply -f snapshot-serviceaccount-clusterrole-rolebinding.yaml
kubectl apply -f storageclass.yaml
kubectl apply -f pv.yaml -f pvc.yaml
kubectl apply -f hostpath-snapshotter-deployment.yaml
