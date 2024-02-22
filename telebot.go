package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	tele "gopkg.in/telebot.v3"
)

var HelpText = "Welcome!\n\n" +
	"Here are some available commands:\n" +
	"/config - Show the current configuration\n" +
	"/start - Start the bot\n" +
	"/stop - Stop the bot\n" +
	"/restart - Restart the bot\n" +
	"/setcron - The frequency the bot checks for updates in a cron expression format\nexample: ```/setcron 0 12/24 * * *```\nTo understand cron expressions visit: https://crontab.guru/\n" +
	"/addrifs - Add RIF identifiers to search by: \nexample: ```/addrifs G200003391,G200038179```\n" +
	"/removerifs - Remove RIF identifiers to search by: \nexample: ```/removerifs G200003391,G200038179```\n" +
	"/addkeywords - Add keywords to search by: \nexample: ```/addkeywords PETRO,GAS,TUBERIA```\n" +
	"/removekeywords - Remove keywords to search by: \nexample: ```/removekeywords PETRO,GAS,TUBERIA```\n" +
	"/setfromdate - The starting date to filter by: \nexample: ```/setfromdate 22/01/2006```\n" +
	"/help - Show this help message"

func isValidChatIdMiddleWare(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		if !ChatID(c.Chat().ID).isValid() {
			return c.Send(fmt.Sprintf("Access denied...\nChatId=%d", c.Chat().ID))
		}

		return next(c)
	}
}

func requestLoggerMiddleWare(next tele.HandlerFunc) tele.HandlerFunc {
	return func(c tele.Context) error {
		fmt.Printf("[DEBUG] [%s] [%d] [%s]\n", c.Message().Text, c.Chat().ID, c.Chat().Username)

		return next(c)
	}
}

func sendContractReport(b *tele.Bot, c tele.Context) {
	var err error
	var contracts []Contract

	sendToChats := func(message string) {
		chatids := GetWhiteListedChatIds()

		for _, ID := range chatids {
			_, err = b.Send(&tele.Chat{ID: ID}, message)

			if err != nil {
				log.Fatal(err)
			}
		}
	}

	contracts, err = ScrapContracts()

	if err != nil {
		sendToChats(fmt.Sprintf("Error while fetching table. %s", err))
		// _, err = b.Send(&tele.Chat{ID: c.Chat().ID}, fmt.Sprintf("Error while fetching table. %s", err))

		// if err != nil {
		// 	log.Fatal(err)
		// }
	}

	filteredContracts := []Contract{}

	for _, contract := range contracts {
		constractDate, _ := time.Parse("02/01/2006", contract.Date)

		if !constractDate.After(Config.fromDate) {
			// fmt.Printf("[DEBUG] Discarding contract by date: %s\n", contract.Date)
			continue
		}

		if len(Config.RIFs) > 0 && !contract.RIF.includes(Config.RIFs) {
			// fmt.Printf("[DEBUG] Discarding contract by RIF: %s\n", contract.RIF.String())
			continue
		}

		if len(Config.keywords) > 0 && !contract.Description.has(Config.keywords) {
			// fmt.Println("[DEBUG] Discarding contract by description")
			continue
		}

		filteredContracts = append(filteredContracts, contract)
	}

	for _, contract := range filteredContracts {
		sendToChats(contract.HumanReadable())
	}

	Config.UpdateFromDate()
}

