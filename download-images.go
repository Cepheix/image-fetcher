package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/gosuri/uiprogress"
)

type DownloadWork struct {
	downloadFolder string
	work           <-chan string
	shutdown       <-chan bool
	wg             *sync.WaitGroup
	progressBar    *uiprogress.Bar
}

func downloadUrls(downloadWork DownloadWork) {
	for {
		select {
		case url := <-downloadWork.work:
			downloadUrl(url, downloadWork.downloadFolder)
			downloadWork.progressBar.Incr()
		case <-downloadWork.shutdown:
			downloadWork.wg.Done()
			return
		}
	}
}

func downloadUrl(url, downloadFolder string) {

	elements := []string{downloadFolder, buildFileName(url)}
	fullFilePath := strings.Join(elements, "/")

	if fileExists(fullFilePath) {
		return
	}

	response, responseError := http.Get(url)

	if responseError != nil {
		fmt.Println("Error during downling", url, responseError.Error())
	}

	defer response.Body.Close()

	file, fileError := os.Create(fullFilePath)

	if fileError != nil {
		fmt.Println("Error during creation of the file", fullFilePath, fileError.Error())
	}

	defer file.Close()

	_, saveError := io.Copy(file, response.Body)

	if saveError != nil {
		fmt.Println("Error during saving image", fullFilePath, fileError.Error())
	}
}

func buildFileName(url string) string {
	segments := strings.Split(url, "/")

	return segments[len(segments)-2] + "-" + segments[len(segments)-1]
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
