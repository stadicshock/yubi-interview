package main

import (
	"log"

	"yubi-loan/workflow"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	// "github.com/temporalio/samples-go/retryactivity"
)

func main() {
	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "retry-activity", worker.Options{})

	w.RegisterWorkflow(workflow.LoanCreationWorkflow)
	w.RegisterActivity(workflow.SyncImageActivity)
	w.RegisterActivity(workflow.VirusCheck)
	w.RegisterActivity(workflow.OcrValidation)
	w.RegisterActivity(workflow.LoanEligibility)
	w.RegisterActivity(workflow.LenderLoadCreation)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("Unable to start worker", err)
	}
}
