package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/gocolly/colly/v2"

	"database/sql"

	"storysource/config"

	_ "github.com/lib/pq"
)

var visitedPages map[string]bool

func createTable() error {
	config := config.GetConfig()
	db, err := sql.Open("postgres", config.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `
	CREATE TABLE IF NOT EXISTS story (
		id SERIAL PRIMARY KEY,
		title TEXT,
		url TEXT,
		description TEXT,
		author TEXT,
		filepath TEXT
	)`
	_, err = db.Exec(query)
	if err != nil {
		return err
	}

	_, err = db.Exec("TRUNCATE story")
	if err != nil {
		return err
	}

	return nil
}

func init() {
	visitedPages = make(map[string]bool)
	err := createTable()
	if err != nil {
		fmt.Println("Error creating table:", err)
	}
}

func saveToDB(title, url, descr, author string) error {
	config := config.GetConfig()
	db, err := sql.Open("postgres", config.DbPath)
	if err != nil {
		return err
	}
	defer db.Close()

	query := `INSERT INTO story (title, url, description, author) VALUES ($1, $2, $3, $4)`
	_, err = db.Exec(query, title, url, descr, author)
	if err != nil {
		return err
	}

	return nil
}

func postQuery(url string, data []byte) (string, error) {
	resp, err := http.Post(url, "application/x-www-form-urlencoded; charset=UTF-8", bytes.NewBuffer(data))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func main() {

	config := config.GetConfig()

	c := colly.NewCollector(
		colly.AllowedDomains(config.AllowedDomain),
	)

	// Find and visit all links
	c.OnHTML("div.navigation div.navigation-in nav.pages a[href]", func(e *colly.HTMLElement) {
		if visitedPages[e.Attr("href")] {
			return
		}
		visitedPages[e.Attr("href")] = true
		c.Visit(e.Attr("href"))
	})

	c.OnHTML("article.post div.post-cont", func(e *colly.HTMLElement) {
		title := e.DOM.Find("h2.post-title").Text()
		url, _ := e.DOM.Find("h2.post-title a[href]").Attr("href")
		descr := e.DOM.Find("p.post-text").Text()
		author := e.DOM.Find("span.post-author").Text()

		// Assuming you have a function to save to the database
		err := saveToDB(title, url, descr, author)
		if err != nil {
			fmt.Println("Error saving to database:", err)
		}
		fmt.Printf("Title: %s, url: %s\ndesc: %s\n", title, url, descr)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	for _, searchURL := range config.SearchUrls {
		c.Visit(searchURL)
	}
}
