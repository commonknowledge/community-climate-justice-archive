# Dudley Climate Justice Archive

## Status

At the time of writing, this contains the second version of the archive, released for Winter 2025.

## Technology Choices

### Go

[Go](https://go.dev/) was chosen for this project because of its simplicity, maintainability, high quality and low carbon footprint. In comparison to higher level languages like Python and JavaScript, Go is a compiled language[^compiled], which makes it more efficient in terms of energy consumption.

### SQLite

The ulimate aim in this project was to use SQLite as the database backend.

[SQLite](https://www.sqlite.org/index.html) was chosen for this project because it is a lightweight, disk-based database[^database]. It allows us to keep all of the archive's data in a single file, making it easier to store and transport. 

In the event of a climate collapse, the database will still be readable and usable and can easily be reproduced.

In comparison to a more traditional database like PostgreSQL, SQLite is also more energy efficient.

The eventual ambition is for the project itself to contain its own backend which uses NocoDB. Currently NocoDB is used as an interface for the backend of the archive, which does ultimately write to a SQLite database stored on disk. At some point we will replace calls to the NocoDB API[^api] with direct reads of an SQLite database, but it was decided that this was too much complexity for now, as the NocoDB interface provided a lot of power.

Accessing the NocoDB database directly from SQLite clients is possible, so the archive does not ultimate depend on NocoDB. All parts of the archive can therefore be replaced.

## How It All Works

This section explains how the archive takes story data from NocoDB and turns it into a complete static website. If you're new to maintaining the archive, this will help you understand how everything fits together.

### What's a Static Site?

Before we dive in, let's be clear about what "static site" means:

A **static site** is just a folder full of regular HTML[^html] files - like the websites from the early days of the internet. When someone visits the archive, their browser just downloads and displays these HTML files directly. There's no database query happening, no server generating pages on-the-fly.

**Think of it like this:**
- **Dynamic sites** (like Facebook): When you visit, the server runs code, fetches data from a database, builds the page right then, and sends it to you. Different every time!
- **Static sites** (like this archive): All the pages are already built and sitting in a folder. When you visit, the server just sends you the pre-made HTML file. Simple!

**Why we use a static site:**
- **Fast**: No waiting for a server to build pages - they're already built
- **Simple**: Just HTML files, so you can host them almost anywhere
- **Resilient**: If something goes wrong with the hosting, you still have all your HTML files. Resilience against all kinds of disaster, especially climate related, is a key goal of this archive
- **Low energy**: Static files use way less server resources than dynamic sites
- **Future-proof**: In 20 and even 100 years, HTML files will still work, even if the database software and all other technologies in the archive are obsolete

**The "Generator" Part:**
We don't write all those HTML files by hand! The archive *generates* them automatically from the NocoDB data and templates. You run the build command, it creates all the HTML files, and then those files get uploaded to a web hosting service called Render so that people can visit the archive with their web browser.

So: **Database + Templates → Build Process → HTML Files → Upload to Render → Live Archive**

### The Big Picture

The archive is a **static site generator**. It takes data we have an makes a static site, as described above. It fetches data from NocoDB, processes images, fills in HTML templates[^templates], and creates a complete website of HTML files. The generated site is then uploaded for hosting. No database or server-side code is needed once the site is built.

```mermaid
graph TB
    A[NocoDB Database<br/>Stories live here] -->|Fetch stories via API| B[Archive Builder<br/>Go program]
    B -->|Process & resize| C[Images]
    B -->|Fill with data| D[HTML Templates]
    C --> E[Generated Website<br/>Just HTML files in a folder!]
    D --> E
    E -->|Upload| F[Render Static Site<br/>Visitors download these HTML files]
    
    style A fill:#e1f5ff
    style B fill:#fff4e1
    style E fill:#e8f5e9
    style F fill:#f3e5f5
```

### The Architecture: Layers

The code is organised into layers, each with a specific job. This keeps things tidy and makes it easier to change one part without breaking others.

```mermaid
graph TB
    subgraph "Entry Point"
        CMD[cmd/archive/main.go<br/>Starts everything]
    end
    
    subgraph "Generation Layer"
        GEN[internal/generate/<br/>Creates HTML pages]
        ASSETS[internal/generate/assets.go<br/>Processes images]
    end
    
    subgraph "Store Layer - The Adapter"
        ADAPTER[internal/store/adapter.go<br/>Defines how to fetch data]
        NOCODB[internal/store/nocodb_adapter.go<br/>Talks to NocoDB]
    end
    
    subgraph "Data Layer"
        TYPES[internal/nocodb/types.go<br/>Translates NocoDB format]
        STORY[data/story.go<br/>Story structs everyone uses]
    end
    
    subgraph "Templates"
        TMPL[templates/<br/>HTML templates]
    end
    
    CMD --> GEN
    CMD --> ASSETS
    GEN --> ADAPTER
    ADAPTER --> NOCODB
    NOCODB --> TYPES
    TYPES --> STORY
    GEN --> TMPL
    
    style CMD fill:#ffebee
    style GEN fill:#e1f5ff
    style ASSETS fill:#e1f5ff
    style ADAPTER fill:#fff9c4
    style NOCODB fill:#fff9c4
    style TYPES fill:#e8f5e9
    style STORY fill:#e8f5e9
    style TMPL fill:#f3e5f5
```

### How Data Flows: From Database to Website

Here's what happens when you run `go run ./cmd/archive`:

```mermaid
flowchart TD
    Start([Run: go run ./cmd/archive]) --> Config[Load settings from .env file]
    Config --> Fetch[Fetch all stories from NocoDB]
    Fetch --> Images[Process images<br/>Resize and convert to WebP]
    Images --> Pages[Generate all HTML pages]
    
    Pages --> Home[Homepage]
    Pages --> Stories[Story pages<br/>One for each story]
    Pages --> Index[Index pages<br/>Themes, types, weather, etc.]
    
    Home --> CSS[Copy CSS and assets]
    Stories --> CSS
    Index --> CSS
    
    CSS --> Done([Done! Files in out/ folder])
    
    style Start fill:#e8f5e9
    style Fetch fill:#e1f5ff
    style Images fill:#fff9c4
    style Pages fill:#ffebee
    style Done fill:#e8f5e9
```

### The Store Layer: The Adapter Pattern

The store layer is like a translator. The rest of the code just asks "give me stories" and doesn't need to know if they're coming from NocoDB, SQLite, or anywhere else.

**Why we have this:**
- We can swap NocoDB for SQLite later without rewriting everything
- The code is cleaner because fetching data is separate from displaying it
- Testing is easier because we can create fake adapters

```go
// The interface - what the adapter promises to do
type DataAdapter interface {
    GetAllStories() ([]data.Story, error)
    GetStoryByID(id string) (data.Story, error)
    GetStoriesForTheme(theme string) ([]data.Story, error)
    // ... and so on
}

// The NocoDB version - how we do it right now
type NocoDBAdapter struct {
    client *nocodb.Client
}

// In the future, we might have:
type SQLiteAdapter struct {
    db *sql.DB
}
```

The rest of the code just uses `GetAdapter()` and doesn't care which adapter it is.

### How Stories Get Transformed

NocoDB sends data in one format, but we need it in another. Here's how a story transforms as it moves through the system:

```mermaid
graph LR
    A["NocoDB JSON<br/>Field: 'Image / video / sound'<br/>Field: 'What was/is/if'"] 
    -->|Parse| B["NocoDBStoryDTO<br/>ImageVideoSound interface<br/>WhatWasIsIf interface"]
    -->|Convert| C["Story Struct<br/>ImageVideoSound string<br/>WhatWasIsIf []WhatWasIsIf"]
    -->|Render| D["HTML Template<br/>.Story.ImageVideoSound<br/>range .Story.WhatWasIsIf"]
    
    style A fill:#ffebee
    style B fill:#fff9c4
    style C fill:#e8f5e9
    style D fill:#f3e5f5
```

**Step by step:**

1. **NocoDB JSON**: Raw data with field names like "Image / video / sound" (notice the spaces)
2. **NocoDBStoryDTO**: A Go struct[^struct] with JSON[^json] tags that match NocoDB's field names exactly
3. **Story Struct**: The proper struct everyone uses, with nice Go field names and proper types
4. **HTML Template**: The template accesses fields like `{{.Story.Finding}}` to display them

The conversion happens in `internal/nocodb/types.go`. It handles tricky things like:
- Converting things like `["Theme A", "Theme B"]` into proper Go structs with URLs and colours
- Parsing attachment JSON into StoryAttachment objects
- Turning `null` into sensible defaults instead of crashing

### Image Processing Pipeline

Images need special treatment. We resize them into different sizes and convert them to WebP[^webp] format (which loads faster and uses less bandwidth).

```mermaid
graph TB
    A[Original Image<br/>photo.jpg<br/>3000×2000px] --> B[Read from images/ folder]
    B --> C{Resize into 3 versions}
    C --> D[Thumbnail<br/>400×267px]
    C --> E[Medium<br/>1024×683px]
    C --> F[Large<br/>2048×1365px]
    D --> G[Convert to WebP]
    E --> G
    F --> G
    G --> H[Save to out/images/processed/]
    H --> I[Templates reference<br/>processed/photo-thumb.webp<br/>processed/photo-medium.webp<br/>processed/photo-large.webp]
    
    style A fill:#ffebee
    style G fill:#e8f5e9
    style I fill:#f3e5f5
```

**Why WebP?**
- Much smaller file sizes than JPG or PNG
- Faster page loading for visitors
- Better for accessibility (people on slow connections)
- More environmentally friendly (less data transfer = less energy)

### The Build Process

When you run the archive, here's what actually happens:

```mermaid
graph TB
    START([Run: go run ./cmd/archive]) --> CONFIG[Load configuration<br/>from .env file]
    CONFIG --> ADAPTER[Initialize NocoDB adapter]
    ADAPTER --> CACHE[Warm the cache<br/>Fetch all stories once]
    CACHE --> IMAGES{Skip images?}
    IMAGES -->|No| PROCESS[Process all images<br/>Resize & convert to WebP]
    IMAGES -->|Yes -s flag| SKIP[Skip image processing]
    PROCESS --> PAGES
    SKIP --> PAGES[Generate all HTML pages]
    
    PAGES --> HOME[Homepage]
    PAGES --> ARCHIVE[Archive page]
    PAGES --> WANDER[Wander page]
    PAGES --> STORIES[Individual story pages<br/>One per story]
    PAGES --> THEMES[Theme index pages]
    PAGES --> TYPES[Type index pages]
    PAGES --> WEATHER[Weather index pages]
    PAGES --> OTHER[Other index pages<br/>Gifted By, Time Period, etc.]
    
    HOME --> CSS
    ARCHIVE --> CSS
    WANDER --> CSS
    STORIES --> CSS
    THEMES --> CSS
    TYPES --> CSS
    WEATHER --> CSS
    OTHER --> CSS
    
    CSS[Copy CSS to out/] --> DONE{Development mode?}
    DONE -->|Yes -d flag| SERVER[Start local server<br/>http://localhost:8080<br/>Watch for template changes]
    DONE -->|No| FINISH([Done! Website in out/ folder])
    
    style START fill:#e8f5e9
    style PROCESS fill:#fff9c4
    style PAGES fill:#e1f5ff
    style SERVER fill:#f3e5f5
    style FINISH fill:#e8f5e9
```

### Development Mode vs Production Mode

**Development Mode** (`-d` flag):
- Builds the website
- Starts a local web server on http://localhost:8080
- Watches for template changes
- Press Enter to regenerate pages
- Perfect for testing changes quickly

**Production Mode** (no flags):
- Builds the website (generates all the HTML files)
- Exits when done
- Render uses this mode when deploying
- The generated `out/` folder (full of HTML files) gets served to visitors by Render

### Key Files and What They Do

If you need to make changes, here's where to look:

| File | What it does |
|------|-------------|
| `cmd/archive/main.go` | Entry point - starts everything |
| `internal/config/config.go` | Loads settings from .env |
| `internal/store/adapter.go` | Defines the adapter interface |
| `internal/store/nocodb_adapter.go` | NocoDB implementation |
| `internal/nocodb/types.go` | Translates NocoDB → Go structs |
| `data/story.go` | Main Story struct & helpers |
| `data/tags.go` | Theme, Type, Weather structs |
| `internal/generate/generate.go` | Creates all HTML pages |
| `internal/generate/assets.go` | Processes images & copies CSS |
| `templates/*.html` | HTML templates |

### Adding New Features: Common Tasks

**Want to add a new field to stories?**
1. Add to `NocoDBStoryDTO` in `internal/nocodb/types.go`
2. Add to `Story` struct in `data/story.go`
3. Map in the conversion function in `internal/nocodb/types.go`
4. Use in template with `{{.Story.NewField}}`

**Want to add a new page type?**
1. Create a template in `templates/`
2. Add a generation function in `internal/generate/generate.go`
3. Call it from `cmd/archive/main.go` in both `generateArchive()` and `hotRegenerate()`

**Want to change how pages look?**
- Edit `css/styles.css` for styling
- Edit templates in `templates/` for structure
- Images in `images/` get processed automatically

### Caching: Keeping Things Fast

The archive caches story data in a JSON file (`debug-cache-nocodb.json`) so it doesn't have to fetch from NocoDB every single time during development. 

If you need fresh data (like after updating stories in NocoDB), just delete the cache file:
```bash
rm debug-cache-nocodb.json
```

The next time you build, it'll fetch everything fresh from NocoDB.

### Troubleshooting Tips

**Images not showing up?**
- Check `images/` folder has the files
- Make sure you're not using `-s` flag (which skips image processing)
- Delete `debug-cache-nocodb.json` and rebuild

**Template changes not appearing?**
- In development mode, press Enter to regenerate
- Or restart the archive

**Story data seems stale?**
- Delete `debug-cache-nocodb.json`
- Rebuild the archive

**Build failing?**
- Check your `.env` file has all the NocoDB settings (environment variables[^envvar])
- Make sure NocoDB is accessible
- Check for error messages in the console

## Deployment

The archive currently deploys to [Render](https://render.com/) as a static site. Render handles the entire build and deployment process automatically.

**How it works:**

When you push commits to the `main` branch, Render automatically:
1. Detects the change in the repository
2. Runs the build command: `go run ./cmd/archive` (production mode)
3. Generates all the HTML files into the `out/` folder
4. Serves these HTML files to visitors

That's it! Render does all the work - building the site and hosting it. No GitHub Actions or manual deployment needed.
## Local Development

### Quick Start (macOS)

We've created an automated setup script that handles everything for you.

1. Download the repository[^repository].

```bash
git clone https://github.com/commonknowledge/community-climate-justice-archive.git
cd community-climate-justice-archive
```

2. Run the setup script.

```bash
bash setup.sh
```

The script will:
- Check for Homebrew[^homebrew] and install it if needed
- Check for Go and prompt you to install it if needed
- Ask you for your NocoDB configuration (endpoint, API key, table ID)
- Build the archive
- Optionally launch it in development mode

That's it! The archive will be running at [http://localhost:8080](http://localhost:8080).

### Manual Setup (macOS)

If you prefer to set things up manually:

1. Install Go.

```bash
brew install go
```

2. Download the repository[^repository].

You can do this on the command line[^commandline] with the following, or use your favoured Git[^git] client.

Its a bit repository, so be patient while it downloads.

```bash
git clone https://github.com/commonknowledge/community-climate-justice-archive.git
```

3. Create your `.env` file.

Copy `env.example` to `.env` and add your NocoDB credentials:

```bash
cp env.example .env
# Then edit .env with your text editor
```

4. Run the archive in development mode.

Enter the repository directory in a terminal[^terminal]. 

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

**Main page templates:**
- [**homepage.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/homepage.html): The main landing page.
- [**story.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/story.html): Displays individual stories with their image, details (type, date, location, etc.), and related stories.
- [**archive.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/archive.html): The main archive page with filtering functionality to browse all stories.
- [**wander.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/wander.html): A randomized/shuffled view of all stories for serendipitous discovery.

**Pages for different ways to organize stories:**
- [**theme-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/theme-index.html): Shows all stories related to a specific theme.
- [**type-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/type-index.html): Shows all stories of a particular type.
- [**weather-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/weather-index.html): Shows all stories that were created in a particular weather condition.
- [**giftedby-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/giftedby-index.html): Shows all stories gifted or co-created by a specific person or organization.
- [**timeperiod-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/timeperiod-index.html): Shows all stories from a specific time period.
- [**whatwasisif-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/whatwasisif-index.html): Shows all stories categorized by "What Was/Is/If" framework.
- [**scalepermanence-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/scalepermanence-index.html): Shows all stories with a specific scale of permanence classification.

**Shared pieces (used by other templates):**
- [**partials/header.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/partials/header.html): The top part of every page with the site title and navigation menu.
- [**partials/footer.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/partials/footer.html): The bottom part of every page with project information and contact details.
- [**partials/stories-list.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/partials/stories-list.html): The code that displays stories in a grid - used by many different pages.
- [**partials/grid-code.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/partials/grid-code.html): The code that makes story previews pop up when you hover over them.

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
- Attachment: Raw JSON marshalling is used (see `ImageVideoSound` field mapping in `NocoDBRecordToStoryWithClient`)

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

## Deprecated Tools

### Image Field Consolidation Tool

**Status: DEPRECATED - No longer needed**

The `cmd/consolidate-image-fields/` tool was created to migrate data from legacy `SourceImage` and `Image` fields into the unified `ImageVideoSound` field. This migration has been completed.

**Important Notes:**
- The tool is **idempotent** and safe to run multiple times
- Running it now will not cause any issues as it will detect no changes are needed
- The tool is kept for historical reference and potential future migrations
- All stories now use only the `ImageVideoSound` field for attachments

**Usage (if needed):**
```bash
go build -o consolidate-image-fields ./cmd/consolidate-image-fields
./consolidate-image-fields --checksum  # Verify consolidation is complete
```

---

## Technical Terms Glossary

[^compiled]: **Compiled language**: A programming language where code is translated into machine instructions before running, making it faster and more efficient. [Learn more](https://www.freecodecamp.org/news/compiled-versus-interpreted-languages/)

[^database]: **Database**: A structured system for storing and organising data. Think of it like a sophisticated filing cabinet for information. [Learn more](https://www.oracle.com/uk/database/what-is-database/)

[^api]: **API (Application Programming Interface)**: A way for programs to talk to each other. NocoDB's API lets us request story data over the internet. [Learn more](https://www.howtogeek.com/343877/what-is-an-api/)

[^html]: **HTML (HyperText Markup Language)**: The standard language for creating web pages. HTML files tell browsers how to display content. [Learn more](https://developer.mozilla.org/en-US/docs/Learn/Getting_started_with_the_web/HTML_basics)

[^templates]: **Templates**: Pre-designed files with placeholders that get filled in with actual data. Like a form letter where you fill in the name and address. [Learn more](https://en.wikipedia.org/wiki/Template_processor)

[^struct]: **Struct (Structure)**: In Go, a struct is a custom data type that groups related pieces of information together. Like a form with labelled fields. [Learn more](https://gobyexample.com/structs)

[^json]: **JSON (JavaScript Object Notation)**: A text format for storing and transmitting data that's easy for both humans and computers to read. [Learn more](https://www.json.org/json-en.html)

[^webp]: **WebP**: A modern image format that creates smaller file sizes than JPEG or PNG while maintaining quality, making pages load faster. [Learn more](https://developers.google.com/speed/webp)

[^repository]: **Repository**: A storage location for your project files and their entire revision history. Often shortened to "repo". [Learn more](https://docs.github.com/en/repositories/creating-and-managing-repositories/about-repositories)

[^commandline]: **Command line**: A text-based interface where you type commands to interact with your computer, rather than clicking with a mouse. [Learn more](https://tutorial.djangogirls.org/en/intro_to_command_line/)

[^git]: **Git**: A version control system that tracks changes to files over time, letting you collaborate with others and revert to previous versions if needed. [Learn more](https://www.atlassian.com/git/tutorials/what-is-git)

[^terminal]: **Terminal**: A program that provides the command line interface. On macOS, it's called "Terminal"; on Windows, "Command Prompt" or "PowerShell". [Learn more](https://www.codecademy.com/article/command-line-commands)

[^envvar]: **Environment variables**: Configuration settings stored outside your code, often in a `.env` file. They keep sensitive information like passwords separate from the code. [Learn more](https://www.twilio.com/en-us/blog/environment-variables-python)

[^homebrew]: **Homebrew**: A package manager for macOS that makes installing and managing software much easier. Think of it as an app store for developer tools. [Learn more](https://brew.sh/)