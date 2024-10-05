package service

import (
	"log/slog"
	"minio-example/model"
	"minio-example/utils"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
)

func UploadFiles(ctx *gin.Context) model.ResponseModel {

	var response model.ResponseModel
	// Create a minio-client
	client := utils.ConnectToMinIo()
	bucket_name := "minio-go-example-bucket"

	// Create a bucket to upload files
	err := client.MakeBucket(ctx, bucket_name, minio.MakeBucketOptions{Region: "us-east-1"}) // Create a bucket in default location
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := client.BucketExists(ctx, bucket_name)
		if errBucketExists == nil && exists {
			slog.Info("We already own ", "bucket name", bucket_name)
		} else {
			slog.Error(err.Error())
		}
	} else {
		slog.Info("Successfully created ", "bucket name", bucket_name)
	}

	// Get the files from multipart header
	form, err := ctx.MultipartForm()
	if err != nil {
		slog.Error("Error while reading files from payload")

		// Build reponse model
		response.Status = false
		response.Status_code = 400
		response.Error = err.Error()
		response.Data = nil
		return response
	}

	// get the files
	files := form.File["files"]

	// Create a channel with length of files
	ch := make(chan model.Uploadstatus, len(files))

	// variable to store the fileuplod status
	var FileUploadStatus []model.Uploadstatus

	// Upload into bucket
	for _, file := range files {
		// Fire each goroutines to upload files into bucket
		// Set WaitGroup to wait till it ends
		// After data upload each goroutines update their status in the structure
		// to protect from Race condition we can use either mutex or channels
		// Im using channels to collect the result without race condition

		// wg.Add(1)
		go utils.PutImageInBucket(ctx, bucket_name, file, client, FileUploadStatus, ch)
	}
	// wait for the goroutines
	// wg.Wait()

	// Run a channel to collect the result
	for i := 0; i < len(files); i++ {
		FileUploadStatus = append(FileUploadStatus, <-ch)
	}

	slog.Info("Collected all results from channel")

	response.Status = true
	response.Status_code = 200
	response.Error = nil
	response.Data = FileUploadStatus
	return response
}

func PreSignedURLs(ctx *gin.Context) model.ResponseModel {
	var response model.ResponseModel
	// Create a minio-client
	client := utils.ConnectToMinIo()
	bucket_name := "minio-go-example-bucket"

	res := utils.GeneratePresignedURL(ctx, client, bucket_name)

	// build response
	response.Data = res
	response.Status = true
	response.Status_code = 200
	response.Error = nil

	return response
}
