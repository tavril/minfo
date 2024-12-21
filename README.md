# minfo

## What is minfo ?

`minfo` fetches some information about the computer/OS
it is running on and display it.
It can also display an ASCII colored Apple logo.

Information that can be fetched can be listed by calling `minfo -i`.

## Installation

Install `minfo` with homebrew:

```text
brew tap tavril/tap
brew install minfo
```

Or directly with (without adding the tap):

```text
brew install tavril/tap/minfo
```

You can also compile from source:

- Clone the repository.
- Once inside the directory of the repository, run

```shell
make
```

Then launch

```text
./minfo
```

## Usage

```text
minfo -h
Usage:
  -c string
     Path to the configuration file (default "~/.config/minfo.yml")
  -i  Display all available information to display
  -h  Show help
  -j  Output in JSON format instead of displaying logo
  -l  Display the ASCII art logo (default true)
  -n  Don't use/update cache
  -r  Refresh cache (or create it if it doesn't exist)
  -v  Show version
```

### Cache file

There is a cache file which caches the computer model, CPU, GPU, memory,
and serial number (as these information are unlikely to change...).
By default, the cache is located at `~/.minfo-cache.json`. You can change
the location of the cache file in the configuration file.

- You can ask the tool to not use the cache with command line parameter `-n`.
- You can ask the tool to refresh the cache with command line parameter `-r`.

### Display / Hide Logo

You can decide not to display the Apple logo with command line parameter `-l=false`.

### JSON output

You can output JSON instead of text by using command line parameter `-j`.

## Configuration file

Configuration file is optional.
It defines:

- Location of the cache file,
- Items to be displayed.

By default, the tool will look for a configuration file located at `~/.config/minfo.yml`,
but you can specify another location with command line parameter `-c <path_to_file>`.

## TODO

- Make it work on x86_64
