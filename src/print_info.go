package main

import (
	"fmt"
	"os"
	"strings"
)

// helper to create a line of information
func createInfoLine(title, info string) []string {
	return []string{colorCyan, title, colorNormal, info}
}

// Print the information in a human-readable format
func printInfo(hostInfo *info, withLogo bool) {
	var output strings.Builder

	if strings.Contains(os.Getenv("TERM"), "256") {
		colorRed = "\033[38;5;160m"
		colorGreen = "\033[38;5;028m"
		colorYellow = "\033[38;5;226m"
		colorBlue = "\033[38;5;021m"
		colorPurple = "\033[38;5;054m"
		colorCyan = "\033[38;5;075m"
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
	// Each item of ìnfo` is a slice of strings which contains
	// - Color code for the title
	// - Title
	// - Color code for the information
	// - infomation
	infoLines := [][]string{}

	/* ---------- Create the information lines ---------- */
	for _, requestedItem := range config.Items {
		switch requestedItem {
		case "user":
			infoLines = append(infoLines, createInfoLine("User",
				fmt.Sprintf("%s (%s)", hostInfo.User.RealName, hostInfo.User.Login),
			))
		case "hostname":
			infoLines = append(infoLines, createInfoLine("Hostname", hostInfo.Hostname))
		case "os":
			infoLines = append(infoLines, createInfoLine("OS",
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
			infoLines = append(infoLines, createInfoLine("macOS SIP", hostInfo.SystemIntegrity))
		case "serial_number":
			infoLines = append(infoLines, createInfoLine("Serial Number", *hostInfo.SerialNumber))
		case "model":
			infoLines = append(infoLines, createInfoLine("Model",
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
			infoLines = append(infoLines, createInfoLine("CPU", cpuCoreInfo))
		case "gpu":
			infoLines = append(infoLines, createInfoLine("GPU",
				fmt.Sprintf("%d cores", *hostInfo.GpuCores),
			))
		case "memory":
			infoLines = append(infoLines, createInfoLine("Memory",
				fmt.Sprintf("%d %s %s",
					hostInfo.Memory.Amount,
					hostInfo.Memory.Unit,
					hostInfo.Memory.MemType,
				),
			))
		case "disk":
			infoLines = append(infoLines, createInfoLine("Disk",
				fmt.Sprintf("%.2f TB (%.2f TB available)",
					hostInfo.Disk.TotalTB,
					hostInfo.Disk.FreeTB,
				),
			))
			infoLines = append(infoLines, createInfoLine("Disk SMART", hostInfo.Disk.SmartStatus))
		case "battery":
			var charging string
			if hostInfo.Battery.Charging {
				charging = "(charging)"
			} else {
				charging = "(discharging)"
			}
			infoLines = append(infoLines, createInfoLine("Battery",
				fmt.Sprintf("%d%% %s | %d%% capacity",
					hostInfo.Battery.StatusPercent,
					charging,
					hostInfo.Battery.CapacityPercent,
				),
			))
			infoLines = append(infoLines, createInfoLine("Battery health", hostInfo.Battery.Health))
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
				tmpStr := fmt.Sprintf("Display #%d", i+1)
				displayLines[i] = createInfoLine(tmpStr, d)
			}
			infoLines = append(infoLines, displayLines...)
		case "terminal":
			infoLines = append(infoLines, createInfoLine("Terminal", hostInfo.Terminal))
		case "software":
			infoLines = append(infoLines, createInfoLine("Software",
				fmt.Sprintf("%d Apps | %d Formulae | %d Casks",
					hostInfo.Software.NumApps,
					hostInfo.Software.NumBrewFormulae,
					hostInfo.Software.NumBrewCasks,
				),
			))
		case "public_ip":
			// Case we have a "Unknown" country (any error in function getPublicIpInfo)
			if len(hostInfo.PublicIp.Country) == 0 {
				infoLines = append(infoLines, createInfoLine("Public IP", hostInfo.PublicIp.IP))
			} else {
				infoLines = append(infoLines, createInfoLine("Public IP",
					fmt.Sprintf("%s (%s)",
						hostInfo.PublicIp.IP,
						hostInfo.PublicIp.Country,
					),
				))
			}
		case "uptime":
			infoLines = append(infoLines, createInfoLine("Uptime", hostInfo.Uptime))
		case "datetime":
			infoLines = append(infoLines, createInfoLine("Date/Time", hostInfo.Datetime))
		}
	}

	/* ---------- Display the information ---------- */
	if withLogo {
		appleLogoLines := [][]string{
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
		lenLogoLine := len(appleLogoLines[0][1])

		/* ---------- Vertically center the logo and the information ---------- */
		// Here, we want to vertically center the display of
		//the logo and the information. So we calculate a padding to be added
		// to the top and bottom of either the logo or the information,
		// depending on which one is shorter.
		lenAppleLogoLines := len(appleLogoLines)
		lenInfoLines := len(infoLines)
		maxLines := max(lenAppleLogoLines, lenInfoLines)

		if lenAppleLogoLines != lenInfoLines {
			minLines := min(lenAppleLogoLines, lenInfoLines)
			topPadding := (maxLines - minLines) / 2
			bottomPadding := maxLines - minLines - topPadding
			prependArr := make([][]string, 0)
			appendArr := make([][]string, 0)

			if lenAppleLogoLines > lenInfoLines {
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
				appleLogoLines = append(prependArr, appleLogoLines...)
				for i := 0; i < bottomPadding; i++ {
					appendArr = append(appendArr, emptyLine)
				}
				appleLogoLines = append(appleLogoLines, appendArr...)
			}
		}

		/* ---------- Prepare the logo and the information ---------- */
		for i := 0; i < maxLines; i++ {
			output.WriteString(fmt.Sprintf("%s%s%s%-15s%s%s\n",
				appleLogoLines[i][0],
				appleLogoLines[i][1],
				infoLines[i][0],
				infoLines[i][1],
				infoLines[i][2],
				infoLines[i][3],
			))
		}
	} else {
		/* ---------- Prepare only the information ---------- */
		for _, i := range infoLines {
			output.WriteString(fmt.Sprintf("%s%-15s%s%s\n",
				i[0],
				i[1],
				i[2],
				i[3],
			))
		}
	}

	fmt.Printf("%s", output.String())
}
