# snipe-it-helpers

Small helper programs for getting data from a catalogue spreadsheet into [Snipe-IT](https://snipeit.app/).

## box-cat-assets-to-snipe-it

Converts my "catalogue" spreadsheet (what is in which storage box, and where
the box lives) into a CSV that Snipe-IT's asset importer can read.

### Building and running

```
go build -o box-cat-assets-to-snipe-it ./box-cat-assets-to-snipe-it
./box-cat-assets-to-snipe-it <catalogue.csv> <snipe-it-output.csv>
```

or without building:

```
go run ./box-cat-assets-to-snipe-it <catalogue.csv> <snipe-it-output.csv>
```

Exactly two arguments are required: the input file and the output file.
The output file is overwritten if it already exists.

### Preparing the input

1. Export the catalogue spreadsheet as CSV.
   options (delimiter, encoding, quoting).
2. The sheet may start with free-form text. Everything is ignored until a row
   whose first two cells are `Box` and `Fullness`. Every row after that is treated
   as data.

Columns, in this order are processed (any further columns are ignored):

| # | Column   | Used?  | Notes |
|---|----------|--------|-------|
| 1 | Box      | Yes    | Box label, e.g. `QQ04`. Becomes the asset-tag prefix and the `BoxName` custom field |
| 2 | Fullness | Checks | How full the box is. Not output, but certain values change how the row is handled (see below) |
| 3 | Sealed   | No     | How the box is closed. Not output |
| 4 | Location | Yes    | Becomes Snipe-IT `Location` |
| 5 | Category | Yes    | Becomes Snipe-IT `Category`.  |
| 6 | Contents | Yes    | Free text describing the item. Becomes `item Name` |

### What gets dropped

Rows are not written to the output if:

- the Box cell is set but the other five cells are all empty
- the Box label starts with `Verification V` (case-insensitive). These are sanity-check
  rows: every other cell should be `V` followed by the rest of the label, and a message
  is printed if not
- Fullness is one of `empty`, `destroyed`, `unassigned`, `not printed`, `printed-unused`
  (case-insensitive). If such a row has data in columns where none is expected, a warning
  is printed
- the Contents cell is empty (a warning is printed)

Warnings go to stdout and do not affect the output.

### What gets written

One output row per surviving input row.

| Snipe-IT column | Value |
|-----------------|-------|
| item Name       | Contents |
| Category        | Category |
| Model name      | Always `Generic-Model` (a Snipe-IT requirement, not a statement about the item) |
| Asset Tag       | `<Box>-<timestamp>-<8-digit row counter>` |
| Location        | Location |
| BoxName         | Box (custom field) |
| everything else | Empty |

The asset tag is synthesised from the run's timestamp, so **it differs on every run**.
TODO: note the consequences for re-importing (duplicates?) and how you handle that.

### Importing into Snipe-IT

- One-off setup: create the `BoxName` custom field and add it to the fieldset used by
  the generic model
- Admin → Import: upload the file, map the columns (especially `BoxName`)
- Check/clean up: categories, locations that did not match, etc.

### Known limitations

- Rows with fewer than six columns will cause a failure
- Locations and categories are passed through as free text; no validation
- No way to specify status, serial number or purchase details

