package executor

import (
	"bytes"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	scheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("pod_executor")

type PodExecutor interface {
	DoRemoteExec(pod *corev1.Pod, command []string) (int, error)
}

type RemotePodExecutor struct {
	config *rest.Config
}

type FakePodExecutor struct {
}

func CreateNewRemotePodExecutor(config *rest.Config) PodExecutor {
	return &RemotePodExecutor{
		config: config,
	}
}

func CreateNewFakePodExecutor() PodExecutor {
	return &FakePodExecutor{}
}

func (executor *FakePodExecutor) DoRemoteExec(pod *corev1.Pod, command []string) (int, error) {
	return 0, nil
}

func (executor *RemotePodExecutor) DoRemoteExec(pod *corev1.Pod, command []string) (int, error) {
	clientSet, err := kubernetes.NewForConfig(executor.config)
	if err != nil {
		return -1, err
	}
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

	exec, err := remotecommand.NewSPDYExecutor(executor.config, "POST", execRequest.URL())
	if err != nil {
		return -1, err
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
	} else {
		return -1, err
	}

	// log.Info(fmt.Sprintf("Exit Code: %v", exitCode))
	if exitCode != 0 {
		exitCode = 2
	}
	return exitCode, nil
}
