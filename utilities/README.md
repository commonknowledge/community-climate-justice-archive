# Utilities

This folder contains optional maintenance utilities that are separate from the normal archive build.

## analyze-api-fields

**Status:** For occasional use when adding new fields

This utility analyzes the NocoDB table schema and compares it against the application's Go structs and template usage. It helps identify:
- New fields added to NocoDB that need to be integrated
- Fields that are mapped but not used in templates
- Multi-select fields with their available options

**Purpose:** Schema analysis and field mapping verification

**Usage:** Run when you've added new fields to NocoDB and need to see what changes are required in the code:

```bash
go run utilities/analyze-api-fields/main.go
```

The tool generates three output formats:
1. Console output with color-coded status
2. `api-field-analysis.csv` - Spreadsheet for review
3. `api-field-analysis.json` - Machine-readable format

See `utilities/analyze-api-fields/README.md` for detailed documentation.

## Regular Operations

The main archive generation is handled by:
- `cmd/archive/main.go` - The primary application

Use the utilities in this folder only for maintenance and schema analysis work.
