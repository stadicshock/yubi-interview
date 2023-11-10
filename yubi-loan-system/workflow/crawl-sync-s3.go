package workflow

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"yubi-loan/model"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.temporal.io/sdk/activity"

	_ "github.com/jinzhu/gorm/dialects/postgres"
)

func SyncImageActivity(ctx context.Context, loanID uint) error {
	logger := activity.GetLogger(ctx)

	if activity.HasHeartbeatDetails(ctx) {
		// we are retry from a failed attempt, and there is reported progress that we should resume from.
		var completedIdx int
		if err := activity.GetHeartbeatDetails(ctx, &completedIdx); err == nil {
			logger.Info("Resuming from failed attempt", "ReportedProgress", completedIdx)
		}
	}

	syncToS3(loanID)
	logger.Info("Activity succeed.")
	return nil
}

const (
	endpoint  = "localhost:9000" // Your Minio server endpoint
	accessKey = "ROOTUSER"       // Your Minio access key
	secretKey = "ROOTUSER123"    // Your Minio secret key
	bucket    = "images"         // Your Minio bucket name
)

func syncToS3(loadID uint) {
	var loan model.Loan
	db, err := gorm.Open("postgres", "dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal("1", err)
	}
	defer db.Close()

	if err := db.Where("id = ?", loadID).First(&loan).Error; err != nil {
		return
	}

	// Initialize Minio client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		fmt.Println("Error initializing Minio client:", err)
		return
	}

	// Create a bucket if it doesn't exist
	exists, err := minioClient.BucketExists(context.Background(), bucket)
	if err != nil {
		fmt.Println("Error checking bucket existence:", err)
		return
	}
	if !exists {
		err = minioClient.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{})
		if err != nil {
			fmt.Println("Error creating bucket:", err)
			return
		}
		fmt.Println("Bucket created successfully:", bucket)
	}
	localURL, err := upload(minioClient, loan.DocURL)
	if err != nil {
		fmt.Printf("Error syncing image from %s: %v\n", loan.DocURL, err)
	} else {
		fmt.Printf("Image synced successfully. Local URL: %s\n", localURL)
	}
	// }
	loan.S3URL = localURL
	db.Save(&loan)
}

func upload(minioClient *minio.Client, imageURL string) (string, error) {
	// Get the image data
	resp, err := http.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Extract the image file name from the URL
	tokens := strings.Split(imageURL, "/")
	imageFileName := generateUUID() + tokens[len(tokens)-1]

	// Upload the image to Minio
	_, err = minioClient.PutObject(
		context.Background(),
		bucket,
		imageFileName,
		resp.Body,
		resp.ContentLength,
		minio.PutObjectOptions{ContentType: "application/octet-stream"},
	)

	if err != nil {
		return "", err
	}

	// Return the local Minio URL
	localURL := fmt.Sprintf("http://%s/%s/%s", endpoint, bucket, imageFileName)
	return localURL, nil
}

func generateUUID() string {
	uuidObj := uuid.New()
	return uuidObj.String()
}
