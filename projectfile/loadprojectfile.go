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

package projectfile

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
)

// LoadProjectFile : loads the project file into memory
func LoadProjectFile(filename string) (map[string]string, int) {
	file, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var split []string
	data := make(map[string]string)
	seen := false

	for scanner.Scan() {

		line := strings.TrimSuffix(scanner.Text(), "\n")

		if !seen {
			match, _ := regexp.MatchString(`Sample_ID`, line)
			if match {
				seen = true
			}
		} else {
			split = strings.Split(line, ",")
			sampleID := split[0]
			joinedID := (split[6] + "_" + split[7])
			data[joinedID] = sampleID
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	mapLength := len(data)

	return data, mapLength
}
