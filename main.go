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

package main

import (
	"archive/zip"
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cardinalb/go50k/loaddata"
	"github.com/cardinalb/go50k/projectfile"
	"github.com/cardinalb/go50k/setup"
	"github.com/cardinalb/go50k/submissionfile"
)

var (
	projectFilePtr = flag.String("project", "", "Location of project file")
)

func main() {
	CallClear()

	timeStart := time.Now()

	flag.Parse()

	setup.Welcome()

	// Project.csv file input
	setup.StartFileLoadImport("(Stage 1/6)", *projectFilePtr+"/Project.csv")

	// mappings is a map that has the sentrix ID as a key and the JHI 50K ID as the value. This
	// is used to update the input files with an ID that we can actually work with.
	// idMapCount is the count of the number of entries in the Project.csv file.
	mappings, idMapCount := projectfile.LoadProjectFile(*projectFilePtr + "/Project.csv")

	fmt.Println("The number of entries in the Project.csv file is:", idMapCount)

	fmt.Println(mappings) // comment this out for the moment but include it back again under debug

	setup.StartFileLoadImport("(Stage 2/6)", *projectFilePtr+"/submission_template.txt")
	submissionData := submissionfile.LoadSubmissionFile(*projectFilePtr + "/submission_template.txt")

	// So we have submissionData which is an array (json) that has all the template data
	// this needs to be put into a database now.

	fmt.Println(submissionData)

	var wg sync.WaitGroup

	// these are the input files. We can change this later so they are passed in from the command line.
	rFile := *projectFilePtr + "/FinalReportR.txt"
	thetaFile := *projectFilePtr + "/FinalReportTheta.txt"

	var files = []string{rFile, thetaFile}

	trackstage := 4
	setup.GoRoutineInfo()
	for _, file := range files {
		stage := fmt.Sprintf("(Stage %d/%d)", trackstage, 6)
		setup.StartFileLoadImport(stage, file)
		trackstage++

		wg.Add(1)
		go func(f string) {
			defer wg.Done()
			loaddata.LoadData(f)
		}(file)
	}
	wg.Wait()

	// Now we need to run the R code on the data
	setup.R()

	//rcode.ReturnCode(*outputPtr)

	setup.StartFileLoadImport("(Stage 6/6)", "jim_modified.R")
	runRCode(mappings)

	timeEnd := time.Now()

	fmt.Println()
	setup.Debug(timeStart)
	setup.Debug(timeEnd)

}

func runRCode(mappings map[string]string) {

	fmt.Println("Running R code...")

	rcodelocation := "NEW_CLUSTER.R"
	rlocation := *projectFilePtr + "/OUT_FinalReportR.txt"
	tlocation := *projectFilePtr + "/OUT_FinalReportTheta.txt"

	fmt.Println(rcodelocation)
	fmt.Println(rlocation)
	fmt.Println(tlocation)

	err, _ := exec.Command("Rscript", rcodelocation, rlocation, tlocation, "70", ".").Output()
	// err, _ := exec.Command("Rscript", "jim_modified.R", "OUTFinalReportR.txt", "OUTFinalReportTheta.txt", "69", "./").Output()

	if err != nil {
		fmt.Println(err)
	}

	convertAB2ACGT()
	ProcessGenotypeData(mappings)
	projectName := filepath.Base(filepath.Clean(*projectFilePtr))
	if err := archiveOutputFiles(projectName); err != nil {
		log.Fatalf("failed to archive output files: %v", err)
	}

}

func archiveOutputFiles(projectName string) error {
	return archiveOutputFilesInDirectory(projectName, "sample_data")
}

