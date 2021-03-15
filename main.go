package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"
)

func main() {
	log.Info("Running Karbon Scraper")

	allProducts := RunScraper()

	log.Info("Converting Products to JSON")
	fmt.Println("------------------------------------------------")
	jobListingsString, _ := json.Marshal(allProducts)
	filename := "products.json"
	ioutil.WriteFile(filename, jobListingsString, os.ModePerm)

	log.Info("Done running Karbon Scraper")
}