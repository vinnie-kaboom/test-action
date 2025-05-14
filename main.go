package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

func main() {
	bucketName := os.Getenv("INPUT_BUCKET_NAME")
	projectID := os.Getenv("INPUT_PROJECT_ID")
	creds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")

	if bucketName == "" || projectID == "" {
		log.Fatalf("bucket_name and project_id are required inputs")
	}

	ctx := context.Background()
	var client *storage.Client
	var err error

	if creds != "" {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(creds))
	} else {
		client, err = storage.NewClient(ctx)
	}
	if err != nil {
		log.Fatalf("Failed to create storage client: %v", err)
	}
	defer client.Close()

	bucket := client.Bucket(bucketName)
	_, err = bucket.Attrs(ctx)
	if err == nil {
		fmt.Printf("Bucket already exists: %s\n", bucketName)
		return
	}

	if err := bucket.Create(ctx, projectID, nil); err != nil {
		log.Fatalf("Failed to create bucket: %v", err)
	}
	fmt.Printf("Bucket created: %s\n", bucketName)
}
