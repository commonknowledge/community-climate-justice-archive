# community-climate-justice-archive

## Technology Choices

### Go

[Go](https://go.dev/) was chosen for this project because of its simplicity, maintainability, and low carbon footprint. In comparison to higher level languages like Python and JavaScript, Go is a compiled language, which makes it more efficient in terms of energy consumption.

### SQLite

[SQLite](https://www.sqlite.org/index.html) was chosen for this project because it is a lightweight, disk-based database. It allows us to keep all of the archive's data in a single file, making it easier to store and transport. 

In the event of a climate collapse, the database will still be readable and usable and can easily be reproduced.

In comparison to a more traditional database like PostgreSQL, SQLite is also more energy efficient.

## Local Development

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
go run main.go
```



