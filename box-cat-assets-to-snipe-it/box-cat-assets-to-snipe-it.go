package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"time"
)

// BoxCatRecord is a data structure that reflects the information in the existing catalogue
type BoxCatRecord struct {
	BoxName  string
	Fullness string
	Sealed   string
	Location string
	Category string
	Contents string
}

// SnipeITRecord is a data structure that holds the information that Snipe-IT needs
type SnipeITRecord struct {
	FullName     string
	Email        string
	Username     string
	ItemName     string
	Category     string
	ModelName    string
	Manufacturer string
	ModelNumber  string
	SerialNumber string
	AssetTag     string
	Location     string
	Notes        string
	PurchaseDate string
	PurchaseCost string
	Company      string
	Status       string
	Warranty     string
	Supplier     string
	BoxName      string
}

func main() {

	flag.Parse()

	inputs := flag.Args()
	if len(inputs) != 2 {
		log.Fatalf("Exactly 2 arguments required but %d supplied\n", len(inputs))
	}

	boxCatFilename := flag.Arg(0)
	outputFile := flag.Arg(1)

	boxCat := processBoxCatContents(boxCatFilename)

	snipeData := BuildSnipeITContents(boxCat)

	WriteSnipeITCSV(outputFile, snipeData)
}

func WriteSnipeITCSV(filename string, snipeData []SnipeITRecord) {
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	headers := []string{"Full Name", "Email", "Username", "item Name", "Category", "Model name", "Manufacturer", "Model Number",
		"Serial number", "Asset Tag", "Location", "Notes", "Purchase Date", "Purchase Cost", "Company", "Status",
		"Warranty", "Supplier",
		"BoxName"}

	writer.Write(headers)
	for _, data := range snipeData {
		row := []string{data.FullName,
			data.Email,
			data.Username,
			data.ItemName,
			data.Category,
			data.ModelName,
			data.Manufacturer,
			data.ModelNumber,
			data.SerialNumber,
			data.AssetTag,
			data.Location,
			data.Notes,
			data.PurchaseDate,
			data.PurchaseCost,
			data.Company,
			data.Status,
			data.Warranty,
			data.Supplier,
			data.BoxName,
		}
		writer.Write(row)
	}
}

func BuildSnipeITContents(boxCatrecords []BoxCatRecord) []SnipeITRecord {
	var snipeITrecords []SnipeITRecord
	for index, entry := range boxCatrecords {
		var data SnipeITRecord

		currentTime := time.Now()
		tag := fmt.Sprintf("%4d%02d%02d%02d%02d%02d-%08d", currentTime.Year(), currentTime.Month, currentTime.Day(), currentTime.Hour(), currentTime.Minute(), currentTime.Second(), index)

		data.FullName = ""
		data.Email = ""
		data.Username = ""
		data.ItemName = entry.Contents
		data.Category = entry.Category
		data.ModelName = "Generic-Model"
		data.Manufacturer = ""
		data.ModelNumber = ""
		data.SerialNumber = ""
		data.AssetTag = entry.BoxName + "-" + tag
		data.Location = entry.Location
		data.Notes = ""
		data.PurchaseDate = ""
		data.PurchaseCost = ""
		data.Company = ""
		data.Status = ""
		data.Warranty = ""
		data.Supplier = ""
		data.BoxName = entry.BoxName

		snipeITrecords = append(snipeITrecords, data)
	}

	return snipeITrecords
}

func processBoxCatContents(filename string) []BoxCatRecord {
	records := readCsvFile(filename)

	var boxcat []BoxCatRecord
	var skipHeaders bool = true
	for index, entry := range records {
		if skipHeaders {
			if (entry[0] == "Box") && (entry[1] == "Fullness") {
				skipHeaders = false
			}
			continue
		}

		// Skip an entry that is entirely blank apart from the box name
		if len(entry[1]) == 0 && len(entry[2]) == 0 && len(entry[3]) == 0 && len(entry[4]) == 0 && len(entry[5]) == 0 {
			continue
		}

		// Skip a verification entry, but complain if it is not properly formatted
		verif_prefix := "verification v"
		if strings.HasPrefix(strings.ToLower(entry[0]), verif_prefix) {
			prefix_len := len(verif_prefix)
			expected_data := "V" + entry[0][prefix_len:]
			if (entry[1] != expected_data) || (entry[2] != expected_data) || (entry[3] != expected_data) || (entry[4] != expected_data) || (entry[5] != expected_data) {
				fmt.Println("Badly formatted verification line:", entry)
			}
			continue
		}

		statesToDrop := []string{"empty", "destroyed", "unassigned", "not printed", "printed-unused"}
		fullness := strings.TrimSpace(strings.ToLower(entry[1]))
		if slices.Contains(statesToDrop, fullness) {
			switch fullness {
			case "empty":
				// An empty box can specify a location
				if len(entry[5]) > 0 {
					fmt.Println("Empty box with data at", index, entry)
				}
			case "destroyed":
				if (len(entry[2]) > 0) || (len(entry[3]) > 0) || (len(entry[4]) > 0) || (len(entry[5]) > 0) {
					fmt.Println("Destroyed box with data at", index, entry)
				}
			case "unassigned":
				if (len(entry[2]) > 0) || (len(entry[3]) > 0) || (len(entry[4]) > 0) || (len(entry[5]) > 0) {
					fmt.Println("Unassigned box with data at", index, entry)
				}
			case "not printed":
				if (len(entry[2]) > 0) || (len(entry[3]) > 0) || (len(entry[4]) > 0) || (len(entry[5]) > 0) {
					fmt.Println("Unprinted box label with data at", index, entry)
				}
			case "printed-unused":
				if (len(entry[2]) > 0) || (len(entry[3]) > 0) || (len(entry[4]) > 0) || (len(entry[5]) > 0) {
					fmt.Println("Unused box label with data at", index, entry)
				}
			default:
				fmt.Println("Unhandled fullness stat: at", index, entry, "[", fullness, "]")
			}
			continue
		} else if len(entry[5]) == 0 {
			fmt.Println("Unhandled no data stat: at", index, entry)
			continue
		}

		var data BoxCatRecord

		data.BoxName = entry[0]
		data.Fullness = entry[1]
		data.Sealed = entry[2]
		data.Location = entry[3]
		data.Category = entry[4]
		data.Contents = entry[5]

		boxcat = append(boxcat, data)
	}
	return boxcat
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()

	csvReader := csv.NewReader(f)
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}
