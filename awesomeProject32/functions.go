package main

import (
	"encoding/csv"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/xuri/excelize/v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

type Day struct {
	date           time.Time
	weekDay        string
	couples        [][]string
	time           [][]time.Time
	font           [][]string
	b              int
	tmpStr         int
	timeEx         time.Duration
	dayNumber      int
	couplesNumber  int
	couplesNumbers []int
	g              int
}

type Week struct {
	week map[string][]string
}

func readCsvFile(filePath string) [][]string {
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Unable to read input file "+filePath, err)
	}
	defer f.Close()
	csvReader := csv.NewReader(f)
	csvReader.Comma = ';'
	csvReader.LazyQuotes = true
	records, err := csvReader.ReadAll()
	if err != nil {
		log.Fatal("Unable to parse file as CSV for "+filePath, err)
	}

	return records
}

func createWeek() Week {
	var week Week
	week.week = make(map[string][]string)
	week.week["Monday"] = []string{"понедельник", "0"}
	week.week["Tuesday"] = []string{"вторник", "1"}
	week.week["Wednesday"] = []string{"среда", "2"}
	week.week["Thursday"] = []string{"четверг", "3"}
	week.week["Friday"] = []string{"пятница", "4"}
	week.week["Saturday"] = []string{"суббота", "5"}
	week.week["Sunday"] = []string{"воскресенье", "6"}
	return week
}

func fmtDuration(d time.Duration) string {
	d = d.Round(time.Minute)
	days := d / (time.Hour * 24)
	d -= days * 24 * time.Hour
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	str := ""
	if days != 0 {
		str += fmt.Sprintf("%1dд", days)
	}
	if h != 0 {
		if len(str) > 0 {
			str += ":"
		}
		str += fmt.Sprintf("%1dч", h)
	}
	if m != 0 {
		if len(str) > 0 {
			str += ":"
		}
		str += fmt.Sprintf("%1dм", m)
	}
	return str

}

