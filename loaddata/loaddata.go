/*
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package loaddata

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path"
	"regexp"
	"strings"
)

// LoadData :
func LoadData(filename string) {
	file, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var split []string // this will hold the split lines data which can then be put into a map

	seen := false

	data := make(map[string]map[string]string)

	lines := make(map[string]int)
	markers := make(map[string]int)

	for scanner.Scan() {

		line := strings.TrimSuffix(scanner.Text(), "\n")

		if !seen {
			match, _ := regexp.MatchString(`SNP Name`, line)
			if match {
				seen = true
			}
		} else {
			split = strings.Split(line, "\t")
			snp := split[0]
			sample := split[1]
			value := split[2]

			lines[sample]++
			markers[snp]++

			mm, ok := data[snp]
			if !ok {
				mm = make(map[string]string)
				data[snp] = mm
			}
			mm[sample] = value

		}

	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	printDataTransposed(data, lines, markers, filename)
}

func printDataTransposed(data map[string]map[string]string, lines map[string]int, markers map[string]int, filename string) {
	// ok now we can look at printing this bad boy out
	// we need to know the exact order of the data that is here so thats markers
	// the lines doesnt matter so much so lets just ignore that for the time being :-)

	var markerOrder []string
	for key := range markers {
		markerOrder = append(markerOrder, key)
	}

	var lineOrder []string
	for key := range lines {
		lineOrder = append(lineOrder, key)
	}

	m := fmt.Sprintf("Number of markers in file (%s)  : %d", filename, len(markerOrder))
	l := fmt.Sprintf("Number of lines in the file (%s) : %d", filename, len(lineOrder))

	fmt.Println(m)
	fmt.Println(l)
	// now we need to iterate over the details and write to a file
	// Get the filename and path
	dir, file := path.Split(filename)

	newFilename := dir + "OUT_" + file

	outFile, err := os.Create(newFilename)
	if err != nil {
		fmt.Println(err)
		return
	}

	//print the markers header out
	outFile.WriteString("Marker/Line")

	for i := range lineOrder {
		outFile.WriteString(",")
		outFile.WriteString(lineOrder[i])
	}
	outFile.WriteString("\n")

	for k := range markerOrder {
		//fmt.Println(k)
		outFile.WriteString(markerOrder[k])

		for m := range lineOrder {

			outFile.WriteString(",")
			outFile.WriteString(data[markerOrder[k]][lineOrder[m]])
		}
		outFile.WriteString("\n")
	}

}

func printData(data map[string]map[string]string, lines map[string]int, markers map[string]int, filename string) {
	// ok now we can look at printing this bad boy out
	// we need to know the exact order of the data that is here so thats markers
	// the lines doesnt matter so much so lets just ignore that for the time being :-)

	var markerOrder []string
	for key := range markers {
		markerOrder = append(markerOrder, key)
	}

	fmt.Println(len(markerOrder))

	// now we need to iterate over the details and write to a file

	outFile, err := os.Create("OUT_transposed_" + filename)
	if err != nil {
		fmt.Println(err)
		return
	}

	//print the markers header out
	outFile.WriteString("/Line/Marker")

	for i := range markerOrder {
		outFile.WriteString(",")
		outFile.WriteString(markerOrder[i])
	}
	outFile.WriteString("\n")

	for k := range lines {
		outFile.WriteString(k)

		for m := range markerOrder {

			outFile.WriteString(",")
			outFile.WriteString(data[markerOrder[m]][k])
		}
		outFile.WriteString("\n")
	}

}
