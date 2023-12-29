package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

/**
 * Configuration
 */

type Configuration struct {
	RIFs     []string
	keywords []string
	// schedule Schedule
	fromDate time.Time
	crontab  string
}

func (c *Configuration) AddRif(rif string) {
	c.RIFs = append(Config.RIFs, strings.ToUpper(rif))
}

func (c *Configuration) RemoveRif(rif string) {
	// Find the index of the element
	index := -1
	for i, value := range c.RIFs {
		if value == strings.ToUpper(rif) {
			index = i
			break
		}
	}

	// If the element was found, remove it
	if index != -1 {
		c.RIFs = append(c.RIFs[:index], c.RIFs[index+1:]...)
	}
}

func (c *Configuration) AddKeyword(keyword string) {
	c.keywords = append(Config.keywords, strings.ToUpper(keyword))
}

func (c *Configuration) RemoveKeyword(keyword string) {
	// Find the index of the element
	index := -1
	for i, value := range c.keywords {
		if value == strings.ToUpper(keyword) {
			index = i
			break
		}
	}

	// If the element was found, remove it
	if index != -1 {
		c.keywords = append(c.keywords[:index], c.keywords[index+1:]...)
	}
}

func (c Configuration) HumanReadable() string {
	return fmt.Sprintf("RIFs: %s\nKeywords: %s\nCrontab: %s\nFromDate: %s", c.RIFs, c.keywords, c.crontab, c.fromDate.Format("02/01/2006"))
}

func (c *Configuration) UpdateFromDate() {
	c.fromDate = time.Now().AddDate(0, 0, -1)
}

func (c *Configuration) UpdateCronTab(cronExp string) {
	c.crontab = cronExp
}

var Config = Configuration{
	// schedule: Schedule(24 * time.Hour),
	fromDate: time.Now().AddDate(0, 0, -1),
	crontab:  "0 12/24 * * *",
}

/**
 * Schedule
 */

// type Schedule time.Duration

// func (s Schedule) update(duration time.Duration) {
// 	Config.schedule = Schedule(duration)
// }

// func (s Schedule) durationToCron() string {
// 	hours := int(s.Duration().Hours())

// 	return fmt.Sprintf("0 12/%d * * *", hours)
// }

// func (s Schedule) String() string {
// 	return time.Duration(s).String()
// }

// func (s Schedule) Duration() time.Duration {
// 	return time.Duration(s)
// }

/**
 * ChatID
 */

type ChatID int64

func (uid ChatID) int64() int64 {
	return int64(uid)
}

func (uid ChatID) isValid() bool {
	groupIdsStr := os.Getenv("CHAT_IDS")

	groupIdsStrs := strings.Split(groupIdsStr, ",")

	// Create a slice to store the parsed numbers
	var groupIds []int64

	// Parse and append each substring to the slice
	for _, str := range groupIdsStrs {
		num, err := strconv.ParseInt(str, 10, 64)

		if err != nil {
			fmt.Printf("Error parsing number: %v\n", err)
			continue
		}

		groupIds = append(groupIds, num)
	}

	var whitelisted = false

	for _, id := range groupIds {
		if id == uid.int64() {
			whitelisted = true
		}
	}

	return whitelisted
}

/**
 * Description
 */

type ContractDescription string

func (d ContractDescription) has(keywords []string) bool {
	// Iterate over the list of words
	for _, word := range keywords {
		word = strings.ToUpper(word)

		// Check if the word is contained in the description
		if strings.Contains(d.String(), word) {
			return true
		}
	}

	return false
}

func (d ContractDescription) String() string {
	return string(d)
}

/**
 * RIF
 */

type ContractRIF string

func (rif ContractRIF) includes(acceptedRIFs []string) bool {
	for _, accepted := range acceptedRIFs {
		if rif.String() == accepted {
			return true
		}
	}
	return false
}

func (rif ContractRIF) String() string {
	return string(rif)
}

/**
 * Contract
 */

type Contract struct {
	RIF         ContractRIF
	Name        string
	ID          string
	Date        string
	Status      string
	Type        string
	Description ContractDescription
	State       string
}

func (c Contract) HumanReadable() string {
	return fmt.Sprintf("Date: %s\nID: %s\nName: %s\nRIF: %s\nStatus: %s\nType: %s\nState: %s\nDescription: %s\n", c.Date, c.ID, c.Name, c.RIF, c.Status, c.Type, c.State, c.Description)
}
