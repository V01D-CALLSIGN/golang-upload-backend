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
	// ✅ Added for ACL
)

var (
	s3Client     *s3.Client
	bucketName   = "my-app-file-uploads-aarush-2025"
	bucketRegion = "us-east-2"
)

func init() {
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(bucketRegion))
	if err != nil {
		log.Fatalf("unable to load AWS SDK config, %v", err)
	}
	s3Client = s3.NewFromConfig(cfg)
}

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	// 🔧 Add CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*") // for dev; lock down later
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// 🔧 Handle preflight request
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 🚫 Limit file size
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
		log.Println("Upload failed:", err)
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
