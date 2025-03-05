# Dudley Climate Justice Archive

## Status

At the time of writing, this repository contains small programs used to bootstrap the archive. Including downloading images from Airtable and storing them in a local SQLite database.

## Technology Choices

### Go

[Go](https://go.dev/) was chosen for this project because of its simplicity, maintainability, and low carbon footprint. In comparison to higher level languages like Python and JavaScript, Go is a compiled language, which makes it more efficient in terms of energy consumption.

### SQLite

[SQLite](https://www.sqlite.org/index.html) was chosen for this project because it is a lightweight, disk-based database. It allows us to keep all of the archive's data in a single file, making it easier to store and transport. 

In the event of a climate collapse, the database will still be readable and usable and can easily be reproduced.

In comparison to a more traditional database like PostgreSQL, SQLite is also more energy efficient.

## Local Development

At the moment, the only program is `dump-images-from-airtable`, in the `dump-images-from-airtable.go` file.

This takes a SQLite export of the Airtable database and downloads all of the images to the `images/` directory in this repository. It also stores the images in the database.

### Creating a SQLite export of the Airtable database

1. You will need three things:
 - Access to the Dudley People's School of Climate Justice Airtable. You can request this from someone involved in the project.
 - An Airtable personal access token, [which you can create here](https://airtable.com/create/tokens). Assigning the token the `data.records:read` scope is sufficient.
 - The Airtable table ID, which you can find by going to the table in Airtable and [copying the ID from the URL](https://support.airtable.com/docs/finding-airtable-ids).

2. Install [airtable-to-sqlite](https://github.com/simonw/airtable-to-sqlite) following the instructions in the README.

3. Run the following command to export the table to a SQLite database:

```bash
airtable-to-sqlite --personal-access-token <your-token> --table-id <your-table-id> --output airtable-export.db
```

This will create a file called `airtable-export.db` in the current directory. The `dump-images-from-airtable` program expects this file to be present in the root of the repository.

### Mac OS X

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
go run dump-images-from-airtable.go
```

As we have already exported all the images into this repository, the only thing that will happen is a series of log messages, noting that the images are already present and the download has been skipped.

