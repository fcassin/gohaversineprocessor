package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"os"

	haversine "github.com/fcassin/gohaversine/haversine"
	json "github.com/fcassin/gojson/json"
	timer "github.com/fcassin/gotimer/timer"
)

var startTimer, startupTimer, readTimer, parseTimer, sumTimer,
	miscOutputTimer, cpuFrequency, binaryTimer int64

func main() {
	startTimer = timer.ReadCPUTimer()
	cpuFrequency = timer.GetCPUTimerFreq(100)

	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Usage: haversine-processor [coordinates.json]")
		fmt.Println("       haversine-processor [coordinates.json] [answers.f64]")
		os.Exit(1)
	}

	startupTimer = timer.ReadCPUTimer()

	var err error
	var jsonAverage float64

	jsonAverage, err = jsonComputation(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	output(jsonAverage)
}

func jsonComputation(name string) (average float64, err error) {
	var file *os.File
	var rawBytes []byte

	file, err = os.Open(name)
	if err != nil {
		return
	}
	defer file.Close()

	rawBytes, err = ioutil.ReadAll(file)
	if err != nil {
		return
	}

	readTimer = timer.ReadCPUTimer()

	var pairs haversine.Pairs
	json.Unmarshall(rawBytes, &pairs)

	parseTimer = timer.ReadCPUTimer()

	var sum float64

	for _, pair := range pairs.Pairs {
		distance := haversine.ReferenceHaversine(
			pair.X0,
			pair.X1,
			pair.Y0,
			pair.Y1,
			6372.8)
		sum = sum + distance
	}

	average = sum / float64(len(pairs.Pairs))

	sumTimer = timer.ReadCPUTimer()

	return
}

func binaryComputation(name string) (average float64, err error) {
	var file *os.File
	file, err = os.Open(name)
	if err != nil {
		return
	}
	defer file.Close()

	var sum, count float64
	var read int = 8

	for read != 0 {
		var buffer []byte = make([]byte, 8)
		read, err = file.Read(buffer)

		if err != nil {
			if err == io.EOF {
				err = nil
				continue
			} else {
				fmt.Println(err)
				os.Exit(1)
			}
		}

		sum = sum + math.Float64frombits(binary.LittleEndian.Uint64(buffer))
		count = count + 1
	}

	average = sum / count

	return
}

func output(jsonAverage float64) {
	var err error

	fmt.Printf("Haversine average: %f\n", jsonAverage)

	miscOutputTimer = timer.ReadCPUTimer()

	if len(os.Args) == 3 {
		var binaryAverage float64

		binaryAverage, err = binaryComputation(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Reference average: %f\n", binaryAverage)
		fmt.Printf("Difference       : %f\n", jsonAverage-binaryAverage)

		binaryTimer = timer.ReadCPUTimer()
	}

	outputTimers()
}

func outputTimers() {
	var totalRuntime, startupRuntime, readRuntime, parseRuntime, sumRuntime,
		miscOutputRuntime, binaryRuntime, cpuFrequencyMs int64
	totalRuntime = binaryTimer - startTimer
	startupRuntime = startupTimer - startTimer
	readRuntime = readTimer - startupTimer
	parseRuntime = parseTimer - readTimer
	sumRuntime = sumTimer - parseTimer
	miscOutputRuntime = miscOutputTimer - sumTimer
	binaryRuntime = binaryTimer - miscOutputTimer
	cpuFrequencyMs = cpuFrequency / 1000

	var totalDuration, startupDuration, readDuration, parseDuration, sumDuration,
		miscOutputDuration, binaryDuration float64

	totalDuration = float64(totalRuntime) / float64(cpuFrequencyMs)
	startupDuration = float64(startupRuntime) / float64(cpuFrequencyMs)
	readDuration = float64(readRuntime) / float64(cpuFrequencyMs)
	parseDuration = float64(parseRuntime) / float64(cpuFrequencyMs)
	sumDuration = float64(sumRuntime) / float64(cpuFrequencyMs)
	miscOutputDuration = float64(miscOutputRuntime) / float64(cpuFrequencyMs)
	binaryDuration = float64(binaryRuntime) / float64(cpuFrequencyMs)

	fmt.Println()
	fmt.Printf("Total time: %10.3fms (CPU freq %d)\n", totalDuration, cpuFrequency)
	fmt.Printf("     Startup: %10.3fms (%5.2f%%)\n", startupDuration, 100*startupDuration/totalDuration)
	fmt.Printf("        Read: %10.3fms (%5.2f%%)\n", readDuration, 100*readDuration/totalDuration)
	fmt.Printf("       Parse: %10.3fms (%5.2f%%)\n", parseDuration, 100*parseDuration/totalDuration)
	fmt.Printf("         Sum: %10.3fms (%5.2f%%)\n", sumDuration, 100*sumDuration/totalDuration)
	fmt.Printf("      Binary: %10.3fms (%5.2f%%)\n", binaryDuration, 100*binaryDuration/totalDuration)
	fmt.Printf("  MiscOutput: %10.3fms (%5.2f%%)\n", miscOutputDuration, 100*miscOutputDuration/totalDuration)

}
