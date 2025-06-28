package main

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var (
	s3Client     *s3.Client
	bucketName   = "my-app-file-uploads-aarush-2025" // change this to your actual bucket name
	bucketRegion = "us-east-2"                       // change this to your bucket's region
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(bucketRegion))
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}
	s3Client = s3.NewFromConfig(cfg)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// Limit file size (10 MB here)
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "File too large or invalid form", http.StatusBadRequest)
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "File not found in form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	key := fmt.Sprintf("uploads/%d_%s", time.Now().Unix(), fileHeader.Filename)

	uploadInput := &s3.PutObjectInput{
		Bucket:        aws.String(bucketName),
		Key:           aws.String(key),
		Body:          file.(multipart.File),
		ContentLength: aws.Int64(fileHeader.Size),
		ContentType:   aws.String(fileHeader.Header.Get("Content-Type")),
	}

	_, err = s3Client.PutObject(context.TODO(), uploadInput)
	if err != nil {
		http.Error(w, "Upload failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucketName, bucketRegion, key)
	fmt.Fprintf(w, "✅ File uploaded successfully: %s\n", url)
}

func main() {
	http.HandleFunc("/upload", uploadHandler)

	port := "8080"
	fmt.Println("🚀 Server running at http://localhost:" + port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
