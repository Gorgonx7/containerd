package containerd

import (
	"syscall"
	"testing"
)

func TestCheckpointRestore(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	client, err := New(address)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	var (
		ctx, cancel = testContext()
		id          = "CheckpointRestore"
	)
	defer cancel()

	image, err := client.GetImage(ctx, testImage)
	if err != nil {
		t.Error(err)
		return
	}
	spec, err := GenerateSpec(WithImageConfig(ctx, image), WithProcessArgs("sleep", "100"))
	if err != nil {
		t.Error(err)
		return
	}
	container, err := client.NewContainer(ctx, id, WithSpec(spec), WithImage(image), WithNewRootFS(id, image))
	if err != nil {
		t.Error(err)
		return
	}
	defer container.Delete(ctx)

	task, err := container.NewTask(ctx, empty())
	if err != nil {
		t.Error(err)
		return
	}
	defer task.Delete(ctx)

	statusC := make(chan uint32, 1)
	go func() {
		status, err := task.Wait(ctx)
		if err != nil {
			t.Error(err)
		}
		statusC <- status
	}()

	if err := task.Start(ctx); err != nil {
		t.Error(err)
		return
	}

	checkpoint, err := task.Checkpoint(ctx, WithExit)
	if err != nil {
		t.Error(err)
		return
	}

	<-statusC

	if _, err := task.Delete(ctx); err != nil {
		t.Error(err)
		return
	}
	if task, err = container.NewTask(ctx, empty(), WithTaskCheckpoint(checkpoint)); err != nil {
		t.Error(err)
		return
	}
	defer task.Delete(ctx)

	go func() {
		status, err := task.Wait(ctx)
		if err != nil {
			t.Error(err)
		}
		statusC <- status
	}()

	if err := task.Start(ctx); err != nil {
		t.Error(err)
		return
	}

	if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
		t.Error(err)
		return
	}
	<-statusC
}

func TestCheckpointRestoreNewContainer(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	client, err := New(address)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	const id = "CheckpointRestoreNewContainer"
	ctx, cancel := testContext()
	defer cancel()

	image, err := client.GetImage(ctx, testImage)
	if err != nil {
		t.Error(err)
		return
	}
	spec, err := GenerateSpec(WithImageConfig(ctx, image), WithProcessArgs("sleep", "100"))
	if err != nil {
		t.Error(err)
		return
	}
	container, err := client.NewContainer(ctx, id, WithSpec(spec), WithImage(image), WithNewRootFS(id, image))
	if err != nil {
		t.Error(err)
		return
	}
	defer container.Delete(ctx)

	task, err := container.NewTask(ctx, empty())
	if err != nil {
		t.Error(err)
		return
	}
	defer task.Delete(ctx)

	statusC := make(chan uint32, 1)
	go func() {
		status, err := task.Wait(ctx)
		if err != nil {
			t.Error(err)
		}
		statusC <- status
	}()

	if err := task.Start(ctx); err != nil {
		t.Error(err)
		return
	}

	checkpoint, err := task.Checkpoint(ctx, WithExit)
	if err != nil {
		t.Error(err)
		return
	}

	<-statusC

	if err := container.Delete(ctx); err != nil {
		t.Error(err)
		return
	}
	if _, err := task.Delete(ctx); err != nil {
		t.Error(err)
		return
	}
	if container, err = client.NewContainer(ctx, id, WithCheckpoint(checkpoint, id)); err != nil {
		t.Error(err)
		return
	}
	if task, err = container.NewTask(ctx, empty(), WithTaskCheckpoint(checkpoint)); err != nil {
		t.Error(err)
		return
	}
	defer task.Delete(ctx)

	go func() {
		status, err := task.Wait(ctx)
		if err != nil {
			t.Error(err)
		}
		statusC <- status
	}()

	if err := task.Start(ctx); err != nil {
		t.Error(err)
		return
	}

	if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
		t.Error(err)
		return
	}
	<-statusC
}
