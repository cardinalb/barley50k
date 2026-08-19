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

package setup

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	tempdirectory = "temp"
)

// Setup : created the temporary directory and files
func Setup() {
	fmt.Println(filepath.Dir(os.Args[0]))
	createTempDirectory()
	time.Sleep(2 * time.Second)
	//deleteTempDirectory()
}

func createTempDirectory() {
	_, err := os.Stat(tempdirectory)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(tempdirectory, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}
}

func deleteTempDirectory() {
	err := os.Remove(tempdirectory)
	if err != nil {
		log.Fatal(err)
	}
}
