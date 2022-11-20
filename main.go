package main

import (
	"context"
	"dagger.io/dagger"
	"fmt"
	"os"
)

func main() {
	if err := build(context.Background()); err != nil {
		fmt.Println(err)
	}
}

func build(ctx context.Context) error {
	fmt.Println("Building with Dagger")

	// define build matrix
	oses := []string{"linux", "darwin"}
	arches := []string{"amd64", "arm64"}
	goVersions := []string{"1.18", "1.19.2"}

	// initialize Dagger client
	client, err := dagger.Connect(ctx, dagger.WithLogOutput(os.Stdout))
	if err != nil {
		return err
	}
	defer client.Close()

	// get reference to the local project
	src := client.Host().Workdir()

	// create empty directory to put build outputs
	outputs := client.Directory()

	for _, goVersion := range goVersions {
		imageTag := fmt.Sprintf("golang:%v", goVersion)

		// get `golang` image
		golang := client.Container().From(imageTag)

		// mount cloned repository into `golang` image
		golang = golang.WithMountedDirectory("/src", src).WithWorkdir("/src")

		for _, goos := range oses {
			for _, goarch := range arches {
				// create a directory for each os and arch
				path := fmt.Sprintf("build/%s/%s/", goos, goarch)

				// set GOARCH and GOOS in the build environment
				build := golang.WithEnvVariable("GOOS", goos)
				build = build.WithEnvVariable("GOARCH", goarch)

				// build application
				build = build.Exec(dagger.ContainerExecOpts{
					Args: []string{"go", "build", "-o", path},
				})

				// get reference to build output directory in container
				outputs = outputs.WithDirectory(path, build.Directory(path))
			}
		}
	}

	// write contents of container build/ directory to the host
	_, err = outputs.Export(ctx, ".")
	if err != nil {
		return err
	}
	return nil
}
