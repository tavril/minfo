package main

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
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

// Print the information in a human-readable format
func printInfo(hostInfo *info, withLogo bool) {
	var output strings.Builder

	if strings.Contains(os.Getenv("TERM"), "256") {
		colorRed = "\033[38;5;160m"
		colorGreen = "\033[38;5;028m"
		colorYellow = "\033[38;5;220m"
		colorBlue = "\033[38;5;021m"
		colorPurple = "\033[38;5;054m"
		colorCyan = "\033[38;5;039m"
		colorOrange = "\033[38;5;202m"
	} else {
		colorRed = "\033[00;31m"
		colorGreen = "\033[00;32m"
		colorYellow = "\033[00;33m"
		colorBlue = "\033[00;34m"
		colorPurple = "\033[00;35m"
		colorCyan = "\033[00;36m"
		colorOrange = "\033[00;91m"
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
		}
	}

	/* ---------- Display the information ---------- */
	if withLogo {

		file, err := os.Open(fmt.Sprintf("%s/git/minfo/logos/apple.yaml", os.Getenv("HOME")))
		if err != nil {
			//return fmt.Errorf("failed to open config file: %w", err)
			return
		}
		defer file.Close()

		// Parse the YAML file into the Config structure
		decoder := yaml.NewDecoder(file)

		var logo logo
		var logoLines [][]string
		if err := decoder.Decode(&logo); err != nil {
			//return err
			return
		}
		var is256Color bool
		if strings.Contains(os.Getenv("TERM"), "256") {
			is256Color = true
		}
		for _, line := range logo.Lines {
			if is256Color {
				logoLines = append(logoLines, []string{line.Color256, line.Text})
			} else {
				logoLines = append(logoLines, []string{line.Color16, line.Text})
			}
		}
		lenLogoLine := len(logoLines[0][1])

		/*
			defaultLogoLines := [][]string{
				{colorGreen, "                    ##           "},
				{colorGreen, "                  ####           "},
				{colorGreen, "                #####            "},
				{colorGreen, "               ####              "},
				{colorGreen, "      ########   ############    "},
				{colorGreen, "    ##########################   "},
				{colorYellow, "  ###########################    "},
				{colorYellow, "  ##########################     "},
				{colorOrange, " ##########################      "},
				{colorOrange, " ##########################      "},
				{colorRed, " ###########################     "},
				{colorRed, "  ############################   "},
				{colorPurple, "  #############################  "},
				{colorPurple, "   ############################  "},
				{colorBlue, "     ########################    "},
				{colorBlue, "      ######################     "},
				{colorBlue, "        #######    #######       "},
			}
		*/

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
			output.WriteString(fmt.Sprintf("%s%s%s%-*s%s%s\n",
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
}
