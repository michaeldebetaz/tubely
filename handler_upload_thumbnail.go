package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

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

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't get video from database", err)
		return
	}
	if video.UserID != userID {
		respondWithError(w, http.StatusForbidden, "You are not allowed to upload thumbnail for this video", nil)
		return
	}

	var maxMemory int64 = 10 << 20 // the same as 10 * 1024 * 1024 = 10 MB
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse form", err)
		return
	}

	thumbnailFile, thumbnailFileHeader, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't get thumbnail file", err)
		return
	}
	defer thumbnailFile.Close()

	contentType := thumbnailFileHeader.Header.Get("Content-Type")

	mediatype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Couldn't parse content type", err)
		return
	}

	mediaTypesExtensions := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
	}

	ext, ok := mediaTypesExtensions[mediatype]
	if !ok {
		respondWithError(w, http.StatusBadRequest, "Unsupported media type", nil)
		return
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate random bytes", err)
		return
	}
	randomStr := base64.RawURLEncoding.EncodeToString(bytes)

	fileName := randomStr + ext
	filePath := filepath.Join(cfg.assetsRoot, fileName)
	file, err := os.Create(filePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't create thumbnail file", err)
		return
	}
	defer file.Close()

	if _, err := io.Copy(file, thumbnailFile); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't copy thumbnail file", err)
		return
	}

	assetsURL := fmt.Sprintf("http://localhost:%s/assets/", cfg.port)
	thumbnailURL := assetsURL + fileName
	video.ThumbnailURL = &thumbnailURL

	if err := cfg.db.UpdateVideo(video); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't update video thumbnail in database", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}
