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
	credsJSON := os.Getenv("INPUT_CREDENTIALS_JSON")

	if bucketName == "" || projectID == "" || credsJSON == "" {
		log.Fatalf("bucket_name, project_id, and credentials_json are required inputs")
	}

	ctx := context.Background()

	credOption := option.WithCredentialsJSON([]byte(credsJSON))

	client, err := storage.NewClient(ctx, credOption)
	if err != nil {
		log.Fatalf("Failed to create storage client: %v", err)
	}
	defer func(client *storage.Client) {
		err := client.Close()
		if err != nil {
			fmt.Printf("Failed to close storage client: %v", err)
		}
	}(client)

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
