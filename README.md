# minfo

## What is minfo ?

minfo fetches some information about the computer/OS
it is running on and display it, alongside an ASCII
logo.
Information fetched contains:

- User, Hostname,
- OS,
- computer model, CPU, GPU, Memory
- Disk (only disk hosting `/`)
- Battery
- Displays (up to 2 displays)
- Software (including homebrew)
- Public IP

## Install

```shell
brew tap tavril/tap
brew install tavril/tap/minfo
```

## Usage

```shell
minfo -h
Usage:
  -h  Show help
  -j  Output in JSON format instead of displaying logo
  -l  Don't display the ASCII art logo
  -n  Don't use/update cache
  -r  Refresh cache (or create it if it doesn't exist)
```

## Cache file

There is a cache file (`~/.minfo-cache.json`) which
caches the computer model, cpu, gpu and memory information.
These information are unlikely to change
