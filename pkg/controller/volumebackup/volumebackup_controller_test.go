package volumebackup

import (
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	core "k8s.io/client-go/testing"
)

func int32Pointer(n int32) *int32 {
	return &n
}

func TestReconcile(t *testing.T) {
	var (
		name                 = "example-backup"
		namespace            = "example"
		applicationRef       = "example-application-to-backup"
		replicas       int32 = 1
	)

	cases := []testCase{
		{
			name: "reconcile add",
			objs: []runtime.Object{
				newDeployment(namespace, applicationRef, &replicas),
				newPod(namespace, applicationRef),
			},
			snapshotObjs: []runtime.Object{},
			volumeBackup: newVolumeBackup(namespace, name, applicationRef),
			expectedActions: []core.Action{
				core.ActionImpl{
					Namespace: namespace,
					Verb:      "create",
					Resource: schema.GroupVersionResource{
						Group:    "snapshots",
						Resource: "volumesnapshots",
						Version:  "v1alpha1",
					},
				},
			},
		},
	}
	testSuite(t, cases)
}
