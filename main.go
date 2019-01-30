package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/nlopes/slack"
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

	// Find done tag
	doneTagRe := regexp.MustCompile(`@done`)
	foundDoneTag := doneTagRe.FindString(taskpaper)

	// Find doing
	doingTagRe := regexp.MustCompile(`@doing`)
	foundDoingTag := doingTagRe.FindString(taskpaper)

	// Remove tags
	tagAndAfterRe := regexp.MustCompile(`@.*`)
	taskpaper = tagAndAfterRe.ReplaceAllString(taskpaper, "")

	// Remove trailing whitespace
	trailingWhitespaceRe := regexp.MustCompile(`\s*$`)
	taskpaper = trailingWhitespaceRe.ReplaceAllString(taskpaper, "")

	// Replace task symbol with emoji
	taskSymbolRe := regexp.MustCompile("- ")
	if foundDoneTag != "" {
		taskpaper = taskSymbolRe.ReplaceAllString(taskpaper, ":done: ")
	} else if foundDoingTag != "" {
		taskpaper = taskSymbolRe.ReplaceAllString(taskpaper, ":doing: ")
	} else {
		taskpaper = taskSymbolRe.ReplaceAllString(taskpaper, ":todo: ")
	}

	// Replace tab indentations as spaces
	tabRe := regexp.MustCompile("\t")
	taskpaper = tabRe.ReplaceAllString(taskpaper, "     ")

	return taskpaper
}

func taskpaperGetSlackMsgID(line string) (msgID string) {
	slackTagValueRe := regexp.MustCompile(`@slack\(([a-zA-Z0-9]*)\)`)
	matches := slackTagValueRe.FindStringSubmatch(line)
	if len(matches) >= 2 {
		return matches[1]
	}
	return ""
}

func taskpaperToSlack(taskpaper string) (msgID string, msgContent string) {
	lines := strings.Split(taskpaper, "\n")
	if len(lines) == 0 {
		return "", ""
	}

	resultLines := make([]string, 0)
	msgID = ""

	for index, line := range lines {
		if index == 0 {
			msgID = taskpaperGetSlackMsgID(line)
			resultLines = append(resultLines, taskpaperToSlackHeader(line))
		} else {
			resultLines = append(resultLines, taskpaperToSlackLine(line))
		}
	}

	return msgID, strings.Join(resultLines, "\n")
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
	messageID, messageContent := taskpaperToSlack(taskpaperNode)
	return messageID, messageContent
}

func slackMessageLinkIDtoTimestamp(id string) (timestamp string) {
	firstPRe := regexp.MustCompile("^p")
	idWithoutP := firstPRe.ReplaceAllString(id, "")
	index := 10
	idWithDot := idWithoutP[:index] + "." + idWithoutP[index:]
	return idWithDot
}

func apiSlackUpdateMessage(slackToken string, channelID string, messageID string, content string) (string, error) {
	msgTimestamp := slackMessageLinkIDtoTimestamp(messageID)
	api := slack.New(slackToken)
	_, _, _, err := api.UpdateMessage(channelID, msgTimestamp, slack.MsgOptionText(content, false))
	if err != nil {
		return messageID, err
	}
	return messageID, nil
}

func watchFileUpdateSlackOnChange(filePath string, slackSubdomain string, slackToken string, slackChannelID string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	fmt.Println("Watching file:", filePath)

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					//log.Println("modified file:", event.Name)
					slackMessageID, err := updateSlackMessageFromFile(filePath, slackToken, slackChannelID)
					if err != nil {
						fmt.Println("Slack update error:", err)
					}
					fmt.Printf("Updated: https://%s.slack.com/archives/%s/%s\n", slackSubdomain, slackChannelID, slackMessageID)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				fmt.Println("File watch error:", err)
			}
		}
	}()

	err = watcher.Add(filePath)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func updateSlackMessageFromFile(filePath string, slackToken string, slackChannelID string) (msgID string, err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	msgID, msgContent := getMessageToSync(string(data))
	_, err = apiSlackUpdateMessage(slackToken, slackChannelID, msgID, msgContent)
	if err != nil {
		return msgID, err
	}
	return msgID, nil
}

func main() {
	slackToken := os.Getenv("SLACK_TOKEN")
	slackChannelID := os.Getenv("SLACK_CHANNEL_ID")
	slackSubdomain := os.Getenv("SLACK_SUBDOMAIN")
	fmt.Println("SLACK_TOKEN: " + slackToken)
	fmt.Println("SLACK_CHANNEL_ID: " + slackChannelID)
	fmt.Println("SLACK_SUBDOMAIN: " + slackSubdomain)

	if len(os.Args) < 2 {
		fmt.Println("ERROR: the first parameter should be taskpaper filename")
		return
	}
	filePath := os.Args[1]

	watchFileUpdateSlackOnChange(filePath, slackSubdomain, slackToken, slackChannelID)
}
