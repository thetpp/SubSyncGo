package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

type Subtitle struct {
	Number         int
	StartHour      int
	StartMin       int
	StartSec       int
	StartMili      int
	EndHour        int
	EndMin         int
	EndSec         int
	EndMili        int
	SubtitleString string
}

func ChangeTime(h, m, s, mili, ha, ma, sa, miliDiff int) (int, int, int, int) {
	newH := h + ha
	newM := m + ma
	newS := s + sa

	stMili := mili + miliDiff

	if stMili > 999 {
		stMili -= 999
		newS++
	} else if stMili < 0 {
		stMili = 999 + stMili
		newS--
	}

	if newM > 59 {
		newH += 1
		newM = newM - 60
	}

	if newS > 59 {
		newM += 1
		newS = newS - 60
	}

	if newH < 0 {
		newH = 0
	}

	if newM < 0 {
		newM = 60 + newM
		newH -= -1
	}

	if newS < 0 {
		newS = 60 + newS
		newM -= 1
	}

	return newH, newM, newS, stMili
}

func inti(s string) int {
	sInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		fmt.Println(err)
	}
	return int(sInt)
}

func main() {
	helpMessage := `
Use the following argumets:
"f=": To specify the .srt file.[path_to_file/filename.srt]
"num=": To specify the subtitle number.[2]
"start=": To specify the starting time for the subtitle.[00:02:33,344]

[All the above argumets must be specified.]

Use the following tags:
"help": To get this help message.
"readsrt": Prints the first 100 lines of the SRT file.
				`

	filePath := ""
	timeToSet := ""
	numStr := ""
	readFile := false

	for _, arg := range os.Args {
		if strings.Contains(arg, "f=") || strings.Contains(arg, ".srt") {
			filePath = strings.Split(arg, "=")[1]
		} else if strings.Contains(arg, "start=") {
			timeToSet = strings.Split(arg, "=")[1]
		} else if strings.Contains(arg, "num=") {
			numStr = strings.Split(arg, "=")[1]
		} else if strings.Contains(arg, "readsrt") {
			readFile = true
		} else if strings.Contains(arg, "help") {
			fmt.Println(helpMessage)
			return
		}
	}

	if !strings.Contains(filePath, ".srt") {
		fmt.Println("The subtitle file should be a .srt file.")
		return
	}
	fileName := strings.Join(strings.Split(filePath, ".")[:len(strings.Split(filePath, "."))-1], ".")
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}

	byt, err := io.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}

	strList := strings.Split(string(byt), "\n")

	if readFile {
		for i := 0; i < 100; i++ {
			fmt.Println(strList[i])
		}
		return
	}
	if filePath == "" || timeToSet == "" || numStr == "" {
		fmt.Println(helpMessage)
		return
	}
	subList := [][]string{}
	sub := []string{}

	for _, subStr := range strList {
		if len([]byte(subStr)) == 1 && []byte(subStr)[0] == 13 {
			subList = append(subList, sub)
			sub = []string{}
		} else {
			sub = append(sub, subStr)
		}

	}

	subtitles := []Subtitle{}

	for _, sub := range subList {
		oneSub := Subtitle{}
		mainSub := ""
		for i, s := range sub {
			if i == 0 {
				s = strings.Trim(s, "\r")
				num, err := strconv.ParseInt(s, 10, 64)
				if err != nil {
					fmt.Println(err)
				}
				oneSub.Number = int(num)
			} else if i == 1 {
				ss := strings.Split(s, " --> ")
				startSplit := strings.Split(strings.Trim(ss[0], "\r"), ",")
				endSplit := strings.Split(strings.Trim(ss[1], "\r"), ",")

				startTimeSplit := strings.Split(startSplit[0], ":")
				endTimeSplit := strings.Split(endSplit[0], ":")

				oneSub.StartHour = inti(startTimeSplit[0])
				oneSub.StartMin = inti(startTimeSplit[1])
				oneSub.StartSec = inti(startTimeSplit[2])
				oneSub.StartMili = inti(startSplit[1])

				oneSub.EndHour = inti(endTimeSplit[0])
				oneSub.EndMin = inti(endTimeSplit[1])
				oneSub.EndSec = inti(endTimeSplit[2])
				oneSub.EndMili = inti(endSplit[1])
			} else {
				s = string([]byte(s)[:len([]byte(s))-1])
				mainSub = mainSub + s + "\n"
			}

			if i == len(sub)-1 {
				oneSub.SubtitleString = strings.Trim(mainSub, "\n")
			}
		}
		subtitles = append(subtitles, oneSub)
	}
	file.Close()

	startSplit := strings.Split(timeToSet, ",")
	startTimeSplit := strings.Split(startSplit[0], ":")

	h := inti(startTimeSplit[0])
	m := inti(startTimeSplit[1])
	s := inti(startTimeSplit[2])
	mili := inti(startSplit[1])
	num := inti(numStr)

	ha := 0
	ma := 0
	sa := 0
	miliDiff := 0

	for _, sub := range subtitles {
		if sub.Number == num {
			ha = h - sub.StartHour
			ma = m - sub.StartMin
			sa = s - sub.StartSec

			miliDiff = mili - sub.StartMili
		}
	}

	subs := []Subtitle{}
	for _, sub := range subtitles {
		nh, nm, ns, mili := ChangeTime(sub.StartHour, sub.StartMin, sub.StartSec, sub.StartMili, ha, ma, sa, miliDiff)
		nh2, nm2, ns2, mili2 := ChangeTime(sub.EndHour, sub.EndMin, sub.EndSec, sub.EndMili, ha, ma, sa, miliDiff)

		sub.StartHour = nh
		sub.StartMin = nm
		sub.StartSec = ns
		sub.StartMili = mili
		sub.EndHour = nh2
		sub.EndMin = nm2
		sub.EndSec = ns2
		sub.EndMili = mili2

		subs = append(subs, sub)
	}

	newFilePath := fileName + "SRTEditer.srt"
	file2, err := os.Create(newFilePath)
	if err != nil {
		fmt.Println(err)
	}

	for _, sub := range subs {
		file2.WriteString(fmt.Sprintln(sub.Number))
		file2.WriteString(fmt.Sprintf("%v:%v:%v,%v --> %v:%v:%v,%v\n", sub.StartHour, sub.StartMin, sub.StartSec, sub.StartMili, sub.EndHour, sub.EndMin, sub.EndSec, sub.EndMili))
		file2.WriteString(sub.SubtitleString + "\n\n")
	}

	fmt.Println("New subtitle created succussfully complete!\nFile name is " + newFilePath)
}
