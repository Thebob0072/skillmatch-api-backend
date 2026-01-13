package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/storage"
)

// GCS client and bucket name (global variables)
var (
	gcsClient     *storage.Client
	gcsBucketName string
)

// initGCS initializes the Google Cloud Storage client
func initGCS(ctx context.Context) error {
	// Check if GOOGLE_APPLICATION_CREDENTIALS is set
	credPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS")
	if credPath == "" {
		return fmt.Errorf("GOOGLE_APPLICATION_CREDENTIALS environment variable not set")
	}

	// Check if credentials file exists
	if _, err := os.Stat(credPath); os.IsNotExist(err) {
		return fmt.Errorf("GCS credentials file not found: %s", credPath)
	}

	// Get bucket name from environment
	gcsBucketName = os.Getenv("GCS_BUCKET_NAME")
	if gcsBucketName == "" {
		return fmt.Errorf("GCS_BUCKET_NAME environment variable not set")
	}

	// Create GCS client
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create GCS client: %w", err)
	}

	// Test bucket access
	bucket := client.Bucket(gcsBucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		client.Close()
		return fmt.Errorf("failed to access GCS bucket '%s': %w", gcsBucketName, err)
	}

	gcsClient = client
	log.Printf("✅ GCS initialized successfully (bucket: %s)", gcsBucketName)
	return nil
}

// isGCSEnabled checks if GCS is available
func isGCSEnabled() bool {
	return gcsClient != nil
}

// getGCSClient returns the GCS client (safe to call even if GCS is disabled)
func getGCSClient() *storage.Client {
	return gcsClient
}

// getGCSBucketName returns the GCS bucket name
func getGCSBucketName() string {
	return gcsBucketName
}

// closeGCS closes the GCS client
func closeGCS() {
	if gcsClient != nil {
		gcsClient.Close()
		log.Println("✅ GCS client closed")
	}
}
