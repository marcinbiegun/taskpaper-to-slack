package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
)

// TODO:
// [ ] read token variable
// [ ] read data from taskpaper file
// [ ] send data to Slack (replace message)

func lineDepth(line string) (result int) {
	rp := regexp.MustCompile(`\t*`)
	tabsString := rp.FindString(line)
	if tabsString == "" {
		return 0
	} else {
		foundTabStrings := rp.FindAllString(tabsString, -1)
		return len(foundTabStrings)
	}
}

func isTodayHeader(line string) (result bool) {
	match, _ := regexp.MatchString(`@slack\(\d{4}-\d{2}-\d{2},[a-zA-Z0-9]+\)`, line)
	return match
}

func extractToday(content string) (result string) {
	lines := strings.Split(content, "\n")
	todayLines := make([]string, 1000)

	startedHeader := false
	startedDepth := 0
	if true {
		startedDepth = startedDepth + 0
	}

	for index, line := range lines {
		// fmt.Println("Line " + strconv.Itoa(index) + ":")
		// fmt.Println(line)
		if isTodayHeader(line) {
			startedHeader = true
			startedDepth = index
		}
		if startedHeader {
			todayLines = append(todayLines, line)
		}
	}

	return content
}

func main() {
	fmt.Println("TOKEN:", os.Getenv("TOKEN"))

	if len(os.Args) < 2 {
		fmt.Println("ERROR: the first parameter should be taskpaper filename")
		return
	}
	filePath := os.Args[1]

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}

	todayTaskpaper := extractToday(string(data))

	fmt.Println("found today block:")
	fmt.Println(todayTaskpaper)
}
