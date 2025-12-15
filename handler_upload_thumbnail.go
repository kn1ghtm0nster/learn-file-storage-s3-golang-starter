package main

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}


	fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	// TODO: implement the upload here

	const maxMemory = 10 << 20 // 10 MB
	err = r.ParseMultipartForm(maxMemory)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse multipart form", err)
		return
	}

	thumbNailFile, headers, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get thumbnail file from form", err)
		return
	}

	contentType, _, err := mime.ParseMediaType(headers.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid Content-Type header", err)
		return
	}

	// if content type is not image/jpeg or image/png, return error
	if contentType != "image/jpeg" && contentType != "image/png"{
		respondWithError(w, http.StatusBadRequest, "Unsupported thumbnail content type", nil)
		return
	}

	fileType := strings.Split(contentType, "/")
	// map to a file extension png -> .png, jpeg -> .jpg etc.
	if len(fileType) != 2 || (fileType[0] != "image") {
		respondWithError(w, http.StatusBadRequest, "Invalid thumbnail file type", nil)
		return
	}
	extension := ""
	switch fileType[1] {
	case "png":
		extension = ".png"
	case "jpeg":
		extension = ".jpg"
	case "gif":
		extension = ".gif"
	default:
		respondWithError(w, http.StatusBadRequest, "Unsupported thumbnail file type", nil)
		return
	}

	filename := filepath.Join(cfg.assetsRoot, videoID.String() + extension)

	outFile, err := os.Create(filename)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create thumbnail file", err)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, thumbNailFile)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't save thumbnail file", err)
		return
	}

	videoData, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get video from database", err)
		return
	}

	if videoData.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "Unauthorized", nil)
		return
	}

	urlString := fmt.Sprintf("http://localhost:%s/assets/%s%s", cfg.port, videoID.String(), extension)
	videoData.ThumbnailURL = &urlString

	err = cfg.db.UpdateVideo(videoData)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video with thumbnail URL", err)
		return
	}
	
	respondWithJSON(w, http.StatusOK, videoData)
}
