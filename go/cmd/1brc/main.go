package main

import (
	"flag"
	"log"
	"os"
	"runtime/pprof"

	"gipsond.github.io/1brc/internal/weather"
)

var profilePath = flag.String("profile", "", "write cpu profile to given file") 

func main() {
    flag.Parse()
    if *profilePath != "" {
        f, err := os.Create(*profilePath)
        if err != nil {
            log.Fatal(err)
        }
        pprof.StartCPUProfile(f)
        defer pprof.StopCPUProfile()
    }

    weather.Summarize(os.Stdin)
}

