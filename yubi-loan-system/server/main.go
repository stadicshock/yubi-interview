package main

import (
	"context"
	"log"
	"yubi-loan/model"
	"yubi-loan/workflow"

	"net/http"

	"github.com/pborman/uuid"
	"go.temporal.io/sdk/client"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

var db *gorm.DB
var err error

func main() {

	db, err = gorm.Open("postgres", "dbname=postgres sslmode=disable")
	if err != nil {
		log.Fatal("1", err)
	}
	defer db.Close()

	db.AutoMigrate(&model.Loan{})

	// Initialize the Gin router
	router := gin.Default()

	// Endpoint to create a new loan
	router.POST("/loan", createLoan)

	// Endpoint to get the status of a loan by loanID
	router.GET("/loan/status/:loanID", getLoanStatus)

	// Run the server on port 8080
	router.Run(":80")
}

// Handler function for creating a new loan
func createLoan(c *gin.Context) {
	var inputLoan model.Loan

	// Bind JSON input to the Loan struct
	if err := c.ShouldBindJSON(&inputLoan); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	inputLoan.Status = "inProgress"

	// Save the loan to the database
	db.Create(&inputLoan)

	startWorkFlow(inputLoan.ID) // triggers workflow in background
	c.JSON(http.StatusOK, gin.H{"status": "Loan created successfully"})

}

// Handler function for getting the status of a loan by loanID
func getLoanStatus(c *gin.Context) {
	loanID := c.Param("loanID")

	var loan model.Loan
	if err := db.Where("id = ?", loanID).First(&loan).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Loan not found"})
		return
	}

	// You can customize the response based on your specific loan status logic
	c.JSON(http.StatusOK, gin.H{"status": loan.Status})
}

func startWorkFlow(loanID uint) {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{
		HostPort: client.DefaultHostPort,
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		ID:        "retry_activity_" + uuid.New(),
		TaskQueue: "retry-activity",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, workflow.LoanCreationWorkflow, loanID)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}
	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())
}
