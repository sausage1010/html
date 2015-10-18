// json
package main

import (
	"encoding/csv"
	"encoding/json"
//	"fmt"
	"io"
	"os"
//	"strings"
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
	fin, fierr := os.Open("lahman-csv_2015-01-24/Schools.csv")
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
	
	row, rerr := csvReader.Read()  // Get the header
	if rerr != nil {
		log.Fatal(rerr)
	}
	
	if _, foerr = fout.WriteString("["); foerr != nil {
		log.Fatal(foerr)
	}
	
	firstRecord := true
	
	for {
		row, rerr = csvReader.Read()
//		fmt.Println(row)
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			log.Fatal(rerr)
		}
		
		var school School
		school.Load(row)
		
		bytes, jerr := json.MarshalIndent(school, "", "  ")
		if jerr != nil {
			log.Fatal(jerr)
		}
		
//		fmt.Println(string(bytes))
		if !firstRecord {
			fout.WriteString(",\n")
		} else {
			firstRecord = false
		}
		
		fout.Write(bytes)
	}
	
	if _, foerr = fout.WriteString("]"); foerr != nil {
		log.Fatal(foerr)
	}

//	fmt.Print(records)
	
/*	var obj map[string]interface{}
	
	array := []float64{12.3, 16.2, 57.6, 1.999, 0.0, -86.3142}
	
	jsonThing, err := json.Marshal(array)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(jsonThing))
	
	err2 := json.Unmarshal([]byte(jsonData), &obj)
	if err2 != nil {
		panic(err2)
	}
	fmt.Println(obj)
*/
}
