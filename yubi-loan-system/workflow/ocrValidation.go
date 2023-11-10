package workflow

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
)

func OcrValidation(ctx context.Context, loanID uint) error {
	logger := activity.GetLogger(ctx)

	if activity.HasHeartbeatDetails(ctx) {
		var completedIdx int
		if err := activity.GetHeartbeatDetails(ctx, &completedIdx); err == nil {
			logger.Info("Resuming from failed attempt", "ReportedProgress", completedIdx)
		}
	}

	OcrValidationServiceDown := rand.Intn(2) == 0
	if OcrValidationServiceDown {
		logger.Info("OcrValidation activity failed, will retry...")
		return temporal.NewApplicationError("some retryable error", "SomeType")
	}
	time.Sleep(1 * time.Second)

	logger.Info("OcrValidation Activity succeed." + fmt.Sprint(loanID))
	return nil
}
