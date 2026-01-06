package main

import (
	"strings"
	"time"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)



func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, nil
	}

	bucketAndKey := strings.Split(*video.VideoURL, ",")
	if len(bucketAndKey) != 2 {
		return video, nil
	}
	
	bucket := bucketAndKey[0]
	key := bucketAndKey[1]

	signedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, 15 * time.Minute)
	if err != nil {
		return video, err
	}

	video.VideoURL = &signedURL
	return video, nil
}