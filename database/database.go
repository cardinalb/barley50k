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

package database

// Adds the information in the submission file

import (
	"database/sql"
	"fmt"

	"github.com/cardinalb/go50k/submissionfile"

	_ "github.com/lib/pq" //importing database drivers
)

// defined the constants to connect to postgres. The driver can be changed
// in the import section if another database is to be used.
const (
	host     = "localhost"
	port     = 5432
	user     = "barley"
	password = "barley"
	dbname   = "barley"
)

// These are just to keep track of the unique and duplicates
// there should only really be an issue with the duplicates if
// there is a problem inserting data and you try again so will
// need to look at a way that we can cache the entries and rollback
// if a problem crops up on insert so we remove all the batch.
var duplicatecounter = 0
var uniquecounter = 0

// DBconnect : connect to the database which will let us store the information from the
// submissions file. In this case we are using postgres for a change from mysql and with
// an eye to moving to postgres as we move forwards.
func DBconnect(data map[string]submissionfile.Submission, idmap map[string]string) {
	fmt.Println("Trying to connect to the database")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	err = db.Ping()
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully connected to the database!")

	// bit of a hack but we can reverse the idmap array now
	reversedIDMap := make(map[string]string)
	for a, b := range idmap {
		fmt.Println(a, " : ", b)
		reversedIDMap[b] = a
	}

	for key := range data {
		value := data[key]

		sentrixid := reversedIDMap[value.UniqueID]
		if sentrixid == "nil" {
			sentrixid = "undef"
		}

		addData(db, value, sentrixid)

	}
	fmt.Println("This batch has a total of ", uniquecounter, " unique values and ", duplicatecounter, " duplicates")

}

// This adds data from a map containing a reference to the Submission struct which holds every line
// in the submission file. Using a switch statement to work out if we have seen stuff before. Also
// running a select on ALL unique IDS just to make sure they definitely dont exist in the database
// already
func addData(db *sql.DB, values submissionfile.Submission, sentrixid string) {
	//start by checking to see if the entry exists
	checkSQLStatement := `SELECT uniqueid, jhiseedstoreid, externalid FROM submissions WHERE uniqueid=$1`
	row := db.QueryRow(checkSQLStatement, values.UniqueID)

	var uniqueid, jhiseedstoreid, externalid string

	switch err := row.Scan(&uniqueid, &jhiseedstoreid, &externalid); err {
	case sql.ErrNoRows:
		// this is what we want and means the ID is totally unique
		uniquecounter++
		sqlStatement := `
			INSERT INTO submissions (uniqueid, batch, well, plate, jhiseedstoreid, externalid, dataowner, dnaquantity, budgetcode,
				projectname, submittedby, datasource, gtdnaextractionref, sentrixid)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`

		_, err := db.Exec(sqlStatement, values.UniqueID, values.Batch, values.Well, values.Plate,
			values.JHISeedstoreID, values.ExternalID, values.DataOwner, values.DNAQuanitity,
			values.BudgetCode, values.ProjectName, values.SubmittedBy, values.DataSource, values.GTDNAExtractionRef, sentrixid)
		if err != nil {
			panic(err)
		}
	case nil:
		//fmt.Println(uniqueid, jhiseedstoreid, externalid)
		duplicatecounter++
	default:
		panic(err)
	}

}
