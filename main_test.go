package main

import (
	"fmt"
	"strconv"
	"testing"
)

func TestTaskpaperToSlackHeader(t *testing.T) {
	source := `Monday, January 28: @slack(messageid)`
	target := `:calendar: *Monday, January 28*`

	result := taskpaperToSlackHeader(source)
	if result != target {
		t.Error("Expected: " + target)
		t.Error("Got: " + result)
	}
}

func TestTaskpaperToSlackLine(t *testing.T) {
	source := `	- getmilk`
	target := `:todo: getmilk`
	result := taskpaperToSlackLine(source)
	if result != target {
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + result)
	}

	source = `		- getmilk`
	target = `      :todo: getmilk`
	result = taskpaperToSlackLine(source)
	if result != target {
		t.Error("Expected: " + target)
		t.Error("Got: " + result)
	}

	source = `- getmilk @done`
	target = `:done: getmilk`
	result = taskpaperToSlackLine(source)
	if result != target {
		t.Error("Expected: " + target)
		t.Error("Got: " + result)
	}

	source = `- getmilk @done(2018-21-29) @doing x`
	target = `:done: getmilk`
	result = taskpaperToSlackLine(source)
	if result != target {
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + result)
	}

	source = `- getmilk @doing asd`
	target = `:doing: getmilk`
	result = taskpaperToSlackLine(source)
	if result != target {
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + result)
	}
}

func TestTaskpaperGetSlackMsgId(t *testing.T) {
	_, err := taskpaperGetSlackMsgRef(`Title @slack(asd)`)
	expectedErr := "Slack tag not found"
	if err.Error() != expectedErr {
		t.Error("Expected: " + expectedErr)
		t.Error("Got: " + fmt.Sprintf("%#v", err))
	}

	_, err = taskpaperGetSlackMsgRef(`Title @slack(asd/123123)`)
	expectedErr = "Slack timestamp too short"
	if err.Error() != expectedErr {
		t.Error("Expected: " + expectedErr)
		t.Error("Got: " + fmt.Sprintf("%#v", err))
	}

	line := "Title @slack(B05KSNDD4/p1549566229043400)"
	target := slackMsgRef{channel: "B05KSNDD4", timestamp: "1549566229.043400", urlTimestamp: "p1549566229043400"}
	slackMsgRef, err := taskpaperGetSlackMsgRef(line)
	if err != nil {
		t.Error("Expected: no error")
		t.Error("Got: " + err.Error())
	}
	if slackMsgRef != target {
		t.Error("Expected: \n" + fmt.Sprintf("%#v", target))
		t.Error("Got: \n" + fmt.Sprintf("%#v", slackMsgRef))
	}
}

func TestTaskpaperToSlack(t *testing.T) {
	source := `Today tasks: @slack(B05KSNDD4/p1549566229043400)
	- do something
		nested text
		- nested task`
	targetMsgRef := slackMsgRef{channel: "B05KSNDD4", timestamp: "1549566229.043400", urlTimestamp: "p1549566229043400"}
	targetContent := `:calendar: *Today tasks*
:todo: do something
      :todo: nested task`

	msgRef, msgContent, err := taskpaperToSlack(source)
	if err != nil {
		t.Error("Expected: no error")
		t.Error("Got: " + err.Error())
	}
	if msgRef != targetMsgRef {
		t.Error("Expected: " + fmt.Sprintf("%#v", targetMsgRef))
		t.Error("Got: " + fmt.Sprintf("%#v", msgRef))
	}
	if msgContent != targetContent {
		t.Error("Expected: " + targetContent)
		t.Error("Got: " + msgContent)
	}
}

func TestLineDepth(t *testing.T) {
	result := lineDepth("something 			")
	if result != 0 {
		t.Error("Wrong line depth: " + strconv.Itoa(result))
	}

	result = lineDepth("		- task	")
	if result != 2 {
		t.Error("Expected: " + strconv.Itoa(2))
		t.Error("Got: " + strconv.Itoa(result))
	}
}

func TestIsTodayHeader(t *testing.T) {
	result := isTodayHeader("Header: @something()")
	if result != false {
		t.Error("Not recognized correctly: " + strconv.FormatBool(result))
	}

	result = isTodayHeader("		Header: @slack(asd)")
	if result != false {
		t.Error("Not recognized correctly: " + strconv.FormatBool(result))
	}

	result = isTodayHeader("		Header: @slack(asd/123)")
	if result != true {
		t.Error("Not recognized correctly: " + strconv.FormatBool(result))
	}
}

func TestTaskpaperReduceDepth(t *testing.T) {
	source := `	foo`
	amount := 1
	target := `foo`

	result := taskpaperReduceDepth(source, amount)
	if result != target {
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + result)
	}

	source = `			foo`
	amount = 2
	target = `	foo`

	result = taskpaperReduceDepth(source, amount)
	if result != target {
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + result)
	}
}

func TestTaskpaperFindSlackNode(t *testing.T) {
	content := `
asdasd
	- asdasdasd
Header with no tag:
Greater thing:
	Today tasks: @slack(B05KSNDD4/p154956622904340)
		- do something
			nested text
		- and other stuff
		- nested task
		some text
		Header inside: 
			- task
other shit
	Second tasks: @slack(B05KSNDD4/p154956622922222)
		- do me
`

	target := `Today tasks: @slack(B05KSNDD4/p154956622904340)
	- do something
		nested text
	- and other stuff
	- nested task
	some text
	Header inside: 
		- task`

	result := taskpaperFindSlackNode(content)
	if result != target {
		t.Error("Expected:\n" + target)
		t.Error("Got:\n" + result)
	}
}

func TestGetMessageToSync(t *testing.T) {
	content := `
asdasd
	- asdasdasd
Header with no tag:
Greater thing:
	Today tasks: @slack(B05KSNDD4/p154956622904340)
		- do something
			nested text
other shit
	Second tasks: @slack(messageid)
		- do me
`

	targetMsgRef := slackMsgRef{channel: "B05KSNDD4", timestamp: "1549566229.04340", urlTimestamp: "p154956622904340"}
	targetMsgContent := `:calendar: *Today tasks*
:todo: do something`

	msgRef, msgContent, err := getMessageToSync(content)
	if err != nil {
		t.Error("Expected: no error")
		t.Error("Got: " + err.Error())
	}
	if msgRef != targetMsgRef {
		t.Error("Expected: \n" + fmt.Sprintf("%#v", targetMsgRef))
		t.Error("Got: \n" + fmt.Sprintf("%#v", msgRef))
	}
	if msgContent != targetMsgContent {
		t.Error("Expected:\n" + targetMsgContent)
		t.Error("Got:\n" + msgContent)
	}
}
