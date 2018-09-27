package k8s

import (
	"context"
	"strings"
	"testing"
	"time"

	"k8s.io/api/core/v1"
)

func TestWaitForContainerAlreadyAlive(t *testing.T) {
	f := newClientTestFixture(t)

	nt := MustParseNamedTagged(blorgDevImgStr)
	podData := fakePod(expectedPod, blorgDevImgStr)
	podData.Status = v1.PodStatus{
		ContainerStatuses: []v1.ContainerStatus{
			{
				Image: nt.String(),
				Ready: true,
			},
		},
	}
	f.addObject(&podData)

	pod, err := f.client.PodWithImage(f.ctx, nt, DefaultNamespace)
	if err != nil {
		f.t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(f.ctx, time.Second)
	defer cancel()

	err = WaitForContainerReady(ctx, f.client, pod, nt)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitForContainerSuccess(t *testing.T) {
	f := newClientTestFixture(t)
	f.addObject(&fakePodList)

	nt := MustParseNamedTagged(blorgDevImgStr)
	pod, err := f.client.PodWithImage(f.ctx, nt, DefaultNamespace)
	if err != nil {
		f.t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(f.ctx, time.Second)
	defer cancel()

	result := make(chan error)
	go func() {
		err := WaitForContainerReady(ctx, f.client, pod, nt)
		result <- err
	}()

	newPod := fakePod(expectedPod, blorgDevImgStr)
	newPod.Status = v1.PodStatus{
		ContainerStatuses: []v1.ContainerStatus{
			{
				Image: nt.String(),
				Ready: true,
			},
		},
	}

	<-f.watchNotify
	f.updatePod(&newPod)
	err = <-result
	if err != nil {
		t.Fatal(err)
	}
}

func TestWaitForContainerFailure(t *testing.T) {
	f := newClientTestFixture(t)
	f.addObject(&fakePodList)

	nt := MustParseNamedTagged(blorgDevImgStr)
	pod, err := f.client.PodWithImage(f.ctx, nt, DefaultNamespace)
	if err != nil {
		f.t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(f.ctx, time.Second)
	defer cancel()

	result := make(chan error)
	go func() {
		err := WaitForContainerReady(ctx, f.client, pod, nt)
		result <- err
	}()

	newPod := fakePod(expectedPod, blorgDevImgStr)
	newPod.Status = v1.PodStatus{
		ContainerStatuses: []v1.ContainerStatus{
			{
				Image: nt.String(),
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{},
				},
			},
		},
	}

	<-f.watchNotify
	f.updatePod(&newPod)
	err = <-result

	expected := "Container will never be ready"
	if err == nil || !strings.Contains(err.Error(), expected) {
		t.Fatalf("Expected error %q, actual: %v", expected, err)
	}
}