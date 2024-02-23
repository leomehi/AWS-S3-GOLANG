package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MyEvent struct {
	Data string `json:"data"`
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var event MyEvent

	// Parse JSON from the request body
	if err := json.Unmarshal([]byte(request.Body), &event); err != nil {
		log.Printf("failed to parse request body: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusBadRequest}, nil
	}

	// Generate a unique bucket name based on the current timestamp
	bucketName := fmt.Sprintf("your-prefix-%d", time.Now().Unix())

	// Use /tmp directory for writing the file (writable in Lambda environment)
	fileName := "/tmp/your-file-name.json"
	file, err := os.Create(fileName)
	if err != nil {
		log.Printf("failed to create file: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}
	defer file.Close()

	// Marshal JSON data and write to file
	jsonData, err := json.Marshal(event)
	log.Printf("Received data: %+v", event)
	if err != nil {
		log.Printf("failed to marshal JSON: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	_, err = file.Write(jsonData)
	if err != nil {
		log.Printf("failed to write to file: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Specify the AWS region explicitly
	awsRegion := "ap-south-1"

	// Initialize AWS SDK configuration
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(awsRegion))
	if err != nil {
		log.Printf("failed to load AWS config: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg)

	// Create S3 bucket
	_, err = s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucketName),
	})
	if err != nil {
		log.Printf("failed to create S3 bucket: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Upload file to S3 bucket
	_, err = s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String(fileName),
		Body:   bytes.NewReader(jsonData),
	})
	if err != nil {
		log.Printf("failed to upload file to S3: %v", err)
		return events.APIGatewayProxyResponse{StatusCode: http.StatusInternalServerError}, nil
	}

	// Return a success response
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Body:       "Data successfully processed and stored in S3.",
	}, nil
}

func main() {
	lambda.Start(handler)
}
