package main

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
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

	downloadWork := DownloadWork{
		downloadFolder: options.DestinationFolder,
		work:           work,
		shutdown:       shutdown,
		wg:             &wg, progressBar: bar}

	for i := 0; i < pool; i++ {
		go downloadUrls(downloadWork)
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
