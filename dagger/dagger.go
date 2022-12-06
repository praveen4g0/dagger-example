package main

import (
	"flag"
	"fmt"
	"os"

	"dagger.io/dagger"
	"github.com/hashicorp/go-hclog"
	"github.com/nicholasjackson/dagger-example/dagger/helper"
)

var ref = flag.String("ref", "dev", "tag or branch or sha")
var registry = flag.String("reg", "quay.io", "provide image registry domain, defaults to quay.io")
var repository = flag.String("user", "praveen4g0", "proide image registry username")
var image_name = flag.String("image", "dagger-example", "Provide image name")

func main() {
	flag.Parse()

	b, err := helper.NewBuild()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	build(b)
}

func build(build *helper.Build) {
	app := buildApplication(build)
	build.Logger.Log(hclog.Debug, packageApplication(build, app, *registry, *repository, *image_name, *ref))
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

func packageApplication(build *helper.Build, app *dagger.File, registry, repository, image_name, branch string) string {
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

	addr, err := prodImage.Publish(build.ContextWithTimeout(helper.DefaultTimeout), registry+"/"+repository+"/"+image_name+":"+branch)

	if err != nil {
		build.LogError(fmt.Errorf("Error creating and pushing container: %s", err))
	}

	return addr
}