func archiveOutputFilesInDirectory(projectName, outputDirectory string) error {
	archivePath := filepath.Join(outputDirectory, projectName+".zip")
	temporaryArchivePath := archivePath + ".tmp"
	filesToArchive := []string{
		"final_calls.txt",
		"OUT_FinalReportR.txt",
		"OUT_FinalReportTheta.txt",
		"parms.csv",
		"stats.csv",
		"STATS.txt",
		"ids.csv",
	}

	temporaryArchive, err := os.Create(temporaryArchivePath)
	if err != nil {
		return fmt.Errorf("create archive: %w", err)
	}

	archiveWriter := zip.NewWriter(temporaryArchive)
	for _, filename := range filesToArchive {
		filePath := filepath.Join(outputDirectory, filename)
		inputFile, err := os.Open(filePath)
		if err != nil {
			archiveWriter.Close()
			temporaryArchive.Close()
			os.Remove(temporaryArchivePath)
			return fmt.Errorf("open %s: %w", filePath, err)
		}

		archiveEntry, err := archiveWriter.Create(filename)
		if err == nil {
			_, err = io.Copy(archiveEntry, inputFile)
		}
		inputFile.Close()
		if err != nil {
			archiveWriter.Close()
			temporaryArchive.Close()
			os.Remove(temporaryArchivePath)
			return fmt.Errorf("add %s to archive: %w", filePath, err)
		}
	}

	if err := archiveWriter.Close(); err != nil {
		temporaryArchive.Close()
		os.Remove(temporaryArchivePath)
		return fmt.Errorf("finish archive: %w", err)
	}
	if err := temporaryArchive.Close(); err != nil {
		os.Remove(temporaryArchivePath)
		return fmt.Errorf("close archive: %w", err)
	}
	if err := os.Remove(archivePath); err != nil && !os.IsNotExist(err) {
		os.Remove(temporaryArchivePath)
		return fmt.Errorf("replace existing archive: %w", err)
	}
	if err := os.Rename(temporaryArchivePath, archivePath); err != nil {
		os.Remove(temporaryArchivePath)
		return fmt.Errorf("install archive: %w", err)
	}

	for _, filename := range filesToArchive {
		if err := os.Remove(filepath.Join(outputDirectory, filename)); err != nil {
			return fmt.Errorf("remove %s: %w", filename, err)
		}
	}

	return nil
}

var clear map[string]func() //create a map for storing clear funcs

