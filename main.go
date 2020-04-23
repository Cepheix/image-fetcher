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
	"time"

	"github.com/gosuri/uiprogress"
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

	count := len(urls)
	uiprogress.Start()
	bar := uiprogress.AddBar(count).AppendCompleted().PrependElapsed()
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return fmt.Sprintf("Image (%d/%d)", b.Current(), count)
	})

	work := make(chan string)
	shutdown := make(chan bool)
	pool := runtime.NumCPU()
	var wg sync.WaitGroup
	wg.Add(pool)

	for i := 0; i < pool; i++ {
		go downloadUrls(work, shutdown, &wg, bar, options.DestinationFolder)
	}

	for _, url := range urls {
		work <- url
	}

	close(shutdown)

	time.Sleep(time.Second) // wait for a second for all the go routines to finish
	wg.Wait()
	uiprogress.Stop()
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

func downloadUrls(urls chan string, shutdown chan bool, wg *sync.WaitGroup, progressBar *uiprogress.Bar, downloadFolder string) {
	for {
		select {
		case url := <-urls:
			downloadUrl(url, downloadFolder)
			progressBar.Incr()
		case <-shutdown:
			wg.Done()
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

	return segments[len(segments)-1]
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
