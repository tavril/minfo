package main

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/jwalton/go-supportscolor"
)

// helper to create a line of information
func createInfoLine(requestedItem, info string) []string {
	var realTitle string
	if config.DisplayNerdSymbols != nil && *config.DisplayNerdSymbols {
		realTitle = fmt.Sprintf("%s %s", availableItems[requestedItem].Nerd, availableItems[requestedItem].Title)
	} else {
		realTitle = availableItems[requestedItem].Title
	}
	//return []string{colorCyan, availableItems[requestedItem].Title, colorNormal, info}
	return []string{colorCyan, realTitle, colorNormal, info}
}

// Each info line gets a Title and actual information
// This function calculates the padding size needed to align
// the information of all the lines (i.e. calulate the longest title).
func getPaddingSize(infoLines [][]string) int {
	paddingSize := 0
	for _, i := range infoLines {
		if len(i[1]) > paddingSize {
			paddingSize = len(i[1])
		}
	}
	return paddingSize + 1
}

// This function ensures the logo is padding (i.e. suffixed) with spaces,
// so that all the lines of the logo have the same lenght.
// We have to deal with ANSI codes .... (not taken into account in line lenght)
// It returns the length of the longest line of the logo.
func padLogoLines(logoLines *[]string) int {
	// Find the longest line
	maxLen := 0
	for _, line := range *logoLines {
		lenLine := utf8.RuneCountInString(reANSI.ReplaceAllString(line, ""))
		if lenLine > maxLen {
			maxLen = lenLine
		}
	}

	for i, line := range *logoLines {
		lenLine := utf8.RuneCountInString(reANSI.ReplaceAllString(line, ""))
		if lenLine < maxLen {
			(*logoLines)[i] = fmt.Sprintf("%s%-*s", line, maxLen-lenLine, " ")
		}
	}
	return maxLen
}

