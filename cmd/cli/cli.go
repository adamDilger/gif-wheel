package main

import (
	"flag"
	"gif_wheel/wheel"
	"os"
	"strings"
)

var (
	items  = flag.String("i", "", "Comma separated list of items")
	output = flag.String("o", "", "Output file name")
	help   = flag.Bool("h", false, "Show help")
)

func main() {
	flag.Parse()

	if *help {
		flag.PrintDefaults()
		return
	}

	is := strings.Split(*items, ",")

	f, err := os.Create(*output)
	if err != nil {
		panic(err)
	}

	wheel := wheel.NewWheel(60, 600, 600, 250, is)

	err = wheel.BuildGif(f)
	if err != nil {
		panic(err)
	}
}
