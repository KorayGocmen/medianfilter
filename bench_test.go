package MedianFilter

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func Benchmark(b *testing.B) {

	testFolder := "test-frames/"
	outPath := testFolder + "out.png"

	var filePaths []string
	files, err := ioutil.ReadDir(testFolder)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		filePaths = append(filePaths, testFolder+f.Name())
	}

	if err := RemoveMovingObjs(filePaths, outPath); err != nil {
		fmt.Println(err)
	}

	os.Remove(outPath)
}
