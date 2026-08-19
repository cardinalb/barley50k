package submissionfile

import (
	"bufio"
	"log"
	"os"
	"regexp"
	"strings"
)

// Submission : submission struct
type Submission struct {
	Batch              string
	UniqueID           string
	Well               string
	Plate              string
	JHISeedstoreID     string
	ExternalID         string
	DataOwner          string
	DNAQuanitity       string
	BudgetCode         string
	ProjectName        string
	SubmittedBy        string
	DataSource         string
	GTDNAExtractionRef string
}

// LoadSubmissionFile : loads the project file into memory
func LoadSubmissionFile(filename string) map[string]Submission {
	file, err := os.Open(filename)

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)

	var split []string
	data := make(map[string]Submission)
	seen := false

	for scanner.Scan() {

		line := strings.TrimSuffix(scanner.Text(), "\n")

		if !seen {
			match, _ := regexp.MatchString(`Batch`, line)
			if match {
				seen = true
			}
		} else {

			split = strings.Split(line, "\t")

			var s Submission
			s.Batch = split[0]
			s.UniqueID = split[1]
			s.Well = split[2]
			s.Plate = split[3]
			s.JHISeedstoreID = split[4]
			s.ExternalID = split[5]
			s.DataOwner = split[6]
			s.DNAQuanitity = split[7]
			s.BudgetCode = split[8]
			s.ProjectName = split[9]
			s.SubmittedBy = split[10]
			s.DataSource = split[11]
			s.GTDNAExtractionRef = split[12]

			data[split[1]] = s

		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return data
}
