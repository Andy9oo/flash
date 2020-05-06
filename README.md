# Flash Search 🔦

Flash is a full-text desktop search engine, designed to help users find their files. Using preprocessing techniques, flash creates an index, allowing searching without having to crawl the filesystem, substantially reducing search times.

## Installation
Clone the repository:
```sh
git clone git@git.cs.sun.ac.za:20721900/flash.git
```

Move into the flash directory:
```sh
cd flash
```

Install the program:
```sh
go install
```

Congrats, you can now use flash! 🎉

## Usage
Once flash has been installed it can be used as follows:
```sh
flash [command]
```

Where `[command]` is one of the commands given in the following table


### Commands
| Command | Description                              | Usage                         |
|---------|------------------------------------------|-------------------------------|
| build   | Builds the index for the given directory | `flash build <path-to dir>`   |
| find    | Searches the index for a given phrase    | `flash find "<search-query>"` |
| help    | Outputs help for the program             | `flash help`                  |

**Note:** Currently, only a single directory (and it's subdirectories) can be indexed at once. If an index is built for a different directory, the original index will be deleted. 

## Development
To edit or build the code yourself, simply clone the repository as shown above.

A binary can be built using:
```sh
go build main.go
```

Otherwise, the project can be built and run in a single command, using:
```sh
go run main.go [command]
```
Where `[command]` is one of the commands described above.

## Authors
- **Andrew Cullis** - Stellenbosch University   
