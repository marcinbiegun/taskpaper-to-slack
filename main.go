package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// TODO:
// [ ] read token variable
// [ ] read data from taskpaper file
// [ ] send data to Slack (replace message)

func taskpaperToSlackHeader(taskpaper string) (result string) {
	tagAndAfterRe := regexp.MustCompile("@.*")
	taskpaper = tagAndAfterRe.ReplaceAllString(taskpaper, "")
	colonAndAfterRe := regexp.MustCompile(":.*")
	taskpaper = colonAndAfterRe.ReplaceAllString(taskpaper, "")

	return ":calendar: *" + taskpaper + "*"
}

func taskpaperToSlackLine(taskpaper string) (result string) {
	taskpaper = taskpaperReduceDepth(taskpaper, 1)
	taskSymbolRe := regexp.MustCompile("- ")
	taskpaper = taskSymbolRe.ReplaceAllString(taskpaper, ":todo: ")
	tabRe := regexp.MustCompile("\t")
	taskpaper = tabRe.ReplaceAllString(taskpaper, "      ")

	return taskpaper
}

func taskpaperToSlack(taskpaper string) (result string) {
	lines := strings.Split(taskpaper, "\n")
	if len(lines) == 0 {
		return ""
	}

	resultLines := make([]string, 0)

	for index, line := range lines {
		if index == 0 {
			resultLines = append(resultLines, taskpaperToSlackHeader(line))
		} else {
			resultLines = append(resultLines, taskpaperToSlackLine(line))
		}
	}

	return strings.Join(resultLines, "\n")
}

func lineDepth(line string) (result int) {
	prefixTabsRe := regexp.MustCompile(`^\t*`)
	found := prefixTabsRe.FindString(line)
	if found == "" {
		return 0
	}
	return len(found)
}

func isTodayHeader(line string) (result bool) {
	// match, _ := regexp.MatchString(`@slack\(\d{4}-\d{2}-\d{2},[a-zA-Z0-9]+\)`, line)
	match, _ := regexp.MatchString(`@slack\([a-zA-Z0-9]+\)`, line)
	return match
}

func taskpaperReduceDepth(line string, amount int) (result string) {
	re := regexp.MustCompile("^\t{" + strconv.Itoa(amount) + "}")
	return re.ReplaceAllString(line, "")
}

func taskpaperFindSlackNode(content string) (result string) {
	lines := strings.Split(content, "\n")
	readLines := make([]string, 0)
	reading := false
	readingDepth := 0

	for _, line := range lines {
		if reading == true {
			if lineDepth(line) < readingDepth {
				break
			}
		}
		if reading == false && isTodayHeader(line) {
			reading = true
			readingDepth = lineDepth(line)
		}
		if reading == true {
			readLines = append(readLines, line)
		}
	}

	if len(readLines) == 0 {
		return ""
	}

	readLinesCorrectedDepth := make([]string, len(readLines))
	depth := lineDepth(readLines[0])
	for i, line := range readLines {
		readLinesCorrectedDepth[i] = taskpaperReduceDepth(line, depth)
	}
	return strings.Join(readLinesCorrectedDepth, "\n")
}

func getMessageToSync(taskpaper string) (msgid string, result string) {
	taskpaperNode := taskpaperFindSlackNode(taskpaper)
	slackMessage := taskpaperToSlack(taskpaperNode)
	return "asd", slackMessage
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

	msgid, message := getMessageToSync(string(data))

	fmt.Println("Slack Message ID: " + msgid)
	fmt.Println("Slack Message Content:\n" + message)
}
