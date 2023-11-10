package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

func LoanCreationWorkflow(ctx workflow.Context, loanID uint) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		HeartbeatTimeout:    60 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    10,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Sync image to s3
	err := workflow.ExecuteActivity(ctx, SyncImageActivity, loanID).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", "Error", err)
		return err
	}

	// check for virus
	err = workflow.ExecuteActivity(ctx, VirusCheck, loanID).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", "Error", err)
		return err
	}

	// OCR validation
	err = workflow.ExecuteActivity(ctx, OcrValidation, loanID).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", "Error", err)
		return err
	}

	// Loan eligilibity check
	err = workflow.ExecuteActivity(ctx, LoanEligibility, loanID).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", "Error", err)
		return err
	}

	// Lender loan creation
	err = workflow.ExecuteActivity(ctx, LenderLoadCreation, loanID).Get(ctx, nil)
	if err != nil {
		workflow.GetLogger(ctx).Info("Workflow completed with error.", "Error", err)
		return err
	}

	workflow.GetLogger(ctx).Info("Workflow completed.")
	return nil
}
