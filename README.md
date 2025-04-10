# Dudley Climate Justice Archive

## Status

At the time of writing, this contains the first version of the archive, intended for release in Spring 2025.

## Technology Choices

### Go

[Go](https://go.dev/) was chosen for this project because of its simplicity, maintainability, high quality and low carbon footprint. In comparison to higher level languages like Python and JavaScript, Go is a compiled language, which makes it more efficient in terms of energy consumption.

### SQLite

[SQLite](https://www.sqlite.org/index.html) was chosen for this project because it is a lightweight, disk-based database. It allows us to keep all of the archive's data in a single file, making it easier to store and transport. 

In the event of a climate collapse, the database will still be readable and usable and can easily be reproduced.

In comparison to a more traditional database like PostgreSQL, SQLite is also more energy efficient.

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

#### Making changes

##### Styling in CSS

Edit the file `css/styles.css`.

If you are running the archive in development mode, then the CSS will automatically be copied to the directory that the development webserver serves.

##### Templating in HTML

The HTML templates are located in the `templates` directory. They use Go's standard `html/template` package syntax.

You can find documentation for Go's templating language at:
- [html/template package documentation](https://pkg.go.dev/html/template)
- [text/template package documentation](https://pkg.go.dev/text/template) (html/template builds on this)

Key template features include:
- `{{.Something}}` for outputting information that the main program gives to templates.
- `{{if .SomeOtherThing}}...{{end}}` for making decisions and displaying different things based on these
- `{{range .Stories}}...{{end}}` for looping over a group of things, in the case of the archive mostly stories.

If you are running the archive in development mode, press enter to regenerate the archive. Your changes in these HTML templates will then be picked up.

#####  Template files overview

The archive templates should be straight forward to understand what they do, but to start you off, here is a brief description of each of them.

- [**homepage.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/homepage.html): The main landing page.
- [**story.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/story.html): Displays individual stories with their image, details (type, date, location, etc.), and related stories.
- [**theme-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/theme-index.html): Shows all stories related to a specific theme.
- [**type-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/type-index.html): Shows all stories of a particular type.
- [**weather-index.html**](https://github.com/commonknowledge/community-climate-justice-archive/blob/main/templates/weather-index.html): Shows all stories that were creating in a particular weather condition.





