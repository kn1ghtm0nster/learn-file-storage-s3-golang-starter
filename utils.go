package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type FFProbeJSON struct {
	Streams []struct {
		Width		 	   int    `json:"width"`
		Height			   int    `json:"height"`
	}
}

func getVideoAspectRatio(filePath string) (string, error) {
	var output bytes.Buffer
	var ffprobeData FFProbeJSON

	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	cmd.Stdout = &output

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	outputBytes := output.Bytes()

	err = json.Unmarshal(outputBytes, &ffprobeData)
	if err != nil {
		return "", err
	}

	if len(ffprobeData.Streams) == 0 {
		return "other", nil
	}

	width := ffprobeData.Streams[0].Width
	height := ffprobeData.Streams[0].Height
	
	if width == 0 || height == 0 {
		return "other", nil
	}

	aspectRatio := float64(width) / float64(height)

	if aspectRatio >= 1.77 && aspectRatio <= 1.79 {
		return "16:9", nil
	}

	if aspectRatio >= 0.55 && aspectRatio <= 0.57 {
		return "9:16", nil
	}

	return "other", nil
}

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	err := cmd.Run()
	if err != nil {
		return "", err
	}
	return outputPath, nil
}

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s3Client)
	presignedObject, err := presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}
	return presignedObject.URL, nil
}