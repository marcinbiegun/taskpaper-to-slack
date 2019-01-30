package main

import (
	"strconv"
	"testing"
)

func TestTaskpaperToSlackHeader(t *testing.T) {
	source := `Monday, January 28: @slack(messageid)`
	target := `:calendar: *Monday, January 28*`

	result := taskpaperToSlackHeader(source)
	if result != target {
		t.Error("Bad transition to slack header format")
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + result)
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
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + result)
	}

	source = `- getmilk @done`
	target = `:done: getmilk`
	result = taskpaperToSlackLine(source)
	if result != target {
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + result)
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
	source := `Title @slack(asd123) @slack(qwe456)`
	target := `asd123`
	msgID := taskpaperGetSlackMsgID(source)
	if msgID != target {
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + msgID)
	}

	source = `Title @slacker(asd123)`
	target = ``
	msgID = taskpaperGetSlackMsgID(source)
	if msgID != target {
		t.Error("Expected: \n" + target)
		t.Error("Got: \n" + msgID)
	}
}

func TestTaskpaperToSlack(t *testing.T) {
	source := `Today tasks: @slack(qwe)
	- do something
		nested text`
	targetID := `qwe`
	targetContent := `:calendar: *Today tasks*
:todo: do something
      nested text`

	msgID, msgContent := taskpaperToSlack(source)
	if msgID != targetID {
		t.Error("Bad message ID")
		t.Error("Expected: \n" + targetID)
		t.Error("Got: \n" + msgID)
	}

	if msgContent != targetContent {
		t.Error("Bad message content")
		t.Error("Expected: \n" + targetContent)
		t.Error("Got: \n" + msgContent)
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

	result = isTodayHeader("		Header: @slack(123,asd)")
	if result != false {
		t.Error("Not recognized correctly: " + strconv.FormatBool(result))
	}

	result = isTodayHeader("		Header: @slack(ab1c50e)")
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
	Today tasks: @slack(messageid)
		- do something
			nested text
		- and other stuff
		- nested task
		some text
		Header inside: 
			- task
other shit
	Second tasks: @slack(messageid)
		- do me
`

	target := `Today tasks: @slack(messageid)
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

func TestSlackMessageLinkIDtoTimestamp(t *testing.T) {
	source := "p1548757857024300"
	target := "1548757857.024300"
	result := slackMessageLinkIDtoTimestamp(source)

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
	Today tasks: @slack(asd123)
		- do something
			nested text
other shit
	Second tasks: @slack(messageid)
		- do me
`

	targetMsgID := `asd123`
	targetMessage := `:calendar: *Today tasks*
:todo: do something
      nested text`

	msgid, message := getMessageToSync(content)
	if msgid != targetMsgID {
		t.Error("Expected:\n" + targetMsgID)
		t.Error("Got:\n" + msgid)
	}
	if message != targetMessage {
		t.Error("Expected:\n" + targetMessage)
		t.Error("Got:\n" + message)
	}
}
