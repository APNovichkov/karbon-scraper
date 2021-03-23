package main

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

func ScrapeSafeway(productsChan chan Product) {
	// Constants
	storeName := "Safeway"
	phoneNumber := "(925) 935-9205"
	address := "600 S Broadway, Walnut Creek, CA 94596"
	storeCoordinates := []float32{-122.05669815449866, 37.898100097265385}
	closingHour := "23:00"

	dataPoints := map[string][]string {
		"paper-cleaning-home": {"cleaners-supplies", "hardware-home-needs", "laundry", "kitchen-accessories"},
		"personal-care-health": {"deodorants", "shave-needs", "first-aid-medical-care"},
	}

	localWg := sync.WaitGroup{}

	for categoryy, subcategories := range dataPoints {
		log.Info("Searching CVS category: ", categoryy)

		for _, searchKeyy := range subcategories {
			log.Info("Searching CVS category: ", searchKeyy)
			maxPageNum := 15

			localWg.Add(1)
			go func(category string, searchKey string, productsChan chan Product){
				defer localWg.Done()

				// Click load more until there is no more load more
				for i := 0; i < maxPageNum; i++ {
					// Initialize Context
					log.Info("Initializing Context for CVS")
					ctx, cancel := chromedp.NewContext(context.Background())
					defer cancel()

					ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
					defer cancel()
		
					safewayUrl := fmt.Sprintf("https://www.safeway.com/shop/aisles/%v/%v.3132.html?sort=&page=%v", category, searchKey, i+1)	
					log.Info(fmt.Sprintf("Set url to: %v", safewayUrl))
					
					// Navigate to url
					if err := chromedp.Run(ctx, chromedp.Navigate(safewayUrl)); err != nil {
						log.Error(fmt.Sprintf("Could not navigate to url: %v", safewayUrl))
					}
					
					log.Info("Looking for product data")
				
					// Look for product data
					var productTitles []*cdp.Node
					var productPrices []*cdp.Node
					var productCards []*cdp.Node
					var priceTmpOne []*cdp.Node
					
					if err := chromedp.Run(ctx, chromedp.Nodes(".product-title", &productTitles),
												chromedp.Nodes(".product-price", &productPrices),
												chromedp.Nodes(".product-price-con", &priceTmpOne),
												chromedp.Nodes(".product-item-inner .container", &productCards)); err != nil {
													log.Warn("Timed out, continuing onto the next page")
													return 
												}
					log.Info(fmt.Sprintf("Found %v Product Titles", len(productTitles)))
					log.Info(fmt.Sprintf("Found %v Product Prices", len(productPrices)))
					log.Info(fmt.Sprintf("Found %v Product Cards", len(productCards)))

					
					priceIndex := 0
					for i := 0; i < len(productCards); i++ {
						// Check if product card has a price
						// fmt.Printf("Card price child at card #%v -> %v\n", i+1, len(productCards[i].Children[0].Children)) 

						if len(productCards[i].Children[0].Children) == 0{
							fmt.Printf("%v does not have a price (out of stock!)\n", productTitles[i].Children[0].NodeValue )
							continue
						}

						productName := strings.TrimSpace(productTitles[i].Children[0].NodeValue)
						productURL := strings.TrimSpace(productTitles[i].AttributeValue("href"))
						productPriceId := strings.TrimSpace(productPrices[priceIndex].AttributeValue("id"))

						var productPrice string
						if err := chromedp.Run(ctx, chromedp.Text(fmt.Sprintf("#%v", productPriceId), &productPrice)); err != nil {
							panic("Could not find price for product")
						}
						
						newProduct := Product{
							StoreName: storeName,
							ClosingHour: closingHour,
							PhoneNumber: phoneNumber,
							Address: address,
							StoreCoordinates: storeCoordinates,
							ProductName: productName,
							ProductURL: productURL,
							ProductImageUrl: "productImageHere",
							InStorePrice: formatSafewayPrice(productPrice),
							OriginalPrice: formatSafewayPrice(productPrice),
						}
			
						productsChan <- newProduct

						priceIndex ++
					}

				}
			
			}(categoryy, searchKeyy, productsChan)
		}		
	}

	localWg.Wait()

	log.Info("Done scraping Safeway website")

	close(productsChan)
}


func formatSafewayPrice(price string) float64 {
	if strings.Contains(price, "$") {
		splitPrice := strings.Split(price, "\n")[1]
		splitPrice = strings.Split(splitPrice, " ")[0]
		floatPrice, err := strconv.ParseFloat(strings.TrimSpace(splitPrice)[1:], 64)
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