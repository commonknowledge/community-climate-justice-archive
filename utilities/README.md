# Utilities

This folder contains one-off command-line tools that were created for specific migration or analysis tasks. These are **archived tools** - they served their purpose during development but are no longer part of the regular workflow.

## Tools in This Directory

### consolidate-image-fields

**Status:** Deprecated (migration complete)

This tool was created to merge the `SourceImage` and `Image` fields into a single `ImageVideoSound` field in NocoDB. The migration has been completed successfully.

**Purpose:** One-time data migration to consolidate image attachment fields

**Usage:** The tool is idempotent, so it can be run multiple times safely. However, the migration is complete and this tool is no longer needed.

```bash
go run utilities/consolidate-image-fields/main.go --help
```

### analyze-api-fields

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

## Why These Are Archived

These tools were created for specific tasks during development:
- **Migration tools** like `consolidate-image-fields` completed their one-time job
- **Analysis tools** like `analyze-api-fields` are only needed occasionally

They're kept here rather than deleted because:
- They document how certain migrations were performed
- They can serve as examples for similar future tasks
- Occasionally, analysis tools are still useful

## Do Not Use These for Regular Operations

The main archive generation is handled by:
- `cmd/archive/main.go` - The primary application

Only use tools in this `utilities/` folder if you specifically need them for development or migration tasks.

