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
)

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Println("Usage: haversine-processor [coordinates.json]")
		fmt.Println("       haversine-processor [coordinates.json] [answers.f64]")
		os.Exit(1)
	}

	var err error
	var jsonAverage float64

	jsonAverage, err = jsonComputation(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Printf("Haversine average: %f\n", jsonAverage)

	if len(os.Args) == 3 {
		var binaryAverage float64

		binaryAverage, err = binaryComputation(os.Args[2])
		fmt.Printf("Reference average: %f\n", binaryAverage)
		fmt.Printf("Difference       : %f\n", jsonAverage-binaryAverage)
	}
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

	var pairs haversine.Pairs
	json.Unmarshall(rawBytes, &pairs)

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
