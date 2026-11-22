package main

import (
	"os"
	"path/filepath"
	"reflect"
        "regexp"
        "strings"
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

func TestBuildSnipeITContents_FieldsAndTags(t *testing.T) {
	in := []BoxCatRecord{
		{BoxName: "QQ01", Fullness: "Full", Sealed: "Yes", Location: "Garage", Category: "Computer", Contents: "Lynx"},
		{BoxName: "PSU Box (ex Wine Box)", Fullness: "Half", Sealed: "No", Location: "Shed", Category: "", Contents: "AC Adapter"},
	}

	got := BuildSnipeITContents(in)

	if len(got) != len(in) {
		t.Fatalf("got %d records, want %d", len(got), len(in))
	}

	// <BoxName>-<14-digit timestamp>-<8-digit index>
	tagRE := regexp.MustCompile(`^(.+)-(\d{14})-(\d{8})$`)

	var stamps []string
	for i, rec := range got {
		// Mapped fields
		if rec.ItemName != in[i].Contents {
			t.Errorf("record %d: ItemName = %q, want %q", i, rec.ItemName, in[i].Contents)
		}
		if rec.Category != in[i].Category {
			t.Errorf("record %d: Category = %q, want %q", i, rec.Category, in[i].Category)
		}
		if rec.Location != in[i].Location {
			t.Errorf("record %d: Location = %q, want %q", i, rec.Location, in[i].Location)
		}
		if rec.BoxName != in[i].BoxName {
			t.Errorf("record %d: BoxName = %q, want %q", i, rec.BoxName, in[i].BoxName)
		}
		if rec.ModelName != "Generic-Model" {
			t.Errorf("record %d: ModelName = %q, want %q", i, rec.ModelName, "Generic-Model")
		}

		// Fields that must stay empty
		for name, val := range map[string]string{
			"FullName": rec.FullName, "Email": rec.Email, "Username": rec.Username,
			"Manufacturer": rec.Manufacturer, "ModelNumber": rec.ModelNumber,
			"SerialNumber": rec.SerialNumber, "Notes": rec.Notes,
			"PurchaseDate": rec.PurchaseDate, "PurchaseCost": rec.PurchaseCost,
			"Company": rec.Company, "Status": rec.Status,
			"Warranty": rec.Warranty, "Supplier": rec.Supplier,
		} {
			if val != "" {
				t.Errorf("record %d: %s = %q, want empty", i, name, val)
			}
		}

		// Tag format
		m := tagRE.FindStringSubmatch(rec.AssetTag)
		if m == nil {
			t.Errorf("record %d: AssetTag %q does not match <box>-<14 digits>-<8 digits>", i, rec.AssetTag)
			continue
		}
		if m[1] != in[i].BoxName {
			t.Errorf("record %d: tag box prefix = %q, want %q", i, m[1], in[i].BoxName)
		}
		wantIndex := strings.Repeat("0", 7) + string(rune('0'+i))
		if m[3] != wantIndex {
			t.Errorf("record %d: tag index = %q, want %q", i, m[3], wantIndex)
		}
		stamps = append(stamps, m[2])
	}

	// All tags from one run share the same timestamp
	for _, s := range stamps {
		if s != stamps[0] {
			t.Errorf("timestamps differ within one run: %v", stamps)
			break
		}
	}
}

func TestBuildSnipeITContents_Empty(t *testing.T) {
	if got := BuildSnipeITContents(nil); len(got) != 0 {
		t.Errorf("got %d records for empty input, want 0", len(got))
	}
}
