# Dudley Climate Justice Archive

## Status

At the time of writing, this contains the first version of the archive. This is intended for release in Spring 2025.

## Technology Choices

### Go

[Go](https://go.dev/) was chosen for this project because of its simplicity, maintainability, high quality and low carbon footprint. In comparison to higher level languages like Python and JavaScript, Go is a compiled language, which makes it more efficient in terms of energy consumption.

### SQLite

The ulimate aim in this project was to use SQLite as the database backend.

[SQLite](https://www.sqlite.org/index.html) was chosen for this project because it is a lightweight, disk-based database. It allows us to keep all of the archive's data in a single file, making it easier to store and transport. 

In the event of a climate collapse, the database will still be readable and usable and can easily be reproduced.

In comparison to a more traditional database like PostgreSQL, SQLite is also more energy efficient.

The eventual ambition is for the project itself to contain its own backend which uses NocoDB. Currently NocoDB is used as an interface for the backend of the archive, which does ultimately write to a SQLite database stored on disk. At some point we will replace calls to the NocoDB API with direct reads of an SQLite database, but it was decided that this was too much complexity for now, as the NocoDB interface provided a lot of power.

Accessing the NocoDB database directly from SQLite clients is possible, so the archive does not ultimate depend on NocoDB. All parts of the archive can therefore be replaced.

## Deployment

The archive currently deploys to [GitHub Pages](https://pages.github.com/). This is done on every commit to the `main` branch.

The deployment process is encapsulated in a [GitHub Action](https://github.com/features/actions), the process of which is in `.github/workflows/deploy.yml`.

## Local Development

### Creating a SQLite export of the Airtable database

1. You will need three things:
 - Access to the Dudley People's School of Climate Justice Airtable. You can request this from someone involved in the project.
 - An Airtable personal access token, [which you can create here](https://airtable.com/create/tokens). Assigning the token the `data.records:read` scope is sufficient.
 - The Airtable table ID, which you can find by going to the table in Airtable and [copying the ID from the URL](https://support.airtable.com/docs/finding-airtable-ids).

2. Install [airtable-to-sqlite](https://github.com/kanedata/airtable-to-sqlite) following the instructions in the README.

3. Run the following command to export the table to a SQLite database:

```bash
airtable-to-sqlite --personal-access-token <your-token> --output airtable-export.db <your-table-id>
```

This will create a file called `airtable-export.db` in the current directory. The `dump-images-from-airtable` program expects this file to be present in the root of the repository.

### Using the `dump-images-from-airtable.go`

This program will dump all images out of an Airtable export and write them to disk.

#### macOS

1. In order to run the programs in this repository, you'll need to install Go.

```bash
brew install go
```

2. Download the repository.

```bash
git clone https://github.com/commonknowledge/community-climate-justice-archive.git
```

3. Run the program.

```bash
go run cmd/cli/dump-images-from-airtable.go
```

As we have already exported all the images into this repository, the only thing that will happen is a series of log messages, noting that the images are already present and the download has been skipped.

### Working with the archive

#### Compiling the archive

##### macOS

1. Install Go.

```bash
brew install go
```

2. Download the repository.

You can do this on the command line with the following, or use your favoured Git client.

Its a bit repository, so be patient while it downloads.

```bash
git clone https://github.com/commonknowledge/community-climate-justice-archive.git
```

3. Run the archive in development mode.

Enter the repository directory in a terminal. 

```bash
cd community-climate-justice-archive
```

Then start the archive in development mode. This will initially compile the archive.

```bash
go run ./cmd/archive -d
```

This will launch a development server at [http://localhost:8080](http://localhost:8080).

## Making changes

### Styling in CSS

Edit the file `css/styles.css`.

If you are running the archive in development mode, then the CSS will automatically be copied to the directory that the development webserver serves.

When you've made your change, simply refresh the page to see the effect.

### Templating in HTML

The HTML templates are located in the `templates` directory. 

If you are running the archive in development mode, press enter to regenerate the archive. Your changes in these HTML templates will then be picked up.

Templates use Go's standard `html/template` package syntax.

You can find documentation for Go's templating language at:
- [html/template package documentation](https://pkg.go.dev/html/template)
- [text/template package documentation](https://pkg.go.dev/text/template) (html/template builds on this)

Key template features include:
- `{{.Something}}` for outputting information that the main program gives to templates.
- `{{if .SomeOtherThing}}...{{end}}` for making decisions and displaying different things based on these.
- `{{range .Stories}}...{{end}}` for looping over a group of things, in the case of the archive mostly stories.

####  Template files overview

The archive templates should be straight forward to understand what they do, but to start you off, here is a brief description of each of them.

- [**homepage.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/homepage.html): The main landing page.
- [**story.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/story.html): Displays individual stories with their image, details (type, date, location, etc.), and related stories.
- [**theme-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/theme-index.html): Shows all stories related to a specific theme.
- [**type-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/type-index.html): Shows all stories of a particular type.
- [**weather-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/weather-index.html): Shows all stories that were creating in a particular weather condition.

### Adding New Fields to Story Templates

When you need to add a **simple field** (text, date, number, etc.) from NocoDB to display in the individual story template, follow these steps:

> **Note**: These instructions are for simple field types only. MultiSelect, Links, and LinkToAnotherRecord fields require different treatment and specialized parsing logic.

#### 1. Add Field to NocoDBStoryDTO Struct
**File**: `internal/nocodb/types.go`

Add the new field to the `NocoDBStoryDTO` struct with the proper JSON tag matching the NocoDB field name:

```go
type NocoDBStoryDTO struct {
    // ... existing fields ...
    NewField interface{} `json:"New Field Name"`
}
```

#### 2. Add Field to Story Struct
**File**: `data/story.go`

Add the corresponding field to the `Story` struct (around lines 27-63):

```go
type Story struct {
    // ... existing fields ...
    NewField string
}
```

#### 3. Map Field in Conversion Function
**File**: `internal/nocodb/types.go`

Update the `NocoDBRecordToStoryWithClient` function (around lines 89-125) to map from DTO to Story:

```go
story := data.Story{
    // ... existing mappings ...
    NewField: toString(dto.NewField),
}
```

#### 4. Add Field to HTML Template
**File**: `templates/story.html`

Add the field display in the template using Go template syntax:

```html
<!-- Simple field display -->
{{.Story.NewField}}

<!-- Conditional display -->
{{if .Story.NewField}}
    <p>{{.Story.NewField}}</p>
{{end}}

<!-- With HTML structure -->
<div class="new-field">
    <label>New Field:</label>
    <span>{{.Story.NewField}}</span>
</div>
```

#### 5. Simple Field Types Supported

This process works for these NocoDB field types:
- **SingleLineText** - Maps to `string`
- **LongText** - Maps to `string` 
- **Number** - Maps to `string` (converted via `toString()`)
- **Date** - Maps to `string`
- **DateTime** - Maps to `string`
- **URL** - Maps to `string`
- **Email** - Maps to `string`
- **PhoneNumber** - Maps to `string`

#### 5. Complex Field Types (Different Process Required)

The following field types require **specialized handling** and are **not covered** by these instructions:
- **MultiSelect** - Requires custom parsing to convert to `[]Theme`, `[]Type`, or `[]Weather` structs
- **Links** - Requires relationship resolution and connection caching
- **LinkToAnotherRecord** - Requires relationship resolution and connection caching
- **Attachment** - Requires image processing and JSON conversion
- **Checkbox** - Requires boolean conversion
- **SingleSelect** - May require enum/option handling

For these complex types, refer to existing examples in the codebase:
- MultiSelect: See `ParseThemesFromNocoDB()`, `ParseTypesFromNocoDB()`, `ParseWeatherFromNocoDB()`
- Links/LinkToAnotherRecord: See `fetchStoryConnectionsDirect()` 
- Attachment: See `ParseAttachmentsFromNocoDB()`

#### 6. Update Filtering Data (optional)
**File**: `internal/generate/generate.go`

If the field should be available for client-side filtering, add it to the `StoryData` struct (around lines 82-96):

```go
type StoryData struct {
    // ... existing fields ...
    NewField string `json:"newField"`
}
```

#### 7. Verify Field Integration

Run the field analysis utility to verify the field shows as `FULLY_MAPPED`:

```bash
go run cmd/analyze-api-fields/main.go
```

The field should appear with `processing_category: "standard"` and `status: "FULLY_MAPPED"` if properly integrated.

#### Data Flow Summary

The data flows through these components:
```
NocoDB API → NocoDBStoryDTO → Story → StoryPage → story.html template
```

1. **NocoDB API** returns raw field data
2. **NocoDBStoryDTO** receives and structures the raw data  
3. **Story struct** gets converted, typed data
4. **StoryPage** wraps Story for template context
5. **story.html** template renders the field using `{{.Story.FieldName}}`

#### Key Files to Modify

- `internal/nocodb/types.go` - DTO struct + conversion logic
- `data/story.go` - Story struct definition  
- `templates/story.html` - Template display
- `internal/generate/generate.go` - (only if field used in filtering/JSON export)

### Adding New MultiSelect Fields (Tags/Categories)

When you need to add a **MultiSelect field** from NocoDB (like Themes, Types, or Weather), follow this comprehensive process. MultiSelect fields are used for tagging and categorization, requiring specialized handling.

> **Note**: This process is for MultiSelect fields that function as tags/categories. For other complex field types (Links, LinkToAnotherRecord, Attachments), refer to existing examples in the codebase.

#### 1. Add Field to NocoDBStoryDTO Struct
**File**: `internal/nocodb/types.go`

Add the new field to the `NocoDBStoryDTO` struct:

```go
type NocoDBStoryDTO struct {
    // ... existing fields ...
    NewMultiSelectField interface{} `json:"New Field Name"`
}
```

#### 2. Create the Field Type Struct
**File**: `data/story.go`

Create a new struct to represent individual tag/category items:

```go
type NewFieldType struct {
    Title  string  // Display name
    URL    string  // Generated URL slug
    Colour string  // Generated hex color
}
```

#### 3. Add Field to Story Struct
**File**: `data/story.go`

Add the field as a slice of the new type:

```go
type Story struct {
    // ... existing fields ...
    NewMultiSelectField []NewFieldType
}
```

#### 4. Create Parsing Function
**File**: `internal/nocodb/types.go`

Create a parsing function similar to existing ones (`ParseThemesFromNocoDB`, `ParseTypesFromNocoDB`, `ParseWeatherFromNocoDB`):

```go
func ParseNewFieldFromNocoDB(field interface{}) ([]data.NewFieldType, error) {
    if field == nil {
        return []data.NewFieldType{}, nil
    }

    fieldStr, ok := field.(string)
    if !ok {
        return []data.NewFieldType{}, fmt.Errorf("expected string, got %T", field)
    }

    if fieldStr == "" {
        return []data.NewFieldType{}, nil
    }

    // Split comma-separated values from NocoDB
    items := strings.Split(fieldStr, ",")
    var result []data.NewFieldType

    for _, item := range items {
        item = strings.TrimSpace(item)
        if item != "" {
            result = append(result, data.NewFieldType{
                Title:  item,
                URL:    util.Slugify(item),
                Colour: data.TitleToHexColor(item),
            })
        }
    }

    return result, nil
}
```

#### 5. Add Conversion Mapping
**File**: `internal/nocodb/types.go`

Update the `NocoDBRecordToStoryWithClient` function:

```go
func NocoDBRecordToStoryWithClient(record map[string]interface{}, client *Client) (data.Story, error) {
    // ... existing DTO conversion ...

    // Convert new field
    newField, err := ParseNewFieldFromNocoDB(dto.NewMultiSelectField)
    if err != nil {
        log.Printf("Warning: failed to parse new field: %v", err)
        newField = []data.NewFieldType{}
    }

    story := data.Story{
        // ... existing mappings ...
        NewMultiSelectField: newField,
    }

    return story, nil
}
```

#### 6. Add to Template Display
**File**: `templates/story.html`

Add the field display using the tag pattern:

```html
{{if $story.NewMultiSelectField}}
<div class="story-tags">
    <span class="story-tags-label">New Field:</span>
    {{range $story.NewMultiSelectField}}
    <a href="/newfield/{{.URL}}.html" class="tag" style="background-color: {{.Colour}};">{{.Title}}</a>
    {{end}}
</div>
{{end}}
```

#### 7. Add Store Functions
**File**: `internal/store/` (add to appropriate adapter file)

Add functions to retrieve and filter by the new field:

```go
func (s *SQLiteAdapter) GetNewFieldTypes() []data.NewFieldType {
    // Get all unique values for the new field
    // Similar to GetThemes(), GetTypes(), GetWeather()
}

func (s *SQLiteAdapter) GetStoriesForNewFieldType(fieldValue string) ([]data.Story, error) {
    // Filter stories by the new field value
    // Similar to GetStoriesForTheme(), GetStoriesForType(), GetStoriesForWeather()
}
```

#### 8. Add Index Page Generation
**File**: `internal/generate/generate.go`

Add a function to generate individual pages for each field value:

```go
func WriteNewFieldIndexPages(stories []data.Story, store store.Adapter) error {
    newFieldTypes := store.GetNewFieldTypes()
    
    for _, fieldType := range newFieldTypes {
        fieldStories, err := store.GetStoriesForNewFieldType(fieldType.Title)
        if err != nil {
            return fmt.Errorf("failed to get stories for new field %s: %w", fieldType.Title, err)
        }

        page := data.Page{
            Title:       fieldType.Title,
            Stories:     fieldStories,
            NewFieldType: &fieldType,
        }

        filename := fmt.Sprintf("newfield/%s.html", fieldType.URL)
        if err := writePageToFile(filename, "newfield-index.html", page); err != nil {
            return fmt.Errorf("failed to write new field page %s: %w", filename, err)
        }
    }

    return nil
}
```

#### 9. Create Index Template
**File**: `templates/newfield-index.html`

Create a template similar to `theme-index.html`, `type-index.html`, `weather-index.html`:

```html
<!DOCTYPE html>
<html lang="en">
<head>
    <title>{{.Title}} - Community Climate Justice Archive</title>
    <!-- ... head content similar to other index templates ... -->
</head>
<body>
    <div class="page-container">
        <h1>{{.Title}}</h1>
        <p>Stories tagged with "{{.Title}}"</p>
        
        <div class="stories-grid">
            {{range .Stories}}
            <!-- ... story display similar to other index templates ... -->
            {{end}}
        </div>
    </div>
</body>
</html>
```

#### 10. Add CSS Styling
**File**: `css/styles.css`

Ensure the new field tags have consistent styling with existing tags:

```css
/* New field tags should inherit existing .tag styles */
.story-tags .tag {
    /* Existing tag styles will apply */
}
```

#### 11. Integrate Index Generation into Main Build Process
**File**: `cmd/archive/main.go`

Add calls to your new index generation functions in both `generateArchive()` and `hotRegenerate()` functions:

```go
// In generateArchive() function, after WriteWeatherIndexes():
if err := generate.WriteNewFieldIndexPages(); err != nil {
    return fmt.Errorf("failed to write new field indexes: %v", err)
}

// In hotRegenerate() function, after WriteWeatherIndexes():
if err := generate.WriteNewFieldIndexPages(); err != nil {
    return fmt.Errorf("failed to write new field indexes: %v", err)
}
```

**Important**: Without this step, the index pages won't be generated and the tag links won't work!






