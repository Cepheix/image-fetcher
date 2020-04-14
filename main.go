package main

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	SourceFile        string `short:"s" long:"source-file" description:"Source file containing all image urls"`
	DestinationFolder string `short:"d" long:"destination-folder" description:"Destination folder where to put all images"`
}

func main() {
	var options Options
	_, optionsErr := flags.Parse(&options)

	if optionsErr != nil {
		panic(optionsErr)
	}

	fmt.Println(options)

	urls, readingError := readLines(options.SourceFile)

	if readingError != nil {
		panic(readingError)
	}

	work := make(chan string)
	shutdown := make(chan bool)
	pool := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(pool)

	for i := 0; i < pool; i++ {
		go downloadUrls(work, shutdown, &wg, options.DestinationFolder)
	}

	for _, url := range urls {
		work <- url
	}

	close(shutdown)

	wg.Wait()
}

func readLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

func downloadUrls(urls chan string, shutdown chan bool, wg *sync.WaitGroup, downloadFolder string) {
	for {
		select {
		case url := <-urls:
			downloadUrl(url, downloadFolder)
		case <-shutdown:
			wg.Done()
			return
		}
	}
}

func downloadUrl(url, downloadFolder string) {
	fmt.Println("Downloading: ", url)

	response, responseError := http.Get(url)

	if responseError != nil {
		fmt.Println("Error during downling", url, responseError.Error())
	}

	fmt.Println("finished")

	defer response.Body.Close()

	elements := []string{downloadFolder, buildFileName(url)}
	fullFilePath := strings.Join(elements, "/")
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

	return segments[len(segments)-1]
}
