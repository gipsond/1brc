package weather

import (
	"bufio"
	"container/heap"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const MAX_STATIONS int = 10_000

type temp = float64

type result struct {
    min temp
    avg temp
    max temp
}

type measurement struct {
    station string
    t temp
}

type station struct {
    name string
    id int
    heapI int
}

type StationSlice []*station

func (sh StationSlice) Len() int { return len(sh) }

func (sh StationSlice) Less(i, j int) bool {
	return sh[i].name < sh[j].name
}

func (sh StationSlice) Swap(i, j int) {
	sh[i], sh[j] = sh[j], sh[i]
	sh[i].heapI = i
	sh[j].heapI = j
}

func (sh *StationSlice) Push(x any) {
	n := len(*sh)
	item := x.(*station)
	item.heapI = n
	*sh = append(*sh, item)
}

func (sh *StationSlice) Pop() any {
	old := *sh
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.heapI = -1
	*sh = old[0 : n-1]
	return item
}

func read(reader io.Reader, lines chan<- string) {
    logger := log.New(os.Stderr, "read:", log.LstdFlags)
    scanner := bufio.NewScanner(reader)
    i := 0
    for scanner.Scan() {
        lines <- scanner.Text()
        i++
        if i % 1_000_000 == 0 {
            logger.Printf("%d lines\n", i)
        }
    }
    logger.Println("done")
    close(lines)
    if err := scanner.Err(); err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        os.Exit(1)
    }
}

func dispatch(lines <-chan string,
              tempouts [MAX_STATIONS]chan<- temp,
              orderout chan<- StationSlice) {
    logger := log.New(os.Stderr, "dispatch:", log.LstdFlags)
    logger.Println("starting")

    maxId := 0
    idByStation := make(map[string]int)
    stations := make(StationSlice, 0)

    i := 0

    for line := range lines {
        measurement := parseMeasurement(line)
        id, exists := idByStation[measurement.station]
        if !exists {
            id = maxId
            idByStation[measurement.station] = maxId
            maxId++
            s := new(station)
            s.name = measurement.station
            s.id = id
            s.heapI = -1
            logger.Printf("new station: %v\n", s)
            heap.Push(&stations, s)
        }
        tempouts[id] <- measurement.t
        i++
        if i % 1_000_000 == 0 {
            logger.Printf("%d lines\n", i)
        }
    }
    orderout <- stations
    for _, tempout := range tempouts {
        close(tempout)
    }
    close(orderout)
}

func parseMeasurement(line string) measurement {
    station, ts, _ := strings.Cut(line, ";")
    t, err := strconv.ParseFloat(ts, 64);
    if err != nil {
        fmt.Fprintln(os.Stderr, "error:", err)
        os.Exit(1)
    }
    return measurement{
        station: station,
        t: t,
    }
}

const MIN_TEMP temp = -99.9
const MAX_TEMP temp = 99.9

func collect(temps <-chan temp, out chan<- result) {
    logger := log.New(os.Stderr, "collect:", log.LstdFlags)
    logger.Println("starting")

    min := MAX_TEMP + 1
    avg := 0.0
    max := MIN_TEMP - 1
    n := 0

    for temp := range temps {
        if temp < min {
            min = temp
        }
        if temp > max {
            max = temp
        }
        avg *= float64(n)
        n += 1
        avg = (avg + temp) / float64(n)
    }

    out <- result{
        min: min,
        avg: avg,
        max: max,
    }
    close(out)
}

const LINE_BUF_N int = 100_000_000
const TEMP_BUF_N int = 100_000

func Summarize(input io.Reader) {
    lines := make(chan string, LINE_BUF_N)
    orderFuture := make(chan StationSlice)
    var results [MAX_STATIONS]chan result
    var tempChans [MAX_STATIONS]chan<- temp
    for id := range results {
        results[id] = make(chan result)
        tempChan := make(chan temp, TEMP_BUF_N)
        tempChans[id] = tempChan
        go collect(tempChan, results[id])
    }
    go dispatch(lines, tempChans, orderFuture)
    go read(input, lines)

    order := <-orderFuture
    n := order.Len();
    for i := 0; i < n; i++ {
        s := heap.Pop(&order).(*station)
        result := <-results[s.id]
        fmt.Printf("%s;%.1f;%.1f;%.1f\n", s.name, result.min, result.avg, result.max)
    }
}

