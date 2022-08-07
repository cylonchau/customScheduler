package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"github.com/cylonchau/customScheduler/pkg"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	command := app.NewSchedulerCommand(
		app.WithPlugin(pkg.Name, pkg.New))
	// 对于外部注册一个plugin
	// command := app.NewSchedulerCommand(
	// 	app.WithPlugin("example-plugin1", ExamplePlugin1.New))

	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