func init() {
	clear = make(map[string]func()) //Initialize it
	clear["linux"] = func() {
		cmd := exec.Command("clear") //Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clear["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") //Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func CallClear() {
	value, ok := clear[runtime.GOOS] //runtime.GOOS -> linux, windows, darwin etc.
	if ok {                          //if we defined a clear func for that platform:
		value() //we execute it
	} else { //unsupported platform
		panic("Your platform is unsupported! I can't clear terminal screen :(")
	}
}

// To Run
//go run go50k.go -template ./sample_data/submission_template.txt -project ./sample_data/Project.csv --r ./sample_data/FinalReportR.txt --theta ./sample_data/FinalReportTheta.txt

func convertAB2ACGT() {
	// Define file paths [cite: 1]
	inputPath := "sample_data/ids.csv"
	outputPath := "sample_data/final_calls.txt"
	refPath := "calls.txt"

	// ---------------------------------------------------------
	// 1. Build the Reference Map
	// Equivalent to %calls_reference hash in Perl [cite: 1]
	// Structure: map[Marker]map[Allele]Call
	// ---------------------------------------------------------
	callsReference := make(map[string]map[string]string)

	refFile, err := os.Open(refPath)
	if err != nil {
		log.Fatalf("Error opening reference file: %v", err)
	}
	defer refFile.Close()

	scanner := bufio.NewScanner(refFile)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")

		// Ensure line has enough parts to avoid index out of range errors
		if len(parts) >= 6 {
			// Perl: ($type, $marker, $a_allele, $a_allele_call, $b_allele, $b_allele_call) [cite: 2]
			marker := parts[1]
			aAllele := parts[2]
			aCall := parts[3]
			bAllele := parts[4]
			bCall := parts[5]

			// Initialize inner map if it doesn't exist
			if _, exists := callsReference[marker]; !exists {
				callsReference[marker] = make(map[string]string)
			}

			// Map alleles to calls [cite: 3]
			callsReference[marker][aAllele] = aCall
			callsReference[marker][bAllele] = bCall
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// ---------------------------------------------------------
	// 2. Process Input and Write Output
	// ---------------------------------------------------------
	inFile, err := os.Open(inputPath) // [cite: 4]
	if err != nil {
		log.Fatalf("Problem here paul... %v\n", err) // keeping original error message style
	}
	defer inFile.Close()

	outFile, err := os.Create(outputPath) // [cite: 4]
	if err != nil {
		log.Fatal(err)
	}
	defer outFile.Close()

	writer := bufio.NewWriterSize(outFile, 256*1024)
	defer writer.Flush()

	scanner = bufio.NewScanner(inFile)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	counter := 0

	for scanner.Scan() { // [cite: 5]
		row := scanner.Text()

		if counter != 0 {
			// --- DATA ROWS LOGIC ---

			// Remove quotes globally
			row = strings.ReplaceAll(row, "\"", "")

			// Split by comma
			calls := strings.Split(row, ",")

			// Extract marker name (first element)
			if len(calls) > 0 {
				markerName := calls[0]

				// Remove the first element (marker) so we are left with just calls
				// Also remove the last element (pop)
				if len(calls) > 1 {
					calls = calls[1 : len(calls)-1]
				} else {
					calls = []string{}
				}

				markerCalls := callsReference[markerName]
				writer.WriteString(markerName)

				for _, call := range calls {
					// Split "XY" into "X" and "Y"
					// Assuming ASCII input for genetic data
					if len(call) != 2 {
						// Handle edge case where data might be missing or malformed
						writer.WriteString("\t--")
						continue
					}

					partA := string(call[0])
					partB := string(call[1])

					// Lookup Part A [cite: 7]
					if val, ok := markerCalls[partA]; ok {
						partA = val
					} else {
						partA = "-"
					}

					// Lookup Part B [cite: 8]
					if val, ok := markerCalls[partB]; ok {
						partB = val
					} else {
						partB = "-"
					}

					writer.WriteString("\t")
					writer.WriteString(partA)
					writer.WriteString(partB)
				}
				writer.WriteString("\n")
			}

		} else {
			// --- HEADER ROW LOGIC ---
			// Perl: $row =~ tr/,/\t/;
			row = strings.ReplaceAll(row, ",", "\t")

			markers := strings.Split(row, "\t")

			// pop(@markers);
			if len(markers) > 0 {
				markers = markers[:len(markers)-1]
			}
			if len(markers) > 0 && markers[0] == "Row.Names" {
				markers[0] = "Lines/Markers"
			}

			//writer.WriteString("Line/Marker")
			// The Perl script splits the first line by tab (after translate),
			// but markers[0] is usually empty or specific in these formats.
			// We join the whole slice as per the Perl logic: print OUTPUT join( "\t", @markers )
			if len(markers) > 0 {
				//writer.WriteString("\t")

				writer.WriteString(strings.Join(markers, "\t"))
			}
			writer.WriteString("\n")

			counter++
		}
	}

	// Flush buffer to ensure all data is written to file
	writer.Flush()

	fmt.Println("Conversion completed.")
}

func ProcessGenotypeData(mappings map[string]string) {

	inputPath := "sample_data/final_calls.txt"
	projectName := filepath.Base(filepath.Clean(*projectFilePtr))
	outputPath := filepath.Join("sample_data", projectName+"_FINAL.txt")
	statsPath := "sample_data/STATS.txt"
	nameMapPath := "20170301_44040_ALL_NAMES_MAP.txt"

	nameMapFile, err := os.Open(nameMapPath)
	if err != nil {
		log.Fatalf("failed to open name map %s: %v", nameMapPath, err)
	}
	defer nameMapFile.Close()

	nameMap := make(map[string]string)
	nameMapScanner := bufio.NewScanner(nameMapFile)
	nameMapScanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	for nameMapScanner.Scan() {
		mapFields := strings.Split(nameMapScanner.Text(), "\t")
		if len(mapFields) >= 2 {
			nameMap[mapFields[0]] = mapFields[1]
		}
	}
	if err := nameMapScanner.Err(); err != nil {
		log.Fatalf("failed to read name map %s: %v", nameMapPath, err)
	}

	// 1. Open Input File
	inFile, err := os.Open(inputPath)
	if err != nil {
		//return fmt.Errorf("failed to open input: %w", err)
	}
	defer inFile.Close()

	// 2. Create Output File (Replaced/Cleaned data)
	outFile, err := os.Create(outputPath)
	if err != nil {
		//return fmt.Errorf("failed to create output: %w", err)
	}
	defer outFile.Close()
	outWriter := bufio.NewWriterSize(outFile, 256*1024)
	defer outWriter.Flush()

	// 3. Initialize Stats Map
	alleles := make(map[string]int)

	scanner := bufio.NewScanner(inFile)
	scanner.Buffer(make([]byte, 64*1024), 16*1024*1024)
	isHeader := true

	// 4. Process Line by Line
	for scanner.Scan() {
		row := scanner.Text() // Removes newline (equivalent to chomp)

		// Handle Header (Row 0)
		if isHeader {
			headerFields := strings.Split(row, "\t")
			for index, field := range headerFields {
				if field == "Row.Names" {
					headerFields[index] = ""
					continue
				}
				if mappedValue, ok := mappings[field]; ok {
					headerFields[index] = mappedValue
				}
			}

			if _, err := outWriter.WriteString(strings.Join(headerFields, "\t") + "\n"); err != nil {
				//return err
			}
			isHeader = false
			continue
		}

		// Handle Data Rows
		fields := strings.Split(row, "\t")
		if len(fields) == 0 {
			continue
		}

		// Perl[cite: 5]: First column is marker, printed immediately
		marker := fields[0]
		if mappedMarker, ok := nameMap[marker]; ok {
			marker = mappedMarker
		}
		outWriter.WriteString(marker)

		// Process remaining columns (calls)
		calls := fields[1:]
		for _, c := range calls {

			// Normalization Logic [cite: 6-12]
			// Collapsed into a switch for efficiency
			switch c {
			case "", "N/A", "-F", "FF", "BB", "BG":
				c = "--"
			case " C C", "-C", "C-", "C", "CC":
				c = "C"
			case "-A", "A", "A-", "AA":
				c = "A"
			case "-T", "T", "T-", "TT":
				c = "T"
			case "-G", "G-", "G", "GG":
				c = "G"
			case "CA", "AC":
				c = "A/C"
			case "TA", "AT":
				c = "A/T"
			case "TC":
				c = "C/T"
			case "TG", "GT":
				c = "G/T"
			case "GC", "CG":
				c = "C/G"
			case "AG", "GA":
				c = "A/G"
			default:
				// If the call doesn't match any known pattern, leave it as is

			}

			// Collect stats
			alleles[c]++

			// Write call with leading tab
			outWriter.WriteString("\t")
			outWriter.WriteString(c)
		}
		outWriter.WriteString("\n")
	}

	if err := scanner.Err(); err != nil {
		//return fmt.Errorf("error reading input: %w", err)
	}

	// 5. Write Stats File [cite: 14]
	statsFileObj, err := os.Create(statsPath)
	if err != nil {
		//return fmt.Errorf("failed to create stats file: %w", err)
	}
	defer statsFileObj.Close()
	statsWriter := bufio.NewWriter(statsFileObj)
	defer statsWriter.Flush()

	// Sort keys to match Perl's `sort keys %alleles`
	sortedKeys := make([]string, 0, len(alleles))
	for k := range alleles {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, call := range sortedKeys {
		line := call + "\t" + strconv.Itoa(alleles[call]) + "\n"
		if _, err := statsWriter.WriteString(line); err != nil {
			//return err
		}
	}

	//return nil
}
