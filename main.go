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

type slackMsgRef struct {
	channel      string
	timestamp    string
	urlTimestamp string
}

func taskpaperToSlackHeader(taskpaper string) (result string) {
	tagAndAfterRe := regexp.MustCompile("@.*")
	taskpaper = tagAndAfterRe.ReplaceAllString(taskpaper, "")
	colonAndAfterRe := regexp.MustCompile(":.*")
	taskpaper = colonAndAfterRe.ReplaceAllString(taskpaper, "")

	return ":calendar: *" + taskpaper + "*"
}

func taskpaperToSlackLine(taskpaper string) (result string) {
	taskpaper = taskpaperReduceDepth(taskpaper, 1)

	startsWithTaskSymbol, _ := regexp.MatchString(`^\t*- `, taskpaper)
	if startsWithTaskSymbol == false {
		return ""
	}

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
	taskpaper = tabRe.ReplaceAllString(taskpaper, "      ")

	return taskpaper
}

func taskpaperGetSlackMsgRef(line string) (msgRef slackMsgRef, err error) {
	slackTagValueRe := regexp.MustCompile(`@slack\(([a-zA-Z0-9]*)/([a-zA-Z0-9]*)\)`)
	matches := slackTagValueRe.FindStringSubmatch(line)
	if len(matches) >= 3 {
		channel := matches[1]
		urlTimestamp := matches[2]
		if len([]rune(urlTimestamp)) < 16 {
			return msgRef, fmt.Errorf("Slack timestamp too short")
		}
		urlTimestmapWithoutP := regexp.MustCompile("^p").ReplaceAllString(urlTimestamp, "")
		apiTimestamp := urlTimestmapWithoutP[:10] + "." + urlTimestmapWithoutP[10:]
		return slackMsgRef{channel: channel, timestamp: apiTimestamp, urlTimestamp: urlTimestamp}, nil
	}
	return msgRef, fmt.Errorf("Slack tag not found")
}

func taskpaperToSlack(taskpaper string) (msgRef slackMsgRef, msgContent string, err error) {
	resultLines := make([]string, 0)
	lines := strings.Split(taskpaper, "\n")

	for index, line := range lines {
		if index == 0 {
			msgRef, err = taskpaperGetSlackMsgRef(line)
			if err != nil {
				return msgRef, "", err
			}
			resultLines = append(resultLines, taskpaperToSlackHeader(line))
		} else {
			newResultLine := taskpaperToSlackLine(line)
			if newResultLine != "" {
				resultLines = append(resultLines, newResultLine)
			}
		}
	}

	if len(resultLines) == 0 {
		return msgRef, "", fmt.Errorf("Nothing to sync")
	}

	return msgRef, strings.Join(resultLines, "\n"), nil
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
	match, _ := regexp.MatchString(`@slack\([a-zA-Z0-9]+/[a-zA-Z0-9]+\)`, line)
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

func getMessageToSync(taskpaper string) (msgRef slackMsgRef, result string, err error) {
	taskpaperNode := taskpaperFindSlackNode(taskpaper)
	msgRef, messageContent, err := taskpaperToSlack(taskpaperNode)
	return msgRef, messageContent, err
}

func apiSlackUpdateMessage(slackToken string, msgRef slackMsgRef, content string) (err error) {
	api := slack.New(slackToken)
	_, _, _, err = api.UpdateMessage(msgRef.channel, msgRef.timestamp, slack.MsgOptionText(content, false))
	if err != nil {
		return err
	}
	return nil
}

func watchFileUpdateSlackOnChange(filePath string, slackToken string, slackSubdomain string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
					slackMsgRef, err := updateSlackMessageFromFile(filePath, slackToken)
					if err != nil {
						fmt.Println("Slack update error:", err)
					}
					fmt.Printf("Updated message https://%s.slack.com/archives/%s/%s\n", slackSubdomain, slackMsgRef.channel, slackMsgRef.urlTimestamp)
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

func updateSlackMessageFromFile(filePath string, slackToken string) (msgRef slackMsgRef, err error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return msgRef, err
	}
	msgRef, msgContent, err := getMessageToSync(string(data))
	if err != nil {
		return msgRef, err
	}

	err = apiSlackUpdateMessage(slackToken, msgRef, msgContent)
	if err != nil {
		return msgRef, err
	}
	return msgRef, nil
}

func main() {
	slackToken := os.Getenv("SLACK_TOKEN")
	slackSubdomain := os.Getenv("SLACK_SUBDOMAIN")

	if len(slackToken) < 2 {
		fmt.Println("ERROR: no SLACK_TOKEN env var found")
		return
	}

	if len(slackToken) < 2 {
		fmt.Println("ERROR: no SLACK_SUBDOMAIN env var found")
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("ERROR: not filename found in first parameter")
		return
	}
	filePath := os.Args[1]

	watchFileUpdateSlackOnChange(filePath, slackToken, slackSubdomain)
	fmt.Println("Watching file: ", filePath)
}
