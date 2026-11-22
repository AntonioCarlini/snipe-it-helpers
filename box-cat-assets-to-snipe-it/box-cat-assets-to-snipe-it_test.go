package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProcessBoxCatContents_FiltersRows(t *testing.T) {
	// Every row has 6 fields: encoding/csv rejects rows with differing field counts.
	input := `Some free-form text,,,,,
More description,,,,,
Box,Fullness,Sealed,Location,Category,Contents
QQ01,Full,Yes,Garage,Computer,Lynx
QQ02,Empty,,,,
QQ03,,,,,
Verification V1,V1,V1,V1,V1,V1
QQ04,Half,No,Shed,Misc,Cables
QQ05,Destroyed,,,,
QQ06,Full,Yes,Garage,Misc,
QQ07,Part,No,Loft,Toy,Beach toys
`

	path := filepath.Join(t.TempDir(), "catalogue.csv")
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}

	got := processBoxCatContents(path)

	want := []BoxCatRecord{
		{BoxName: "QQ01", Fullness: "Full", Sealed: "Yes", Location: "Garage", Category: "Computer", Contents: "Lynx"},
		{BoxName: "QQ04", Fullness: "Half", Sealed: "No", Location: "Shed", Category: "Misc", Contents: "Cables"},
		{BoxName: "QQ07", Fullness: "Part", Sealed: "No", Location: "Loft", Category: "Toy", Contents: "Beach toys"},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("unexpected records\n got: %+v\nwant: %+v", got, want)
	}
}
