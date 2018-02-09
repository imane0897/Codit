package main

import (
	"bufio"
	"archive/tar"
	// "bytes"
	"errors"
	"io"
	"os"
	"os/exec"
	"strconv"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"golang.org/x/net/context"
)

func main() {
	// compileFile(8, 1, 1000)
	runFile(8, 0, 1000)
}

func compileFile(rid int, ftype int, pid int) error {
	cmd := exec.Command("")
	switch ftype {
	case 0:
		cmd = exec.Command("gcc", "-o", strconv.Itoa(rid), strconv.Itoa(rid)+".c")
	case 1:
		cmd = exec.Command("gcc", "-o", strconv.Itoa(rid), strconv.Itoa(rid)+".cpp", "-lstdc++")
	}
	cmd.Dir = "../../filesystem/submissions"

	out, err := cmd.CombinedOutput()
	if err != nil {
		return errors.New(string(out))
	}

	return nil
}

func runFile(rid int, ftype int, pid int) error {
	ctx := context.Background()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:      "busybox",
		WorkingDir: "/usr/src/c",
		Cmd:        []string{"ls"},
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	file, err := os.Open("../../filesystem/submissions/" + strconv.Itoa(rid))
	if err != nil {
		panic(err)
	}
	if err := cli.CopyToContainer(ctx, resp.ID, "/usr/src/c", tar.NewReader(bufio.NewReader(file)), types.CopyToContainerOptions{}); err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	io.Copy(os.Stdout, out)

	return nil
}