func StartTelebot() {
	cronJob := cron.New()

	b, err := tele.NewBot(tele.Settings{
		Token:  os.Getenv("BOT_TOKEN"),
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})

	b.Use(isValidChatIdMiddleWare)
	b.Use(requestLoggerMiddleWare)
	// b.Use(middleware.Logger())

	if err != nil {
		log.Fatal(err)

		return
	}

	b.Handle("/help", func(c tele.Context) error {
		return c.Send(HelpText)
	})

	b.Handle("/config", func(c tele.Context) error {
		entries := cronJob.Entries()

		if len(entries) > 0 {
			c.Send(fmt.Sprintf("Next run: %s\n", entries[0].Next.String()))
		}

		return c.Send("Configuration:\n" + Config.HumanReadable())
	})

	b.Handle("/setcron", func(c tele.Context) error {
		args := strings.Fields(c.Message().Text)

		if len(args) < 6 {
			return c.Send("Argument required!\nexample: ```/setcron 0 12/24 * * *```")
		}

		Config.UpdateCronTab(strings.Join(args[1:], " "))

		return c.Send(fmt.Sprintf("Current cron [%s]", Config.crontab))
	})

	b.Handle("/addrifs", func(c tele.Context) error {
		args := strings.Fields(c.Message().Text)

		if len(args) < 2 {
			return c.Send("Argument required!\nexample: ```/addrifs G200003391,J200003392```")
		}

		for _, rif := range strings.Split(args[1], ",") {
			Config.AddRif(rif)
		}

		return c.Send(fmt.Sprintf("Current RIFs %s", Config.RIFs))
	})

	b.Handle("/removerifs", func(c tele.Context) error {
		args := strings.Fields(c.Message().Text)

		if len(args) < 2 {
			return c.Send("Argument required!\nexample: ```/removerifs G200003391,J200003392```")
		}

		for _, rif := range strings.Split(args[1], ",") {
			Config.RemoveRif(rif)
		}

		return c.Send(fmt.Sprintf("Current RIFs: %s", Config.RIFs))
	})

	b.Handle("/addkeywords", func(c tele.Context) error {
		args := strings.Fields(c.Message().Text)

		if len(args) < 2 {
			return c.Send("Argument required!\nexample: ```/addkeywords PETRO,GAS```")
		}

		for _, keyword := range strings.Split(args[1], ",") {
			Config.AddKeyword(keyword)
		}

		return c.Send(fmt.Sprintf("Current keywords %s", Config.keywords))
	})

	b.Handle("/removekeywords", func(c tele.Context) error {
		args := strings.Fields(c.Message().Text)

		if len(args) < 2 {
			return c.Send("Argument required!\nexample: ```/removekeywords PETRO,GAS```")
		}

		for _, keyword := range strings.Split(args[1], ",") {
			Config.RemoveKeyword(keyword)
		}

		return c.Send(fmt.Sprintf("Current keywords %s", Config.keywords))
	})

	b.Handle("/setfromdate", func(c tele.Context) error {
		args := strings.Fields(c.Message().Text)

		if len(args) < 2 {
			return c.Send("Argument required!\nexample: ```/setkeywords PETRO,GAS```")
		}

		fromDate, _ := time.Parse("02/01/2006", args[1])

		Config.fromDate = fromDate

		return c.Send(fmt.Sprintf("Successful from date set [%s]", fromDate))
	})

	b.Handle("/start", func(c tele.Context) error {
		entries := cronJob.Entries()

		if len(entries) > 0 {
			return c.Send("Skipping job already started!")
		}

		sendContractReport(b, c)

		fmt.Printf("[DEBUG] /start using crontab: %s\n", Config.crontab)

		cronJob.AddFunc(Config.crontab, func() {
			sendContractReport(b, c)
		})

		cronJob.Start()

		c.Send(fmt.Sprintf("Next run: %s\n", cronJob.Entries()[0].Next.String()))

		fmt.Println(Config.HumanReadable())

		return c.Send("Report finished!\n" + Config.HumanReadable())
	})

	b.Handle("/restart", func(c tele.Context) error {
		entries := cronJob.Entries()

		if len(entries) == 0 {
			return c.Send("Job already stopped!")
		}

		c.Send("Shutting down cron job...")

		for _, entry := range entries {
			cronJob.Remove(entry.ID)
		}

		cronJob.Stop()

		c.Send("Restarting job...")

		sendContractReport(b, c)

		fmt.Printf("[DEBUG] /restart using crontab: %s\n", Config.crontab)

		cronJob.AddFunc(Config.crontab, func() {
			sendContractReport(b, c)
		})

		cronJob.Start()

		c.Send(fmt.Sprintf("Next run: %s\n", cronJob.Entries()[0].Next.String()))

		fmt.Println(Config.HumanReadable())

		return c.Send("Report finished!\n" + Config.HumanReadable())
	})

	b.Handle("/stop", func(c tele.Context) error {
		entries := cronJob.Entries()

		if len(entries) == 0 {
			return c.Send("Job already stopped!")
		}

		c.Send("Shutting down cron job...")

		for _, entry := range entries {
			cronJob.Remove(entry.ID)
		}

		cronJob.Stop()

		return c.Send("Job stopped...")
	})

	b.Start()
}
