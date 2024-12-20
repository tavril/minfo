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
	// - Information
	info := [][]string{}

	/* ---------- Create the information lines ---------- */
	for _, requestedItem := range config.Items {
		switch requestedItem {
		case "user":
			info = append(info, createInfoLine("User",
				fmt.Sprintf("%s (%s)", hostInfo.User.RealName, hostInfo.User.Login),
			))
		case "hostname":
			info = append(info, createInfoLine("Hostname", hostInfo.Hostname))
		case "os":
			info = append(info, createInfoLine("OS",
				fmt.Sprintf("%s %s %s (%s) %s %s",
					hostInfo.Os.System,
					hostInfo.Os.SystemVersionCodeNname,
					hostInfo.Os.SystemVersion,
					hostInfo.Os.SystemBuild,
					hostInfo.Os.KernelType,
					hostInfo.Os.KernelVersion,
				),
			))
		case "model":
			info = append(info, createInfoLine("Model",
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
			info = append(info, createInfoLine("CPU", cpuCoreInfo))
		case "gpu":
			info = append(info, createInfoLine("GPU",
				fmt.Sprintf("%d cores", *hostInfo.GpuCores),
			))
		case "memory":
			info = append(info, createInfoLine("Memory",
				fmt.Sprintf("%d %s %s",
					hostInfo.Memory.Amount,
					hostInfo.Memory.Unit,
					hostInfo.Memory.MemType,
				),
			))
		case "disk":
			info = append(info, createInfoLine("Disk",
				fmt.Sprintf("%.2f TB (%.2f TB available)",
					hostInfo.Disk.TotalTB,
					hostInfo.Disk.FreeTB,
				),
			))
			info = append(info, createInfoLine("Disk SMART", hostInfo.Disk.SmartStatus))
		case "battery":
			var charging string
			if hostInfo.Battery.Charging {
				charging = "(charging)"
			} else {
				charging = "(discharging)"
			}
			info = append(info, createInfoLine("Battery",
				fmt.Sprintf("%d%% %s | %d%% capacity",
					hostInfo.Battery.StatusPercent,
					charging,
					hostInfo.Battery.CapacityPercent,
				),
			))
			info = append(info, createInfoLine("Battery health", hostInfo.Battery.Health))
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
			info = append(info, displayLines...)
		case "terminal":
			info = append(info, createInfoLine("Terminal", hostInfo.Terminal))
		case "software":
			info = append(info, createInfoLine("Software",
				fmt.Sprintf("%d Apps | %d Formulae | %d Casks",
					hostInfo.Software.NumApps,
					hostInfo.Software.NumBrewFormulae,
					hostInfo.Software.NumBrewCasks,
				),
			))
		case "public_ip":
			info = append(info, createInfoLine("Public IP", hostInfo.PublicIP))
		case "uptime":
			info = append(info, createInfoLine("Uptime", hostInfo.Uptime))
		case "datetime":
			info = append(info, createInfoLine("Date/Time", hostInfo.Datetime))
		}
	}

	/* ---------- Display the information ---------- */
	if withLogo {
		appleLogo := [][]string{
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
		lenLogoLine := len(appleLogo[0][1])

		/* ---------- Vertically center the logo and the information ---------- */
		// Here, we want to vertically center the display of
		//the logo and the information. So we calculate a padding to be added
		// to the top and bottom of either the logo or the information,
		// depending on which one is shorter.
		lenAppleLogo := len(appleLogo)
		lenInfo := len(info)
		maxLines := max(lenAppleLogo, lenInfo)

		if lenAppleLogo != lenInfo {
			minLines := min(lenAppleLogo, lenInfo)
			topPadding := (maxLines - minLines) / 2
			bottomPadding := maxLines - minLines - topPadding
			prependArr := make([][]string, 0)
			appendArr := make([][]string, 0)

			if lenAppleLogo > lenInfo {
				emptyLine := []string{"", "", "", ""}
				for i := 0; i < topPadding; i++ {
					prependArr = append(prependArr, emptyLine)
				}
				info = append(prependArr, info...)
				for i := 0; i < bottomPadding; i++ {
					appendArr = append(appendArr, emptyLine)
				}
				info = append(info, appendArr...)
			} else {
				emptyLine := []string{"", strings.Repeat(" ", lenLogoLine)}
				for i := 0; i < topPadding; i++ {
					prependArr = append(prependArr, emptyLine)
				}
				appleLogo = append(prependArr, appleLogo...)
				for i := 0; i < bottomPadding; i++ {
					appendArr = append(appendArr, emptyLine)
				}
				appleLogo = append(appleLogo, appendArr...)
			}
		}

		/* ---------- Prepare the logo and the information ---------- */
		for i := 0; i < maxLines; i++ {
			output.WriteString(fmt.Sprintf("%s%s%s%-15s%s%s\n",
				appleLogo[i][0],
				appleLogo[i][1],
				info[i][0],
				info[i][1],
				info[i][2],
				info[i][3],
			))
		}
	} else {
		/* ---------- Prepare only the information ---------- */
		for _, i := range info {
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
