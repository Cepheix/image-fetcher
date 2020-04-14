package main

import (
	"fmt"

	"github.com/jessevdk/go-flags"
)

type Options struct {
	SourceFile        string `short:"s" long:"source-file" description:"Source file containing all image urls"`
	DestinationFolder string `short:"d" long:"destination-folder" description:"Destination folder where to put all images"`
}

func main() {
	var options Options
	_, err := flags.Parse(&options)

	if err != nil {
		panic(err)
	}

	fmt.Println(options)

}
