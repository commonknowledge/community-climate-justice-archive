# API Field Analysis Utility

This utility analyzes the NocoDB API fields by fetching the table schema directly from NocoDB, then compares them against the current application structure to identify new fields that need to be integrated. It efficiently detects multi-select fields used for tagging and provides native NocoDB field type information.

## Features

- **Schema-Based Analysis**: Fetches field definitions directly from NocoDB API (no wasteful table scans)
- **Minimal Data Fetching**: Only fetches one sample record for field values (extremely efficient)
- **Multi-Select Detection**: Identifies multi-select fields used for tagging with their available options
- **Native Field Types**: Shows actual NocoDB field types (MultiSelect, SingleSelect, Links, etc.)
- **Comprehensive Mapping**: Compares against Go structs and template usage
- **Multiple Output Formats**: Console, CSV, and JSON outputs

## Usage

```bash
# Set your NocoDB connection settings
export NOCODB_ENDPOINT=your_endpoint
export NOCODB_API_KEY=your_api_key
export NOCODB_TABLE_ID=your_table_id

# Run the analysis
go run utilities/analyze-api-fields/main.go
```

## Output

The utility generates three types of output:

1. **Console Table**: Immediate overview of all fields with their status and NocoDB types
2. **CSV File**: `api-field-analysis.csv` - Spreadsheet-friendly format with full details
3. **JSON File**: `api-field-analysis.json` - Machine-readable format for automation

## Field Status Types

- **NEW**: Field exists in API but not mapped in application
- **FULLY_MAPPED**: Field is mapped in structs and used in templates
- **MAPPED_NOT_DISPLAYED**: Field is in structs but not shown in templates
- **MAPPED_NOT_USED**: Field is mapped but not used anywhere

## NocoDB Field Types Detected

- **MultiSelect**: Multi-select fields used for tagging (shows available options)
- **SingleSelect**: Single-select dropdown fields (shows available options)
- **Links**: Relationship fields connecting to other tables
- **Attachment**: File/image upload fields
- **DateTime**: Date and time fields
- **LongText**: Multi-line text fields
- **SingleLineText**: Single-line text fields
- **ID**: Auto-increment ID fields
- **URL**: URL fields
- And more...

## Analysis Columns

- **Field Name**: The exact field name from the NocoDB API
- **NocoDB Type**: Native NocoDB field type (MultiSelect, Links, etc.)
- **NocoDB Options**: Available options for select fields (truncated in console)
- **Data Type**: Detected runtime data type (string, number, array, object, etc.)
- **Currently Mapped**: Whether field exists in `NocoDBStoryDTO`
- **Used in Story**: Whether field is mapped to the final `Story` struct
- **Used in Templates**: Whether field appears in HTML templates
- **Status**: Overall classification of the field
- **Sample Value**: Truncated example of the field's content
- **Template Files**: Which template files use this field (CSV/JSON only)

## Example Output

```
FIELD NAME          NOCODB TYPE         DATA TYPE    MAPPED  IN STORY  IN TEMPLATES  STATUS
----------          -----------         ---------    ------  --------  ------------  ------
Themes              MultiSelect         string       true    true      true          FULLY_MAPPED
Type                MultiSelect         string       true    true      true          FULLY_MAPPED
Weather             MultiSelect         string       true    true      true          FULLY_MAPPED
Image / video       Attachment          null         false   false     false         NEW
```

## Multi-Select Fields Summary

The utility specifically identifies and highlights multi-select fields used for tagging:

```
Multi-Select fields (used for tagging):
  - Themes [Tiny Things, Care, Control, Joy, ...]
  - Type [Text, Photo, Drawing, Map, ...]
  - Weather [Sunny, Rainy, Cloudy, ...]
  - Season [Spring, Summer, Autumn, Winter]
```

## What to Do with Results

1. **Review NEW fields**: These are new API fields that may need to be added to your structs
2. **Check Multi-Select fields**: These are tagging fields - ensure they're properly handled as arrays
3. **Review MAPPED_NOT_DISPLAYED**: These fields are available but not shown to users
4. **Consider MAPPED_NOT_USED**: These might be legacy fields that can be removed

## Integration Steps

For NEW fields you want to integrate:

1. Add the field to `NocoDBStoryDTO` in `internal/nocodb/types.go`
2. Add corresponding field to `Story` struct in `data/story.go` if needed
3. Update the conversion logic in `NocoDBRecordToStoryWithClient()`
4. For multi-select fields, use the parsing functions like `ParseThemesFromNocoDB()`
5. Add the field to relevant templates in `templates/` directory
6. Re-run this utility to verify the integration

## Efficiency Benefits

- **No Table Scans**: Uses NocoDB's schema API instead of scanning all records
- **Minimal Data Transfer**: Only fetches one sample record instead of all records
- **Native Type Detection**: Gets actual field types from NocoDB metadata
- **Option Discovery**: Automatically discovers available options for select fields
- **Relationship Mapping**: Identifies link fields and their relationships
- **Fast Execution**: Completes analysis in seconds instead of minutes

This approach is extremely efficient - it gets complete field information from the schema API and only needs one sample record for field values, making it orders of magnitude faster than scanning entire tables.
