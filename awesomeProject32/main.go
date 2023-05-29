package main

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"os"
	"strconv"
)


func main() {

	way := "C:\\Users\\buGeo\\go\\go1.18.4\\src\\awesomeProject32\\files\\"

	//way := "root/Shedule/files/"
	week := createWeek()

	days := parseDays(week, way)
	bot, _ := tgbotapi.NewBotAPI("")

	usersIds, allDeleteIds := getUsers(days,bot, way)

	go dayInfo(days, &usersIds,&allDeleteIds, bot, way, week)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	var resp string

	for update := range updates {
		var deleteIds []int
		if update.Message != nil {
			if !check(usersIds, update.Message.Chat.ID) {
				usersIds = append(usersIds,update.Message.Chat.ID )
				allDeleteIds = append(allDeleteIds, []int{})
				file, _ := os.OpenFile(way +"ids", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644 )
				file.WriteString(strconv.FormatInt(update.Message.Chat.ID, 10)+"\n")
				file.Close()
				os.Create(way + strconv.FormatInt(update.Message.Chat.ID, 10))

			}
			deleteIds = append(deleteIds, update.Message.MessageID)
			_, id := getDay(days, update.Message.Text)
			var msgEnt []tgbotapi.MessageEntity
			sleep := 0

			if id != -1 {
				resp, msgEnt = generateResp(id, "weeks", days)
				sleep = 150
			} else {
				resp = "Введён неправильный поисковый запрос."
				sleep = 10
			}
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, resp)
			msg.Entities = msgEnt
			inMsg,_ := bot.Send(msg)
			deleteIds = append(deleteIds, inMsg.MessageID)
			file, _ := os.OpenFile(way + strconv.FormatInt(update.Message.Chat.ID, 10), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644 )
			for _, deleteId := range deleteIds{
				file.WriteString(strconv.Itoa(deleteId)+"\n")
			}
			file.Close()
			go deleteSleep(update.Message.Chat.ID,deleteIds,bot,sleep, way)
		}
	}
}
