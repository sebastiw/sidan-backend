package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type FileHandler struct {
}

type File struct {
	Filename string `json:"filename"`
}

func fileExtension(imageType string) (string, error) {
	switch imageType {
	case "image/gif":
		return "gif", nil
	case "image/png":
		return "png", nil
	case "image/jpeg":
		return "jpeg", nil
	}
	return "", errors.New("unknown image type")
}

func CheckError(w http.ResponseWriter, r *http.Request, err error) {
	if err != nil {
		log.Println(getRequestId(r), err)
		panic(err.Error())
	}
}

func (fh FileHandler) createImageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files. (bitshift 10 in decimal 20 times)
	err := r.ParseMultipartForm(10 << 20)
	CheckError(w, r, err)

	file, handler, err := r.FormFile("data")
	CheckError(w, r, err)
	defer file.Close()

	// DetectContentType take only first 512 bytes into consideration
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	CheckError(w, r, err)
	// Reset seek
	file.Seek(0, 0)

	contentType := http.DetectContentType(buff)
	fileExt, err := fileExtension(contentType)
	CheckError(w, r, err)

	tempFilename := fmt.Sprintf("upload-*.%s", fileExt)
	tempFile, err := ioutil.TempFile("static", tempFilename)
	CheckError(w, r, err)
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	CheckError(w, r, err)

	tempFile.Write(fileBytes)
	bareFilename := strings.TrimPrefix(tempFile.Name(), "static/")
	size := fmt.Sprintf("%+vb", handler.Size)
	log.Println(getRequestId(r), "Uploaded", handler.Filename, size, bareFilename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(File{Filename: bareFilename})
}
