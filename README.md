# minfo

## What is minfo ?

`minfo` is a tool which displays informatino about your computer/OS.
It only works on macOS.

Information is displayed in plain text, with an ASCII art logo.
You can display the information without the logo, or just in JSON.

You can list all the information that `minfo` can fetch by running `minfo -i`.

## Installation

Install `minfo` with homebrew:

```text
brew tap tavril/tap
brew install minfo
```

Or directly with `brew install tavril/tap/minfo`.

You can also compile from source:

- Clone the repository.
- Once inside the directory of the repository, run

```shell
make
```

Then run minfo:

```text
./minfo
```

## Usage

```text
Usage:
    %s [--config <path>] [-j|--json] [-i|--items] [-v|--version]
    %s [-r|--refresh[=false]] [-c|--cache[=false]] [-l|--logo[=false]]

Options:
    --config <path>          Path to the configuration file (default: ~/.config/minfo.yaml).
    -l, --logo[=false]       Display the ASCII art logo (default: true).
    -j, --json[=false]       Display information in JSON instead of plain text (default: false).
    -c, --cache[=false]      Use cache file (default: true).
    -r, --refresh[=false]    Refresh the cache file (default: false).
    -i, --items              Display all available information to display and exit.
    -v, --version            Show version and exit.
    -h, --help               Show this help message and exit.

Having a configuration file is optional. If you don't provide the --config parameter,
minfo will look for a configuration file at the default path.
If the configuration file does not exist, minfo will use the default values.

--refresh=false and --cache=true are mutually exclusive.
```

### Cache file

There is a cache file (by default `~/.minfo-cache.json`) which caches some information
that are unlikely to change: computer model, serial number, CPU, GPU, memory.
Location of the cache file can be customized in the configuration file.

- You can ask the tool to not use the cache with command line parameter `--cache`.
- You can ask the tool to refresh the cache with command line parameter `--refresh`.

`--refresh=true` and `--cache=false` are mutualy exclusive.

### Logo

You can decide not to display the Apple logo with command line parameter `--logo=false`.
Logo is displayed in color, and support 16/256colors terminals.

### JSON output

You can output JSON instead of text by using command line parameter `--json`.

## Configuration file

Configuration file is optional. If no configuration file exist, default choices will be made.
In the configuration file, you can define

- Location of the cache file,
- Items to be displayed.

Choose the list of items to be displayed among the items listed when running `minfo --items`.

By default, the tool will look for a configuration file located at `~/.config/minfo.yml`,
but you can specify another location with command line parameter `--config <path_to_file>`.

## Examples

```text
$ minfo
                                 User           John Doe (jdoe)
                    ##           Hostname       jdoe-laptop
                  ####           OS             macOS Sequoia 15.2 (24C101) Darwin 24.2.0
                #####            macOS SIP      Enabled
               ####              Serial         XXXXXXXXXX
      ########   ############    Model          MacBook Pro 16-inch (Nov 2024) Z1FW0008GSM/A
    ##########################   CPU            Apple M4 Max 16 cores (12 P and 4 E)
  ###########################    GPU            40 cores
  ##########################     Memory         64 GB LPDDR5
 ##########################      Disk           2.00 TB (1.14 TB available)
 ##########################      Disk SMART     Verified
 ###########################     Battery        94% (discharging) | 100% capacity
  ############################   Battery health Good
  #############################  Display #1     3456 x 2234 | 1728 x 1117 @ 120 Hz
   ############################  Terminal       iTerm.app
     ########################    Software       65 Apps | 227 Formulae | 37 Casks
      ######################     Public IP      178.195.102.237 (Switzerland)
        #######    #######       Uptime         1 days, 19 hours
                                 Date/Time      Sun, 22 Dec 2024 16:58:33 CET
````

```text
$ minfo -j
{
  "model": {
    "name": "MacBook Pro",
    "sub_name": "16-inch",
    "date": "Nov 2024",
    "number": "Z1FW0008GSM/A"
  },
  "cpu": {
    "model": "Apple M4 Max",
    "cores": 16,
    "performance_cores": 12,
    "efficiency_cores": 4
  },
  "gpu_cores": 40,
  "memory": {
    "amount": 64,
    "unit": "GB",
    "type": "LPDDR5"
  },
  "serial_number": "XXXXXXXXXX",
  "user": {
    "real_name": "John Doe",
    "login": "jdoe"
  },
  "hostname": "jdoe-laptop",
  "os": {
    "system": "macOS",
    "system_version": "15.2",
    "system_build": "24C101",
    "system_version_code_name": "Sequoia",
    "kernel_type": "Darwin",
    "kernel_version": "24.2.0"
  },
  "system_integrity": "integrity_enabled",
  "disk": {
    "total_tb": 1.9952183,
    "free_tb": 1.1365209,
    "smart_status": "Verified"
  },
  "battery": {
    "status_percent": 93,
    "capacity_percent": 100,
    "health": "Good"
  },
  "displays": [
    {
      "pixels_width": 3456,
      "pixels_height": 2234,
      "resolution_width": 1728,
      "resolution_height": 1117,
      "refresh_rate_hz": 120
    }
  ],
  "software": {
    "num_apps": 65,
    "num_homebrew_formulae": 227,
    "num_homebrew_casks": 37
  },
  "terminal": "iTerm.app",
  "uptime": "1 days, 19 hours",
  "datetime": "Sun, 22 Dec 2024 16:58:35 CET",
  "public_ip": {
    "query": "178.195.102.237",
    "country": "Switzerland"
  }
}
```

## TODO

- Make it work on x86_64
