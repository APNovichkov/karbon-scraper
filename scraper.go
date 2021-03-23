package main

import (
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

type Product struct {
	ProductName string `json:"product_name"`
	ProductURL string `json:"product_url"`
	ProductImageUrl string `json:"product_image_url"`
	InStorePrice float64 `json:"in_store_price"`
	OriginalPrice float64 `json:"original_price"`
	StoreName string `json:"store_name"`
	PhoneNumber string `json:"phone_number"`
	Address string `json:"address"`
	StoreCoordinates []float32 `json:"store_coordinates"`
	ClosingHour string `json:"closing_hour"`
}

func RunScraper() []Product{
	log.Info("Running Scraper Module")

	allProducts := []Product{}
	productsChan := make(chan Product, 1000000)
	// defer close(productsChan)
	
	// Run Ace Scraper
	ScrapeAce(productsChan);
	ScrapeCvs(productsChan);
	ScrapeSafeway(productsChan);

	// Read in items from channel
	for i := 0; i < 1000000; i++ {
        select {
        case product, ok := <- productsChan:
            if !ok {
                productsChan = nil
            }
            allProducts = append(allProducts, product)
		}
        if productsChan == nil {
			break
        }
    }

	return allProducts
}


func formatPrice(price string) float64{
	if strings.Contains(price, "$") {
		floatPrice, err := strconv.ParseFloat(price[1:], 64)
		if err != nil {
			panic(err)
		}
		return float64(int(floatPrice * 100)) / 100 
	}

	floatPrice, err := strconv.ParseFloat(price, 32)
	if err != nil {
		panic(err)
	}
	
	return float64(int(floatPrice * 100)) / 100
}