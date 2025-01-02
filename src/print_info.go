package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/jwalton/go-supportscolor"
)

// helper to create a line of information
func createInfoLine(requestedItem, info string) []string {
	return []string{colorCyan, availableItems[requestedItem].Title, colorNormal, info}
}

// each info line gets a Title and information
// This function calculates the padding size needed to align
// the information of all the lines.
func getPaddingSize(infoLines [][]string) int {
	paddingSize := 0
	for _, i := range infoLines {
		if len(i[1]) > paddingSize {
			paddingSize = len(i[1])
		}
	}
	return paddingSize + 1
}

func padLogoLines(logoLines *[]string) {
	// Find the longest line
	maxLen := 0
	for _, line := range *logoLines {
		if len(line) > maxLen {
			maxLen = len(line)
		}
	}

	// Pad each line
	for i, line := range *logoLines {
		if len(line) < maxLen {
			(*logoLines)[i] = fmt.Sprintf("%-*s", maxLen, line) // Left-align and pad with spaces
		}
	}
}

// Print the information in a human-readable format
func printInfo(hostInfo *info) error {
	var output strings.Builder

	if supportscolor.Stdout().Has256 || supportscolor.Stderr().Has16m {
		//colorRed = "\033[38;5;160m"
		//colorGreen = "\033[38;5;028m"
		//colorYellow = "\033[38;5;220m"
		//colorBlue = "\033[38;5;021m"
		//colorPurple = "\033[38;5;054m"
		colorCyan = "\033[38;5;039m"
		//colorOrange = "\033[38;5;202m"
	} else {
		//colorRed = "\033[00;31m"
		//colorGreen = "\033[00;32m"
		//colorYellow = "\033[00;33m"
		//colorBlue = "\033[00;34m"
		//colorPurple = "\033[00;35m"
		colorCyan = "\033[00;36m"
		//colorOrange = "\033[00;91m"
	}
	colorReset := "\033[0m"

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
		case "uptime":
			infoLines = append(infoLines, createInfoLine(requestedItem, hostInfo.Uptime))
		case "datetime":
			infoLines = append(infoLines, createInfoLine(requestedItem, hostInfo.Datetime))
		case "weather":
			infoLines = append(infoLines, createInfoLine(requestedItem,
				fmt.Sprintf("%s, %s: %s",
					hostInfo.Weather.LocationName,
					hostInfo.Weather.LocationCountryCode,
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
			tmp[1] = "Temp. | Wind"
			infoLines = append(infoLines, tmp)
		}
	}

	/* ---------- Display the information ---------- */
	if *config.DisplayLogo {
		// The logo file consists of lines of text that will be displayed
		// Either each line just contains the text to be displayed, or
		// each line contains two color codes (for 256 colors terminals and 16 colors one)
		// followed by the text to be displayed. --> In that case the fields
		// are separated by a colon.
		var logoLines [][]string
		colorField := 1 // (first field = 256 colors, second field = 16 colors)
		if supportscolor.Stdout().Has256 || supportscolor.Stderr().Has16m {
			colorField = 0
		}

		data, err := os.ReadFile(*config.Logo)
		if err != nil {
			return err
		}
		if len(data) == 0 {
			return fmt.Errorf("Invalid logo (empty)")
		}

		// Let's remove empty lines and comments
		allLines := strings.Split(string(data), "\n")
		lines := make([]string, 0)
		for _, line := range allLines {
			if line == "" || strings.HasPrefix(line, "//") {
				continue
			}
			lines = append(lines, line)
		}
		// Padding each lines with spaces to that each lines is the same length
		padLogoLines(&lines)

		isColoredLogo := false
		if strings.HasPrefix(lines[0], "\\") {
			isColoredLogo = true
		}
		var logoColorLine string
		for _, line := range lines {
			if isColoredLogo {
				fields := strings.SplitN(line, " ", 3)
				// Replace literal \033 with the actual ANSI escape character
				logoColorLine = strings.ReplaceAll(fields[colorField], `\033`, "\033")
				logoLines = append(logoLines, []string{logoColorLine, fields[2]})
			} else {
				logoColorLine = colorReset
				logoLines = append(logoLines, []string{logoColorLine, line})
			}
		}
		lenLogoLine := len(logoLines[0][1])

		/* ---------- Vertically center the logo and the information ---------- */
		// Here, we want to vertically center the display of
		//the logo and the information. So we calculate a padding to be added
		// to the top and bottom of either the logo or the information,
		// depending on which one is shorter.

		lenLogoLines := len(logoLines)
		lenInfoLines := len(infoLines)
		maxLines := max(lenLogoLines, lenInfoLines)

		if lenLogoLines != lenInfoLines {
			minLines := min(lenLogoLines, lenInfoLines)
			topPadding := (maxLines - minLines) / 2
			bottomPadding := maxLines - minLines - topPadding
			prependArr := make([][]string, 0)
			appendArr := make([][]string, 0)

			if lenLogoLines > lenInfoLines {
				emptyLine := []string{"", "", "", ""}
				for i := 0; i < topPadding; i++ {
					prependArr = append(prependArr, emptyLine)
				}
				infoLines = append(prependArr, infoLines...)
				for i := 0; i < bottomPadding; i++ {
					appendArr = append(appendArr, emptyLine)
				}
				infoLines = append(infoLines, appendArr...)
			} else {
				emptyLine := []string{"", strings.Repeat(" ", lenLogoLine)}
				for i := 0; i < topPadding; i++ {
					prependArr = append(prependArr, emptyLine)
				}
				logoLines = append(prependArr, logoLines...)
				for i := 0; i < bottomPadding; i++ {
					appendArr = append(appendArr, emptyLine)
				}
				logoLines = append(logoLines, appendArr...)
			}
		}

		/* ---------- Prepare the logo and the information ---------- */
		dynamicPadding := getPaddingSize(infoLines)
		for i := 0; i < maxLines; i++ {
			output.WriteString(fmt.Sprintf("%s%s  %s%-*s%s%s\n",
				logoLines[i][0],
				logoLines[i][1],
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
