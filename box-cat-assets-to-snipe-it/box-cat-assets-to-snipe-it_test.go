package main

import (
	"encoding/csv"
	"io"
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

func TestWriteSnipeITCSV_RoundTrip(t *testing.T) {
	recs := []SnipeITRecord{
		{
			ItemName:  `Cable, "special" (2 off) [KEEP]`,
			Category:  "Connector",
			ModelName: "Generic-Model",
			AssetTag:  "QQ01-20260101000000-00000000",
			Location:  "Garage",
			BoxName:   "QQ01",
		},
		{
			ItemName:  "DB25M ↔ DB25M",
			Category:  "",
			ModelName: "Generic-Model",
			AssetTag:  "PSU Box (ex Wine Box)-20260101000000-00000001",
			BoxName:   "PSU Box (ex Wine Box)",
		},
	}

	path := filepath.Join(t.TempDir(), "out.csv")
	WriteSnipeITCSV(path, recs)

	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	rows, err := csv.NewReader(f).ReadAll() // also fails if column counts are inconsistent
	if err != nil {
		t.Fatalf("output is not valid CSV: %v", err)
	}

	if len(rows) != 3 {
		t.Fatalf("got %d rows, want 3 (header + 2)", len(rows))
	}

	wantHeader := []string{"Full Name", "Email", "Username", "item Name", "Category", "Model name",
		"Manufacturer", "Model Number", "Serial number", "Asset Tag", "Location", "Notes",
		"Purchase Date", "Purchase Cost", "Company", "Status", "Warranty", "Supplier", "BoxName"}
	if !reflect.DeepEqual(rows[0], wantHeader) {
		t.Errorf("header mismatch\n got: %v\nwant: %v", rows[0], wantHeader)
	}

	// Columns by position: 3=item Name, 4=Category, 9=Asset Tag, 10=Location, 18=BoxName
	for i, rec := range recs {
		row := rows[i+1]
		checks := []struct {
			col  int
			want string
		}{
			{3, rec.ItemName}, {4, rec.Category}, {9, rec.AssetTag},
			{10, rec.Location}, {18, rec.BoxName},
		}
		for _, c := range checks {
			if row[c.col] != c.want {
				t.Errorf("row %d col %d (%s) = %q, want %q", i+1, c.col, wantHeader[c.col], row[c.col], c.want)
			}
		}
	}
}

func writeTempCSV(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// captureStdout runs f and returns whatever it printed to stdout.
func captureStdout(t *testing.T, f func()) string {
	t.Helper()
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	f()
	w.Close()
	os.Stdout = old
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

const catHeader = "Box,Fullness,Sealed,Location,Category,Contents\n"

func TestProcessBoxCatContents_Warnings(t *testing.T) {
	cases := []struct {
		name, row, wantMsg string // wantMsg == "" means no output expected
	}{
		{"good verification", "Verification V1,V1,V1,V1,V1,V1", ""},
		{"bad verification", "Verification V1,V1,V1,X,V1,V1", "Badly formatted verification line"},
		{"empty box with data", "QQ01,Empty,,,,Stuff", "Empty box with data"},
		{"destroyed with data", "QQ01,Destroyed,Yes,,,", "Destroyed box with data"},
		{"no contents", "QQ01,Full,Yes,Garage,Misc,", "Unhandled no data stat"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := writeTempCSV(t, catHeader+c.row+"\n")
			out := captureStdout(t, func() { processBoxCatContents(path) })
			if c.wantMsg == "" && out != "" {
				t.Errorf("unexpected output: %q", out)
			}
			if c.wantMsg != "" && !strings.Contains(out, c.wantMsg) {
				t.Errorf("output %q does not contain %q", out, c.wantMsg)
			}
		})
	}
}

func TestProcessBoxCatContents_DroppedStates(t *testing.T) {
	cases := []struct{ name, row string }{
		{"empty lower", "QQ01,empty,,,,"},
		{"empty title", "QQ01,Empty,,,,"},
		{"empty upper", "QQ01,EMPTY,,,,"},
		{"destroyed padded", "QQ01, Destroyed ,,,,"},
		{"unassigned", "QQ01,unassigned,,,,"},
		{"not printed", "QQ01,Not Printed,,,,"},
		{"printed-unused", "QQ01,printed-unused,,,,"},
		{"empty but with contents", "QQ01,Empty,,Garage,Misc,Stuff"},
		{"verification upper", "VERIFICATION V2,V2,V2,V2,V2,V2"},
		{"verification lower", "verification v3,V3,V3,V3,V3,V3"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path := writeTempCSV(t, catHeader+c.row+"\n")
			if got := processBoxCatContents(path); len(got) != 0 {
				t.Errorf("row %q should be dropped, got %+v", c.row, got)
			}
		})
	}
}
func TestProcessBoxCatContents_NoHeader(t *testing.T) {
	path := writeTempCSV(t, "some,text,,,,\nQQ01,Full,Yes,Garage,Misc,Thing\n")
	if got := processBoxCatContents(path); len(got) != 0 {
		t.Errorf("expected nothing without a Box/Fullness header, got %+v", got)
	}
}
