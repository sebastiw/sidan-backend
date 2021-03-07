package router

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

type FileHandler struct {
}

type File struct {
	Filename string `json:filename`
}

func file_extension(image_type string) string {
	switch image_type {
	case "image/gif":
		return "gif"
	case "image/png":
		return "png"
	case "image/jpeg":
		return "jpeg"
	}
	return ""
}

func (fh FileHandler) createImageHandler(w http.ResponseWriter, r *http.Request) {
	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files. (bitshift 10 in decimal 20 times)
	r.ParseMultipartForm(10 << 20)

	file, handler, err := r.FormFile("data")
	if err != nil {
		log.Println(get_request_id(r), "Error Retrieving the File", err)
		return
	}
	defer file.Close()

	// DetectContentType take only first 512 bytes into consideration
	buff := make([]byte, 512)
	_, err = file.Read(buff)
	if err != nil {
		fmt.Println(err)
		return
	}

	contentType := http.DetectContentType(buff)
	fileExt := file_extension(contentType)
	if "" == fileExt {
		log.Println(get_request_id(r), "Not an image file")
		return
	}

	tempFilename := fmt.Sprintf("upload-*.%s", fileExt)
	tempFile, err := ioutil.TempFile("static", tempFilename)
	if err != nil {
		log.Println(get_request_id(r), err)
	}
	defer tempFile.Close()

	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		log.Println(get_request_id(r), err)
	}
	tempFile.Write(fileBytes)
	bareFilename := strings.TrimPrefix(tempFile.Name(), "static/")
	size := fmt.Sprintf("%+vb", handler.Size)
	log.Println(get_request_id(r), "Uploaded", handler.Filename, size, bareFilename)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(File{Filename: bareFilename})
}