func delete(userId int64, deleteIds []int, bot *tgbotapi.BotAPI, way string) {
	file, _ := os.OpenFile(way+strconv.FormatInt(userId, 10), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	data, _ := ioutil.ReadFile(way + strconv.FormatInt(userId, 10))
	ids := strings.Split(string(data), "\n")
	file.Close()
	file, _ = os.Create(way + strconv.FormatInt(userId, 10))
	if len(ids) > 0 {
		for _, id := range ids[:len(ids)-1] {
			intId, _ := strconv.Atoi(id)
			if !checkMsgs(deleteIds, intId) {
				file.WriteString(id + "\n")
			} else {
				msg := tgbotapi.NewDeleteMessage(userId, intId)
				bot.Send(msg)
			}
		}
	}
	file.Close()
}

func parseDays(week Week, way string) []Day {

	fileName := way + "test.xlsx"
	out, _ := os.Create(fileName)
	resp, _ := http.Get("")
	io.Copy(out, resp.Body)
	out.Close()
	resp.Body.Close()

	sheetName := "Лист1"
	dateCell := "A"
	coupleNumber := "B"
	coupleCell := "E"

	f, _ := excelize.OpenFile(fileName)
	defer f.Close()

	raws, _ := f.GetRows(sheetName)

	var firstCell string
	var secondCell string
	var thirdCell string
	var days []Day
	var day Day
	var nextDate time.Time
	var lastTime time.Time

	var strNumber string

	style, _ := f.NewStyle(`{"number_format":14}`)

	for i := 3; i <= len(raws); i++ {
		strNumber = strconv.Itoa(i)
		f.SetCellStyle(sheetName, dateCell+strNumber, dateCell+strNumber, style)
		firstCell, _ = f.GetCellValue("Лист1", dateCell+strNumber)
		secondCell, _ = f.GetCellValue("Лист1", coupleNumber+strNumber)
		thirdCell, _ = f.GetCellValue("Лист1", coupleCell+strNumber)

		var t time.Time
		t, _ = time.ParseInLocation("01-02-06", firstCell, time.Now().Location())
		for {
			if nextDate.Year() > 1000 && t != nextDate {
				day.weekDay = week.week[nextDate.Weekday().String()][0]
				day.dayNumber, _ = strconv.Atoi(week.week[nextDate.Weekday().String()][1])
				day.date = nextDate
				couple := []string{"Выходной"}
				day.couples = append(day.couples, couple)
				day.couplesNumbers = append(day.couplesNumbers, day.couplesNumber)
				day.time = append(day.time, []time.Time{lastTime, day.date.Add(15 * time.Hour)})
				day.b = 0
				day.tmpStr = -1
				nextDate = nextDate.AddDate(0, 0, 1)
				days = append(days, day)
				day = Day{}
			} else {
				break
			}
		}
		day.weekDay = week.week[t.Weekday().String()][0]
		day.dayNumber, _ = strconv.Atoi(week.week[t.Weekday().String()][1])
		day.date, _ = time.ParseInLocation("01-02-06", firstCell, time.Now().Location())
		dur := strings.Split(secondCell, "\n")[1]
		startTime, _ := time.ParseInLocation("01-02-06 15:04", firstCell+" "+strings.Split(dur, " - ")[0], time.Now().Location())
		endTime, _ := time.ParseInLocation("01-02-06 15:04", firstCell+" "+strings.Split(dur, " - ")[1], time.Now().Location())
		secondCell = strings.Split(secondCell, "\n")[0] + " (" + strings.Split(secondCell, "\n")[1] + ")"
		secondCell = strings.Replace(secondCell, " (3)", "", 1)
		if len(thirdCell) > 0 {
			if lastTime.Year() > 1000 {
			} else {
				lastTime = day.date.Add(-9 * time.Hour)
			}
			m := startTime.Sub(lastTime)
			if m.Minutes() != 0 {
				day.b = 1
				day.couples = append(day.couples, []string{"перерыв", fmtDuration(m)})
				day.couplesNumbers = append(day.couplesNumbers, day.couplesNumber)
				day.time = append(day.time, []time.Time{lastTime, startTime})
			}
			lastTime = endTime
			day.couples = append(day.couples, []string{secondCell, thirdCell})
			day.time = append(day.time, []time.Time{startTime, endTime})
			day.couplesNumbers = append(day.couplesNumbers, day.couplesNumber)
			day.couplesNumber += 1
		}
		if (i-2)%6 == 0 {
			nextDate = t.AddDate(0, 0, 1)
			if len(day.couples) == 0 {
				couple := []string{"Выходной"}
				day.couples = append(day.couples, couple)
				day.couplesNumbers = append(day.couplesNumbers, day.couplesNumber)
				day.weekDay = week.week[day.date.Weekday().String()][0]
				day.dayNumber, _ = strconv.Atoi(week.week[day.date.Weekday().String()][1])
				day.b = 0
				day.tmpStr = -1
				day.time = append(day.time, []time.Time{lastTime, day.date.Add(15 * time.Hour)})
			}
			days = append(days, day)
			//lastTime = time.Time{}
			day = Day{}
		}
	}
	return days
}

func getEnt(font string, offset string, length string) tgbotapi.MessageEntity {
	var newEnt tgbotapi.MessageEntity
	newEnt.Type = font
	newEnt.Offset = utf8.RuneCountInString(offset)
	newEnt.Length = utf8.RuneCountInString(length)
	return newEnt
}

func getResp(dayNumber int, id int, days []Day, userId int64, reqType string) (tgbotapi.MessageConfig, tgbotapi.MessageConfig) {
	idStart := id-dayNumber
	idEnd := idStart+13
	var newResp string
	var newmMsgEnt []tgbotapi.MessageEntity
	var resp string
	var msgEnt []tgbotapi.MessageEntity
	idEndTwo := idEnd
	if len(days)-1 - (id + 1+(6-dayNumber))<6{
		idEndTwo-=6 - len(days)-1 - (id + 1+(6-dayNumber))
	}
	for i := idEndTwo; i >= idEnd-6; i-- {
		newResp, newmMsgEnt = generateResp(i, reqType, days)
		for _, ent := range newmMsgEnt {
			ent.Offset += utf8.RuneCountInString(resp)
			msgEnt = append(msgEnt, ent)
		}
		resp += newResp
	}
	firstMsg := tgbotapi.NewMessage(userId, resp)
	firstMsg.Entities = msgEnt
	resp = ""
	msgEnt = []tgbotapi.MessageEntity{}
	if len(days)-1 - (id + (6-dayNumber))<0{
		idEnd -=(id + (6-dayNumber)) - len(days)-1
	}
	if id-dayNumber<0{
		idStart=0
	}
	for i := idStart+dayNumber; i >= idStart; i-- {
		newResp, newmMsgEnt = generateResp(i, reqType, days)
		for _, ent := range newmMsgEnt {
			ent.Offset += utf8.RuneCountInString(resp)
			msgEnt = append(msgEnt, ent)
		}
		resp += newResp
	}
	for i := idEnd - 7; i >= idStart+dayNumber+1; i-- {
		newResp, newmMsgEnt = generateResp(i, reqType, days)
		for _, ent := range newmMsgEnt {
			ent.Offset += utf8.RuneCountInString(resp)
			msgEnt = append(msgEnt, ent)
		}
		resp += newResp
	}
	secondMsg := tgbotapi.NewMessage(userId, resp)
	secondMsg.Entities = msgEnt
	return firstMsg, secondMsg
}

func sendWeeks(bot *tgbotapi.BotAPI, dayNumber int, days []Day, userId int64, id int) []int {
	var msgIds []int
	reqType := "weeks"
	firstMsg, secondMsg := getResp(dayNumber, id, days, userId, reqType)
	if len(firstMsg.Text)!=0{
		inMsg, _ := bot.Send(firstMsg)
		msgIds = append(msgIds, inMsg.MessageID)
	}
	if len(secondMsg.Text)!=0{
		inMsg, _ := bot.Send(secondMsg)
		msgIds = append(msgIds, inMsg.MessageID)
	}
	return msgIds
}

func sendDay(bot *tgbotapi.BotAPI, days []Day, id int, userID int64) int {
	reqType := "day"
	resp, ent := generateResp(id, reqType, days)
	var inMsg tgbotapi.Message
	msg := tgbotapi.NewMessage(userID, resp)
	msg.Entities = ent
	inMsg, _ = bot.Send(msg)
	return inMsg.MessageID
}

func deleteSleep(userId int64, deleteIds []int, bot *tgbotapi.BotAPI, sleep int, way string) {
	time.Sleep(time.Duration(sleep) * time.Second)
	fmt.Println(deleteIds)
	delete(userId, deleteIds, bot, way)
}

func getUsers(days []Day, bot *tgbotapi.BotAPI, way string) ([]int64, [][]int) {
	var usersIds []int64
	var allDeleteIds [][]int
	fFile, err := os.OpenFile(way+"ids", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fFile, err = os.Create(way + "ids")
	}

	fContent, err := ioutil.ReadFile(way + "ids")
	textUsersIds := strings.Split(string(fContent), "\n")
	if len(textUsersIds) > 0 {
		for _, textUserId := range textUsersIds[:len(textUsersIds)-1] {
			userId, _ := strconv.ParseInt(textUserId, 10, 64)
			usersIds = append(usersIds, userId)
			allDeleteIds = append(allDeleteIds, []int{})
			sFile, err := os.OpenFile(way+strconv.FormatInt(userId, 10), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				sFile, err = os.Create(way + strconv.FormatInt(userId, 10))
			}
			sContent, err := ioutil.ReadFile(way + strconv.FormatInt(userId, 10))
			var msgIds []int
			textMsgsIds := strings.Split(string(sContent), "\n")
			if len(textMsgsIds) > 0 {
				for _, textMsgId := range textMsgsIds[:len(textMsgsIds)-1] {
					msgId, _ := strconv.Atoi(textMsgId)
					msgIds = append(msgIds, msgId)
				}
				delete(userId, msgIds, bot, way)
			}
			sFile.Close()
		}
	}
	fFile.Close()
	return usersIds, allDeleteIds
}

func dayInfo(days []Day, userIds *[]int64, allDeleteIds *[][]int, bot *tgbotapi.BotAPI, way string, week Week) {

	var fontClear []int
	var finId, sunId, tmpStr, b int
	var timeTo time.Time
	var t, y time.Time
	var sun, newDay bool
	twoCheck := true
	id := -1
	//test
	//t, _ = time.ParseInLocation("02.01.2006 15:04", "30.05.2022 09:55", time.Now().Location())
	//_, id = getDay(days, t.Format("02.01.2006"))
	//timeTo = t
	//c := 0
	//
	for {

		//Test
		//switch c {
		//case 0:
		//	t, _ = time.ParseInLocation("02.01.2006 15:04", "04.05.2022 14:55", time.Now().Location())
		//	_, id = getDay(days, t.Format("02.01.2006"))
		//	timeTo = t
		//	c = 1
		//	break
		//case 1:
		//	t, _ = time.ParseInLocation("02.01.2006 15:04", "04.05.2022 14:56", time.Now().Location())
		//	timeTo = t
		//	time.Sleep(time.Second*20)
		//	c = 6
		//	break
		//case 2:
		//	t, _ = time.ParseInLocation("02.01.2006 15:04", "13.05.2022 11:05", time.Now().Location())
		//	timeTo = t
		//	time.Sleep(time.Second*20)
		//	c++
		//	break
		//case 3:
		//	t, _ = time.ParseInLocation("02.01.2006 15:04", "13.05.2022 13:30", time.Now().Location())
		//	timeTo = t
		//	time.Sleep(time.Second*20)
		//	break
		//}
		//if !newDay {
		//	time.Sleep(time.Second * 5)
		//	timeTo = timeTo.Add(time.Minute)
		//	t = timeTo
		//} else {
		//	newDay = false
		//}
		//
		if id == -1 || id >= len(days) {
			if time.Now().Hour() >= 15 {
				_, id = getDay(days, time.Now().AddDate(0, 0, 1).Format("02.01.2006"))
			} else {
				_, id = getDay(days, time.Now().Format("02.01.2006"))
			}
			if id != -1 {
				if twoCheck == false {
					for m, userId := range *userIds {
						delete(userId, (*allDeleteIds)[m], bot, way)
						(*allDeleteIds)[m] = []int{}
					}
				}
				twoCheck = true
			}
		}
		if id != -1 {
			if !newDay {
				timeTo = time.Now().Add(time.Minute).Truncate(1 * time.Minute)
				sleep := timeTo.Sub(time.Now())
				time.Sleep(sleep)
			} else {
				newDay = false
			}
			t = time.Now()
			y = t.AddDate(0, 0, 1)
			check := false
			var i int
			var couple []time.Time

			for i, couple = range days[id].time {
				if t.Sub(couple[0]) >= 0 && t.Sub(couple[1]) < 0 {
					days[id].font = append(days[id].font, []string{"underline", strconv.Itoa(i)})
					fontClear = append(fontClear, id)
					check = true
					g := strings.Split(days[id].couples[i][0], " ")
					days[id].g = days[id].couplesNumber - days[id].couplesNumbers[i]
					if len(g) > 1 {
						tmpStr = 2
					} else {
						if i == 0 {
							tmpStr = 1
						} else {
							tmpStr = 3
						}
					}
					b = i
					break
				}
			}
			if days[id].couples[0][0] == "Выходной" {
				sun = true
				tmpStr = -1
			}
			if i < len(days[id].time)-1 {
				days[id].font = append(days[id].font, []string{"italic", strconv.Itoa(i + 1)})
				fontClear = append(fontClear, id)
			} else {
				if t.Sub(days[id].time[0][0]) < 0 {
					y = t
				} else {
					if !check {
						if sun {
							days[id].b = 0
						} else {
							days[id].b = 1
						}
						days[id].tmpStr = -1
						days[id].g = 0
						id += 1
						if sunId == id {
							sunId = 0
							sun = false
						}
						b = 0
						for m, userId := range *userIds {
							delete(userId, (*allDeleteIds)[m], bot, way)
							(*allDeleteIds)[m] = []int{}
						}
						days = parseDays(week, way)
						newDay = true
						continue
					}
				}

				for {

					_, tmpId := getDay(days, y.Format("02.01.2006"))
					if tmpId == -1 {
						break
					}
					if days[tmpId].couples[0][0] != "Выходной" {
						days[tmpId].font = append(days[tmpId].font, []string{"italic", "1"})
						if sun {
							days[tmpId].b = 0
							days[tmpId].tmpStr = 1
							sunId = tmpId
						}
						break
					}
					y = y.AddDate(0, 0, 1)
				}
			}
			//test2,_:= time.ParseInLocation("02.01.2006 15:04:05", "13.05.2022 11:29:50", time.Now().Location())
			dayNumber := days[id].dayNumber
			days[id].b = b
			days[id].tmpStr = tmpStr
			if sunId != 0 {
				finId = sunId
				days[finId].font = [][]string{{"italic", "1"}}
			} else {
				finId = id
			}
			days[finId].timeEx = days[finId].time[days[finId].b][1].Sub(timeTo)
			for m, userId := range *userIds {
				deleteIds := (*allDeleteIds)[m]
				if len(deleteIds) == 0 {
					days[id].tmpStr = tmpStr
					deleteIds = sendWeeks(bot, dayNumber, days, userId, id)
					deleteIds = append(deleteIds, sendDay(bot, days, id, userId))
					//os.Create(way + strconv.FormatInt(userId, 10))
					file, _ := os.OpenFile(way+strconv.FormatInt(userId, 10), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
					for _, id := range deleteIds {
						file.WriteString(strconv.Itoa(id) + "\n")
					}
					file.Close()
					(*allDeleteIds)[m] = deleteIds
				} else {
					editWeeks(userId, deleteIds, bot, dayNumber, days, id)
					editDay(userId, deleteIds, bot, days, id)
				}
			}
			dayClear(fontClear, days)
		} else {
			if twoCheck {
				for m, userId := range *userIds {
					deleteIds := (*allDeleteIds)[m]
					msg := tgbotapi.NewMessage(userId, "Такого дня нет в твоём расписании.")
					inMsg, _ := bot.Send(msg)
					deleteIds = append(deleteIds, inMsg.MessageID)
					(*allDeleteIds)[m] = deleteIds
					twoCheck = false
				}
			}
		}
	}
	//var test time.Time
	//if c < 2 {
	//	test,_ = time.ParseInLocation("02.01.2006 15:04:05", "12.05.2022 14:49:55", time.Now().Location())
	//} else {
	//	test,_ = time.ParseInLocation("02.01.2006 15:04:05", "13.05.2022 12:49:55", time.Now().Location())
	//}
}

func dayClear(fontClear []int, days []Day) {
	for _, fontCl := range fontClear {
		days[fontCl].font = [][]string{}
	}
}

func editWeeks(userId int64, deleteIds []int, bot *tgbotapi.BotAPI, dayNumber int, days []Day, id int) {
	//if days[id].b > 0{
	//	day.b = 1
	//	day.tmpStr = -1
	//}
	reqType := "weeks"
	firstMsg, secondMsg := getResp(dayNumber, id, days, userId, reqType)
	if len(firstMsg.Text)!=0{
		msg := tgbotapi.NewEditMessageText(userId, deleteIds[0], firstMsg.Text)
		msg.Entities = firstMsg.Entities
		bot.Send(msg)
	}
	if len(secondMsg.Text)!=0 {
		msg := tgbotapi.NewEditMessageText(userId, deleteIds[1], secondMsg.Text)
		msg.Entities = secondMsg.Entities
		bot.Send(msg)
	}
}

func generateResp(id int, reqType string, days []Day) (string, []tgbotapi.MessageEntity) {
	b := days[id].b
	tmpStr := days[id].tmpStr
	timeEx := days[id].timeEx

	var msgEnt []tgbotapi.MessageEntity
	var resp string

	resp = days[id].date.Format("02.01.2006") + " (" + days[id].weekDay + ") "
	if days[id].couplesNumber != 0 {
		if reqType == "day" {
			resp += strconv.Itoa(days[id].g)
		} else {
			resp += strconv.Itoa(days[id].couplesNumber)
		}
	} else {
		resp += "нет"
	}
	resp += " п.\n\n"
	msgEnt = append(msgEnt, getEnt("bold", "", resp))
	for i, couple := range days[id].couples {
		if reqType == "day" {
			if i < b {
				continue
			}
		} else {
			if i == 0 && b != 0 {
				continue
			}
		}
		firstPart := couple[0]
		secondPart := ""
		if len(couple) > 1 {
			firstPart = couple[0] + " - "
			secondPart = couple[1]
		}
		if i == b {
			if !(reqType == "weeks" && b != 0) {
				switch tmpStr {
				case 1:
					firstPart = "до начала пар осталось: "
					secondPart = fmtDuration(timeEx)
					break
				case 2:
					firstPart = strings.Split(couple[0], " ")[0] + " " + strings.Split(couple[0], " ")[1] + " (осталось: " + fmtDuration(timeEx) + ") - "
					secondPart = couple[1]
					break
				case 3:
					firstPart = "до конца перерыва осталось:  "
					secondPart = fmtDuration(timeEx)
					break
				}
			}
		}
		msgEnt = append(msgEnt, getEnt("bold", resp, firstPart))
		newResp := firstPart + secondPart
		for _, font := range days[id].font {
			for _, elem := range font[1:] {
				id, _ := strconv.Atoi(elem)
				if i == id {
					msgEnt = append(msgEnt, getEnt(font[0], resp+firstPart, secondPart))
				}
			}
		}
		resp += newResp + "\n\n"
	}
	resp += "\n\n"
	return resp, msgEnt
}

func editDay(userId int64, deleteIds []int, bot *tgbotapi.BotAPI, days []Day, id int) {
	reqType := "day"
	thirdMsgText, ent := generateResp(id, reqType, days)

	msg := tgbotapi.NewEditMessageText(userId, deleteIds[2], thirdMsgText)
	msg.Entities = ent
	bot.Send(msg)

}

func check(arr []int64, element int64) bool {
	result := false
	for _, x := range arr {
		if x == element {
			result = true
			break
		}
	}
	return result
}

func checkMsgs(arr []int, element int) bool {
	result := false
	for _, x := range arr {
		if x == element {
			result = true
			break
		}
	}
	return result
}

func getDay(days []Day, date string) (Day, int) {
	for id, day := range days {
		if strings.Contains(day.date.Format("02.01.2006"), date) {
			return day, id
		}
	}
	return Day{}, -1
}