// Print the information in a human-readable format
func printInfo(hostInfo *info) error {
	var output strings.Builder

	if supportscolor.Stdout().Has256 || supportscolor.Stderr().Has16m {
		colorCyan = "\u001B[38;5;039m"
	} else {
		colorCyan = "\u001B[00;36m"
	}

	// Each item of infoLines is a slice of strings representing a line
	// of information. Each line contains:
	// - Color code for the Item title
	// - Title
	// - Color code for the information
	// - infomation
	infoLines := [][]string{}

	/* ---------- Create the information lines ---------- */
	for _, requestedItem := range config.Items {
		switch requestedItem {
		case "user":
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%s (%s)", hostInfo.User.RealName, hostInfo.User.Login),
			))
		case "hostname":
			infoLines = append(infoLines, createInfoLine(requestedItem, hostInfo.Hostname))
		case "os":
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%s %s %s (%s) %s %s",
					hostInfo.Os.System,
					hostInfo.Os.SystemVersionCodeNname,
					hostInfo.Os.SystemVersion,
					hostInfo.Os.SystemBuild,
					hostInfo.Os.KernelType,
					hostInfo.Os.KernelVersion,
				),
			))
		case "system_integrity":
			hostInfo.SystemIntegrity = capitalizeFirstLetter(
				strings.TrimPrefix(hostInfo.SystemIntegrity, "integrity_"),
			)
			infoLines = append(infoLines, createInfoLine(requestedItem, hostInfo.SystemIntegrity))
		case "serial_number":
			infoLines = append(infoLines, createInfoLine(requestedItem, *hostInfo.SerialNumber))
		case "model":
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%s %s (%s) %s",
					hostInfo.Model.Name,
					hostInfo.Model.SubName,
					hostInfo.Model.Date,
					hostInfo.Model.Number,
				),
			))
		case "cpu":
			var cpuCoreInfo string
			if strings.HasPrefix(hostInfo.Cpu.Model, "Apple") {
				cpuCoreInfo = fmt.Sprintf("%s %d cores (%d P and %d E)",
					hostInfo.Cpu.Model,
					hostInfo.Cpu.Cores,
					hostInfo.Cpu.PerformanceCores,
					hostInfo.Cpu.EfficiencyCores,
				)
			} else {
				// Intel CPU: The model also includes the number of cores
				// (Ex: "6-Core Intel Core i7")
				cpuCoreInfo = fmt.Sprintf("%s", hostInfo.Cpu.Model)
			}
			infoLines = append(infoLines, createInfoLine(requestedItem, cpuCoreInfo))
		case "gpu":
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%d cores", *hostInfo.GpuCores),
			))
		case "memory":
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%d %s %s",
					hostInfo.Memory.Amount,
					hostInfo.Memory.Unit,
					hostInfo.Memory.MemType,
				),
			))
		case "disk":
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%.2f TB (%.2f TB available)",
					hostInfo.Disk.TotalTB,
					hostInfo.Disk.FreeTB,
				),
			))
			tmp := createInfoLine(requestedItem, hostInfo.Disk.SmartStatus)
			tmp[1] = fmt.Sprintf("%s SMART", tmp[1])
			infoLines = append(infoLines, tmp)
		case "battery":
			var charging string
			if hostInfo.Battery.Charging {
				charging = "(charging)"
			} else {
				charging = "(discharging)"
			}
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%d%% %s | %d%% capacity",
					hostInfo.Battery.StatusPercent,
					charging,
					hostInfo.Battery.CapacityPercent,
				),
			))
			tmp := createInfoLine(requestedItem, hostInfo.Battery.Health)
			tmp[1] = fmt.Sprintf("%s health", tmp[1])
			infoLines = append(infoLines, tmp)
		case "display":
			var displayInfo []string
			for _, display := range hostInfo.Displays {
				displayInfo = append(displayInfo, fmt.Sprintf("%d x %d | %d x %d @ %.0f Hz",
					display.PixelsWidth,
					display.PixelsHeight,
					display.ResolutionWidth,
					display.ResolutionHeight,
					display.RefreshRateHz,
				))
			}
			displayLines := make([][]string, len(hostInfo.Displays))
			for i, d := range displayInfo {
				displayLines[i] = createInfoLine(requestedItem, d)
				displayLines[i][1] = fmt.Sprintf("%s #%d", displayLines[i][1], i+1)
			}
			infoLines = append(infoLines, displayLines...)
		case "terminal":
			infoLines = append(infoLines, createInfoLine(requestedItem, hostInfo.Terminal))
		case "software":
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%d Apps | %d Formulae | %d Casks",
					hostInfo.Software.NumApps,
					hostInfo.Software.NumBrewFormulae,
					hostInfo.Software.NumBrewCasks,
				),
			))
		case "public_ip":
			if hostInfo.PublicIp != nil {
				// Case we have a "Unknown" country (any error in function getPublicIpInfo)
				if len(hostInfo.PublicIp.Country) == 0 {
					infoLines = append(infoLines, createInfoLine(requestedItem, hostInfo.PublicIp.IP))
				} else {
					infoLines = append(infoLines, createInfoLine(requestedItem,
						fmt.Sprintf("%s (%s)",
							hostInfo.PublicIp.IP,
							hostInfo.PublicIp.Country,
						),
					))
				}
			} else {
				infoLines = append(infoLines, createInfoLine(requestedItem, "Unknown"))
			}
		case "uptime":
			infoLines = append(infoLines, createInfoLine(requestedItem, hostInfo.Uptime))
		case "datetime":
			infoLines = append(infoLines, createInfoLine(requestedItem, hostInfo.Datetime))
		case "weather":
			var location string

			if hostInfo.Weather.LocationName != "" {
				location = fmt.Sprintf("%s, %s", hostInfo.Weather.LocationName, hostInfo.Weather.LocationCountryCode)
			} else if hostInfo.PublicIp != nil {
				location = fmt.Sprintf("%s, %s", hostInfo.PublicIp.City, hostInfo.PublicIp.CountryCode)
			} else {
				location = fmt.Sprintf("(%f, %f)", hostInfo.Weather.Latitude, hostInfo.Weather.Longitude)
			}
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%s: %s",
					location,
					hostInfo.Weather.CurrentWeather,
				),
			))
			tmp := createInfoLine(requestedItem,
				fmt.Sprintf(
					"%s (%s) %s | %s %.0f (%.0f) %s",
					formatFloat(roundToNearestHalf(hostInfo.Weather.Temperature)),
					formatFloat(roundToNearestHalf(hostInfo.Weather.FeelsLike)),
					hostInfo.Weather.TempUnit,
					windArrow(hostInfo.Weather.WindDirection),
					hostInfo.Weather.WindSpeed,
					hostInfo.Weather.WindGusts,
					hostInfo.Weather.WindUnit,
				),
			)
			if config.DisplayNerdSymbols != nil && *config.DisplayNerdSymbols {
				tmp[1] = " Temp. | Wind"
			} else {
				tmp[1] = "Temp. | Wind"
			}
			infoLines = append(infoLines, tmp)
		}
	}

	/* ---------- Display the information ---------- */
	if *config.DisplayLogo {
		var logoLines []string

		data, err := os.ReadFile(*config.Logo)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return fmt.Errorf("Invalid logo (empty)")
		}

		// Let's remove empty lines and comments
		allLines := strings.Split(string(data), "\n")
		for _, line := range allLines {
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			logoLines = append(logoLines, line)
		}
		// Padding each lines with spaces to that each lines is the same length
		lenLogoLine := padLogoLines(&logoLines)

		/* ---------- Vertically center the logo and the information ---------- */
		// Here, we want to vertically center the display of
		// the logo and the information. So we calculate a padding to be added
		// to the top and bottom of either the logo or the information,
		// depending on which one is shorter.

		lenLogoLines := len(logoLines)
		lenInfoLines := len(infoLines)
		maxLines := max(lenLogoLines, lenInfoLines)

		if lenLogoLines != lenInfoLines {
			minLines := min(lenLogoLines, lenInfoLines)
			topPadding := (maxLines - minLines) / 2
			bottomPadding := maxLines - minLines - topPadding
			prependInfoArr := make([][]string, 0)
			appendInfoArr := make([][]string, 0)
			prependLogoArr := make([]string, 0)
			appendLogoArr := make([]string, 0)

			if lenLogoLines > lenInfoLines {
				emptyLine := []string{"", "", "", ""}
				for range topPadding {
					prependInfoArr = append(prependInfoArr, emptyLine)
				}
				infoLines = append(prependInfoArr, infoLines...)
				for range bottomPadding {
					appendInfoArr = append(appendInfoArr, emptyLine)
				}
				infoLines = append(infoLines, appendInfoArr...)
			} else {
				emptyLine := strings.Repeat(" ", lenLogoLine)
				for range topPadding {
					prependLogoArr = append(prependLogoArr, emptyLine)
				}
				logoLines = append(prependLogoArr, logoLines...)
				for range bottomPadding {
					appendLogoArr = append(appendLogoArr, emptyLine)
				}
				logoLines = append(logoLines, appendLogoArr...)
			}
		}

		/* ---------- Prepare the logo and the information ---------- */
		dynamicPadding := getPaddingSize(infoLines)
		for i := range maxLines {
			output.WriteString(fmt.Sprintf("%s  %s%-*s%s%s\n",
				logoLines[i],
				infoLines[i][0],
				dynamicPadding,
				infoLines[i][1],
				infoLines[i][2],
				infoLines[i][3],
			))
		}
	} else {
		/* ---------- Prepare only the information ---------- */
		dynamicPadding := getPaddingSize(infoLines)
		for _, i := range infoLines {
			output.WriteString(fmt.Sprintf("%s%-*s%s%s\n",
				i[0],
				dynamicPadding,
				i[1],
				i[2],
				i[3],
			))
		}
	}

	fmt.Printf("%s", output.String())
	return nil
}
