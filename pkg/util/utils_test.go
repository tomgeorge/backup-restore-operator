package util

import (
	v1alpha1 "github.com/tomgeorge/backup-restore-operator/pkg/apis/backups/v1alpha1"
	testHelpers "github.com/tomgeorge/backup-restore-operator/pkg/controller/test/helpers"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	fakeClient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

type fixture struct {
	objs     []runtime.Object
	cr       *v1alpha1.VolumeBackup
	expected string
}

func CreateFixture() *fixture {
	f := &fixture{}
	f.objs = []runtime.Object{}
	return f
}

func TestGetVolumeSnapshotFromCR(t *testing.T) {
	type testCase struct {
		fixture *fixture
	}

	testCases := map[string]testCase{
		"app=app, vol=test-claim-1, snap=app-test-claim-1": {
			fixture: &fixture{
				objs: []runtime.Object{
					testHelpers.NewPod("default", "app-pod", 1),
				},
				expected: "app-test-claim-1",
				cr:       testHelpers.NewVolumeBackup("default", "backup", "app", "busybox", "volume", testHelpers.StatusEmpty),
			},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(*testing.T) {
			fixture := testCase.fixture
			utils := &BackupRestoreUtils{
				client: fakeClient.NewFakeClientWithScheme(scheme.Scheme, fixture.objs...),
				cfg:    &rest.Config{},
			}
			pod, ok := fixture.objs[0].(*corev1.Pod)
			if !ok {
				t.Errorf("Make the first argument in the fixture a pod")
			}
			actual := utils.GetVolumeSnapshotFromCR(fixture.cr, pod)
			if fixture.expected != actual {
				t.Errorf("Expected %v but was %v", fixture.expected, actual)
			}
		})
	}
}
