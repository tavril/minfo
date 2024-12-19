# minfo

## What is minfo ?

`minfo` fetches some information about the computer/OS
it is running on and display it, alongside an ASCII logo.

Information fetched contains:

- User, Hostname,
- OS,
- computer model, CPU, GPU, Memory,
- Disk (only disk hosting `/`),
- Battery,
- Displays,
- Terminal program being used,
- Software (including homebrew),
- Public IP

## Install

Install `minfo` with homebrew:

```text
brew tap tavril/tap
brew install minfo
```

Or directly with (without adding the tap):

```text
brew install tavril/tap/minfo
```

## Usage

```text
minfo -h
Usage:
  -h  Show help
  -j  Output in JSON format instead of displaying logo
  -l  Display the ASCII art logo (default true)
  -n  Don't use/update cache
  -r  Refresh cache (or create it if it doesn't exist)
  -v  Show version
```

### Cache file

There is a cache file (`~/.minfo-cache.json`) which caches the computer model, CPU, GPU and memory information (as these information are unlikely to change...).

- You can ask the tool to not use the cache with command line parameter `-n`.
- You can ask the tool to refresh the cache with command line parameter `-r`.

### Display / Hide Logo

You can decide not to display the Apple logo with command line parameter `-l=false`.

### JSON output

You can output JSON instead of text by using command line parameter `-j`.

## TODO

- Have a configuration file where the user can specify what information to display.
