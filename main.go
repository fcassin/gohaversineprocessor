package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"

	haversine "github.com/fcassin/gohaversine/haversine"
	json "github.com/fcassin/gojson/json"
	timer "github.com/fcassin/gotimer/timer"
)

func main() {
	timer.Start("startup")

	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Usage: haversine-processor [coordinates.json]")
		fmt.Println("       haversine-processor [coordinates.json] [answers.f64]")
		os.Exit(1)
	}

	timer.Stop("startup")

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

	timer.Start("read")

	file, err = os.Open(name)
	if err != nil {
		return
	}
	defer file.Close()

	rawBytes, err = io.ReadAll(file)
	if err != nil {
		return
	}

	timer.Stop("read")
	timer.Start("parse")

	var pairs haversine.Pairs
	json.Unmarshall(rawBytes, &pairs)

	timer.Stop("parse")
	timer.Start("sum")

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

	timer.Stop("sum")

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

	if len(os.Args) == 3 {
		timer.Start("binary")
		var binaryAverage float64

		binaryAverage, err = binaryComputation(os.Args[2])
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Reference average: %f\n", binaryAverage)
		fmt.Printf("Difference       : %f\n", jsonAverage-binaryAverage)

		timer.Stop("binary")
	}

	outputTimers()
}

func outputTimers() {
	timer.Output()
}
