package main

import (
	"log/slog"
	"minio-example/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	// routers
	r.GET("/", func(ctx *gin.Context) {
		// Call service to handle the request
		ctx.JSON(http.StatusOK, gin.H{"message": "minio-go-example`"})
	})

	// end point to upload the files
	r.POST("/upload-files", func(ctx *gin.Context) {
		// Call service to handle the request
		result := service.UploadFiles(ctx)
		ctx.JSON(result.Status_code, result)
	})

	// end point to get the presigned urls of uploaded files
	r.GET("/get-presigned-urls", func(ctx *gin.Context) {
		// Call service to handle the request
		result := service.PreSignedURLs(ctx)
		ctx.JSON(result.Status_code, result)
	})

	slog.Info("Starting server...")
	// start the server
	r.Run()
}
