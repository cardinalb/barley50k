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
	"time"

	"github.com/gookit/color"
)

// https://en.wikipedia.org/wiki/Category:Scottish_legendary_creatures

// Welcome : print out the welcome text
func Welcome() {

	var welcomestring string = ("Barley 50K Processing (Am Fear Liath Mòr) v0.05.01.24")

	fmt.Println("\033[H\033[2J")
	color.Info.Print(welcomestring)
	fmt.Println()
	fmt.Println()
	color.Blue.Print("██████   █████  ██████  ██      ███████ ██    ██     ███████  ██████  ██   ██\n")
	color.Blue.Print("██   ██ ██   ██ ██   ██ ██      ██       ██  ██      ██      ██  ████ ██  ██\n")
	color.Blue.Print("██████  ███████ ██████  ██      █████     ████       ███████ ██ ██ ██ █████\n")
	color.Blue.Print("██   ██ ██   ██ ██   ██ ██      ██         ██             ██ ████  ██ ██  ██\n")
	color.Blue.Print("██████  ██   ██ ██   ██ ███████ ███████    ██        ███████  ██████  ██   ██\n")
	fmt.Println()
	color.Info.Print("Developed by Paul Shaw (ICS), Jim McNicol (BioSS) and Malcolm Macaulay (CMS)")
	fmt.Println()
	color.Info.Print("©2026 International Barley Hub / James Hutton Institute, Invergowrie, Scotland, DD2 5DA")
	fmt.Println()
	fmt.Println()
}

// StartFileLoadImport :
func StartFileLoadImport(stage string, file string) {
	color.Blue.Print(stage + " Dealing with ")
	color.Yellow.Print(file)
	fmt.Println()
}

// GoRoutineInfo :
func GoRoutineInfo() {
	fmt.Println()
	color.BgRed.Println("Loading files in parallel...")
	color.Red.Println("This takes about a minute to run on a reasonably fast machine.")
	fmt.Println()
}

// R :
func R() {
	fmt.Println()
	color.BgRed.Println("Getting R up and running...")
	color.Red.Println("This may take up to 10 minutes depending on the size of input files.")
}

func DBConnect() {
	fmt.Println()
	color.BgRed.Println("Connecting to DB...")
	color.Red.Println("This bit is super quick!")
}

// Debug :
func Debug(text time.Time) {
	color.Debug.Println(text)
}
