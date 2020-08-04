# Flash Search 🔦

Flash is a full-text desktop search engine, designed to help users find their files. Using preprocessing techniques, flash creates an index, allowing searching without having to crawl the filesystem, substantially reducing search times.

## Installation

Clone the repository:

``` sh
git clone git@git.cs.sun.ac.za:20721900/flash.git
```

Move into the flash directory:

``` sh
cd flash
```

Install the program:

``` sh
go install
```

Congrats, you can now use flash! 🎉

## Usage

Once flash has been installed it can be used as follows:

``` sh
flash [command]
```

| Command | Description                                      | Usage                         |
|---------|--------------------------------------------------|-------------------------------|
| add     | Adds a file or directory to the watch list       | `flash add <path-to dir>` |
| find    | Searches the index for a given phrase            | `flash find "<search-query>"` |
| daemon  | Used to control the file monitor daemon          | `flash daemon [command]` |
| remove  | Removes a file or directory from the watch list  | `flash remove <path-to dir>` |
| reset   | Removes all files from the index                 | `flash reset` |
| help    | Outputs help for the program                     | `flash help` |

## Development

To edit or build the code yourself, simply clone the repository as shown above.
A binary can be built using:

``` sh
go build main.go
```

Otherwise, the project can be built and run in a single command, using:

``` sh
go run main.go [command]
```
Where `[command]` is one of the commands described above.

## Authors

* **Andrew Cullis** - Stellenbosch University   
