package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/gocolly/colly"
	"github.com/mitchellh/mapstructure"
)

type person struct {
	Gender             int     `json:"gender"`
	Job                string  `json:"job"`
	Name               string  `json:"name"`
}

var apiKey string = os.Getenv("TMDB_API_KEY")
var total int
var women int

func main() {
	var wg sync.WaitGroup
	csvFile, _ := os.Open("watched.csv")
	defer csvFile.Close()
	//Skip First Line
	row1, err := bufio.NewReader(csvFile).ReadSlice('\n')
	if err != nil {
		log.Fatal(err)
	}
	_, err = csvFile.Seek(int64(len(row1)), io.SeekStart)
	if err != nil {
		log.Fatal(err)
	}

	reader := csv.NewReader(csvFile)
	for {
		line, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}
		url := line[3]

		wg.Add(1)
		go isWomen(url, &wg)
	}
	wg.Wait()
	fmt.Printf("With Women: %d\n", women)
	fmt.Printf("Total Able To Get: %d\n", total)
	fmt.Printf("Total percentage %.3f\n", float64(women)/float64(total)*100)
}

func isWomen(url string, wg *sync.WaitGroup) {
	hasWomen := false
	defer wg.Done()
	c := colly.NewCollector()
	var id string
	c.OnHTML("body", func(e *colly.HTMLElement) {
		id = e.Attr("data-tmdb-id")
	})
	c.Visit(url)

	tmdbURL := fmt.Sprintf(
		"%s%s%s?api_key=%s",
		"https://api.themoviedb.org/3/movie/",
		id,
		"/credits",
		apiKey,
	)
	res, err := http.Get(tmdbURL)
	if err != nil {
		log.Fatal(err)
	}
	if res.Body != nil {
		defer res.Body.Close()
	}
	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	var cd map[string]interface{}
	jsonErr := json.Unmarshal(body, &cd)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	if cd["crew"] != nil {
		crew := cd["crew"].([]interface{})
		total++
		for _, v := range crew {
			var crew person
			mapstructure.Decode(v, &crew)
			if crew.Job == "Director" {
				if crew.Gender == 1 {
					hasWomen = true
				}
			}
		}
	}
	if hasWomen {
		women++
	}
}
