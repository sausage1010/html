// jsonInMem
package main

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"os"
	"log"
	"errors"
)

type School struct {
	SchoolID		string
	Name_full	string
	City			string
	State		string
	Country		string
}

func (s *School) Load(row []string) error {
	if len(row) != 5 {
		err := errors.New("Invalid row - needs 5 fields")
		return err
	}
	s.SchoolID = row[0]
	s.Name_full = row[1]
	s.City = row[2]
	s.State = row[3]
	s.Country = row[4]
	
	return nil
}

func main() {
	fin, fierr := os.Open("../json/lahman-csv_2015-01-24/Schools.csv")
	if fierr != nil {
		log.Fatal(fierr)
	}
	defer fin.Close()
	
	fout, foerr := os.Create("Schools.json")
	if foerr != nil {
		log.Fatal(foerr)
	}
	defer fout.Close()

	csvReader := csv.NewReader(fin)
	
	row, rerr := csvReader.Read() // remove headers
	if rerr != nil {
		log.Fatal(rerr)
	}
	
	var schools []School
	
	for {
		row, rerr = csvReader.Read()
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			log.Fatal(rerr)
		}
		
		var school School
		lerr := school.Load(row)
		if lerr != nil {
			log.Fatal(lerr)
		}
		schools = append(schools, school)
		
	}
	
	jBytes, encErr := json.MarshalIndent(schools, "", "  ")
	if encErr != nil {
		log.Fatal(encErr)
	}
	
	if _, werr := fout.Write(jBytes); werr != nil {
		log.Fatal(werr)
	}
	
}
