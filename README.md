# Flash Search 🔦

Flash is a full-text desktop search engine, designed to help users find their files. Using preprocessing techniques, flash creates an index, allowing searching without having to crawl the file system, substantially reducing search times.

## Installation

Before flash can be installed, ensure that all dependencies are installed by running:

```sh
sudo apt install libmagic-dev
```
and
```sh
sudo apt install libgtk-3-dev libcairo2-dev libglib2.0-dev
```

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

Finally, run the `install` command to set up flash on your system:
```sh
flash install
```

Congrats, you can now use flash! 🎉

Start by adding the directories you want indexed using

```sh
flash add <path-to-dir>
```
Once it has been indexed, the files can be searched by either using
```sh
flash find <search-query>
```
or 
```sh
flash gui
```

## Usage

Once flash has been installed it can be used as follows:

``` sh
flash [command]
```

| Command   | Description                                     | Usage                         |
|-----------|-------------------------------------------------|-------------------------------|
| add       | Adds a file or directory to the watch list      | `flash add <path-to-dir>`     |
| blacklist | Blacklists all files which match a given regex  | `flash blacklist [command]`   |
| daemon    | Used to control the file monitor daemon         | `flash daemon [command]`      |
| find      | Searches the index for a given phrase           | `flash find "<search-query>"` |
| gui       | Opens a graphical search box                   | `flash gui`                   |
| help      | Outputs help for the program                    | `flash help`                  |
| install   | Performs all setup required for flash to run    | `flash install`               |
| remove    | Removes a file or directory from the watch list | `flash remove <path-to dir>`  |
| reset     | Removes all files from the index                | `flash reset`                 |

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
