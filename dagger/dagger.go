package main

import (
	"context"
	"fmt"
	"os"
	"path"

	"dagger.io/dagger"
	"github.com/pkg/errors"
)

func main() {
	ctx := context.Background()

	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		fmt.Println(errors.Wrap(err, "error connecting to dagger engine"))
		os.Exit(1)
	}

	defer func() {
		err = client.Close()
		if err != nil {
			fmt.Println(errors.Wrap(err, "error while closing dagger client"))
			os.Exit(1)
		}
	}()

	src, err := client.Host().Workdir().Read().ID(ctx)
	if err != nil {
		fmt.Println(errors.Wrap(err, "error getting reference to host directory id"))
		os.Exit(1)
	}
	buildPath := "build/"

	golang := client.Container().From("golang:latest")
	golang = golang.WithMountedDirectory("/src", src).WithWorkdir("/src")

	golang = golang.WithEnvVariable("CGO_ENABLED", "0")
	golang = golang.Exec(dagger.ContainerExecOpts{
		Args: []string{"go", "build", "-o", buildPath},
	})

	if err != nil {
		panic(err)
	}

	build, err := golang.Directory(buildPath).ID(ctx)
	if err != nil {
		panic(err)
	}

	err = os.MkdirAll(path.Join(".", buildPath), os.ModePerm)
	if err != nil {
		panic(err)
	}

	workdir := client.Host().Workdir()
	_, err = workdir.Write(ctx, build, dagger.HostDirectoryWriteOpts{
		Path: buildPath,
	})
	if err != nil {
		panic(err)
	}

	_, err = client.Container().Build(src).Publish(ctx, "vahidmostofi/dagger-example:latest")
	if err != nil {
		panic(err)
	}

}
