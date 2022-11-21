package main

import (
	"flag"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/nicholasjackson/dagger-example/dagger/helper"
)

var ref = flag.String("ref", "dev", "tag or branch or sha")

func main() {
	flag.Parse()

	build, err := helper.NewBuild()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	apply(build)
	if build.HasError() {
		build.Logger.Error(build.LastError().Error())
		os.Exit(1)
	}
}

func apply(build *helper.Build) {

	app := buildApplication(build)

	fmt.Println(packageApplication(build, app, *ref))
}

func buildApplication(build *helper.Build) *dagger.File {
	done := build.LogStart("Application Build")
	defer done()

	src := build.Client.Host().Workdir()

	golang := build.Client.Container().From("golang:latest")
	golang = golang.WithMountedDirectory("/src", src).
		WithWorkdir("/src").
		WithEnvVariable("CGO_ENABLED", "0")

	golang = golang.Exec(
		dagger.ContainerExecOpts{
			Args: []string{"go", "build", "-o", "build/dagger-example"},
		},
	)

	return golang.Directory("./build").File("dagger-example")
}

func packageApplication(build *helper.Build, app *dagger.File, branch string) string {
	if build.Cancelled() {
		return ""
	}

	done := build.LogStart("Package Application")
	defer done()

	prodImage := build.Client.Container().From("alpine:latest")
	prodImage = prodImage.WithFS(
		prodImage.FS().WithFile("/bin/myapp",
			app,
		)).
		WithEntrypoint([]string{"/bin/myapp"})

	addr, err := prodImage.Publish(build.ContextWithTimeout(helper.DefaultTimeout), "praveen4g0/dagger-example:"+branch)

	if err != nil {
		build.LogError(fmt.Errorf("Error creating and pushing container: %s", err))
	}

	return addr
}
