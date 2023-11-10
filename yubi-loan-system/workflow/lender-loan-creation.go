package workflow

import (
	"context"
	"fmt"
	"math/rand"
	"time"
	"yubi-loan/model"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

func LenderLoadCreation(ctx context.Context, loanID uint) error {
	logger := activity.GetLogger(ctx)

	if activity.HasHeartbeatDetails(ctx) {
		var completedIdx int
		if err := activity.GetHeartbeatDetails(ctx, &completedIdx); err == nil {
			logger.Info("Resuming from failed attempt", "ReportedProgress", completedIdx)
		}
	}

	LenderLoadCreationServiceDown := rand.Intn(2) == 0
	if LenderLoadCreationServiceDown {
		logger.Info("LenderLoadCreation activity failed, will retry...")
		return temporal.NewApplicationError("some retryable error", "SomeType")
	}
	time.Sleep(1 * time.Second)

	var loan model.Loan
	db, err := gorm.Open("postgres", "dbname=postgres sslmode=disable")
	if err != nil {
		return err
	}
	defer db.Close()

	if err := db.Where("id = ?", loanID).First(&loan).Error; err != nil {
		return err
	}

	loadAccpetedFromLender := rand.Intn(2) == 0

	if loadAccpetedFromLender {
		loan.Status = "Accepted"
	} else {
		loan.Status = "Rejected"
	}
	db.Save(&loan)
	logger.Info("LenderLoadCreation Activity succeed." + fmt.Sprint(loanID))
	return nil
}
