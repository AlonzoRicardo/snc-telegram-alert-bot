package main

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/gocolly/colly"
)

func scrap() ([]Contract, error) {
	c := colly.NewCollector()

	var fetchError error
	var contracts []Contract
	var contract Contract

	c.OnHTML("table", func(e *colly.HTMLElement) {
		e.ForEach("tr", func(_ int, row *colly.HTMLElement) {
			row.ForEach("td", func(_ int, cell *colly.HTMLElement) {
				switch cell.Index {
				case 0:
					contract.RIF = ContractRIF(strings.ToUpper(strings.TrimSpace(cell.Text)))
				case 1:
					contract.Name = strings.TrimSpace(cell.Text)
				case 2:
					contract.ID = strings.TrimSpace(cell.Text)
				case 3:
					contract.Date = strings.TrimSpace(cell.Text)
				case 4:
					contract.Status = strings.TrimSpace(cell.Text)
				case 5:
					contract.Type = strings.TrimSpace(cell.Text)
				case 6:
					contract.Description = ContractDescription(strings.ToUpper(strings.TrimSpace(cell.Text)))
				case 7:
					contract.State = strings.TrimSpace(cell.Text)
				default:
					return
				}
			})

			contracts = append(contracts, contract)
		})
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		fmt.Println("Something went wrong:", err)

		fetchError = err
	})

	c.Visit("http://sistemaintegrado.snc.gob.ve/index.php/llamadoxterno")

	if fetchError != nil {
		return nil, fetchError
	} else {
		fmt.Printf("Successfully scrapped %v contracts...\n", len(contracts))

		return contracts, nil
	}
}

func ScrapContracts() ([]Contract, error) {
	maxRetries := 1
	retryDelay := 1 * time.Minute
	var contracts []Contract

	for i := 0; i <= maxRetries; i++ {
		scrappedContracts, err := scrap()

		if err == nil {
			contracts = scrappedContracts

			break
		}

		if i < maxRetries {
			fmt.Printf("Retry #%d failed with error: %v. Retrying in %v...\n", i+1, err, retryDelay)
			time.Sleep(retryDelay)
		} else {
			fmt.Printf("Operation failed after %d retries: %v\n", maxRetries, err)

			return nil, err
		}
	}

	// Custom sorting function to parse and compare dates
	sort.Slice(contracts, func(i, j int) bool {
		date1, _ := time.Parse("02/01/2006", contracts[i].Date)
		date2, _ := time.Parse("02/01/2006", contracts[j].Date)

		return date1.After(date2)
	})

	return contracts, nil
}
