package main

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

func ScrapeAce(productsChan chan Product) {
	// Constants
	storeName := "Ace Hardware"
	phoneNumber := "(925) 705-7500"
	address := "2044 Mt Diablo Blvd Walnut Creek, CA 94596"
	storeCoordinates := []float32{-122.06905777332277, 37.89759212473446}
	storeCoordinatesString := []float32{37.89759212473446, -122.06905777332277}
	closingHour := "19:00"

	localWg := sync.WaitGroup{}

	dataPoints := map[string][]string{
		"tools": {"dewalt", "craftsman", "milwaukee"},
		"lighting-and-electrical": {"light-bulbs", "batteries", "circuit-breakers-fuses-and-panels", "electrical-tools", "wire", "home-electronics"},
	}
	for categoryy, subcategories := range dataPoints{

		log.Info("Looking at ACE category: ", categoryy)

		for _, searchKeyy := range subcategories {
			
			log.Info("Looking at ACE subcategory: ", searchKeyy)
			localWg.Add(1)

			// TODO - look this up
			go func(searchKey string, category string, productsChan chan Product) {
				defer localWg.Done()
				
				// TODO - Figure out best way to do this!
				maxPageSize := 300
		
				aceUrl := fmt.Sprintf("https://www.acehardware.com/departments/%v/%v?pageSize=%v", category, searchKey, maxPageSize)
		
				// create context
				log.Info(fmt.Sprintf("Initializing Context for Ace Scraper: %v", aceUrl))
				ctx, cancel := chromedp.NewContext(context.Background())
				defer cancel()
				ctx, cancel = context.WithTimeout(ctx, 30*time.Second)
				defer cancel()
		
				// Navigate to page
				if err := chromedp.Run(ctx, chromedp.Navigate(aceUrl)); err != nil {
					log.Error(fmt.Sprintf("Error navigating to ACE url: %v"), aceUrl)
					return
				}
		
				// Look for product data
				var productTitles []*cdp.Node
				var productPrices []*cdp.Node
				var productImages []*cdp.Node
				var productItems []*cdp.Node
				if err := chromedp.Run(ctx, chromedp.Nodes(".mz-productlisting-title", &productTitles), 
											chromedp.Nodes(".sales-price, .custom-price", &productPrices), 
											chromedp.Nodes("a > img", &productImages), 
											chromedp.Nodes(".mz-productlisting", &productItems)); err != nil {
					log.Warn("Error getting product data from ace store, skipping current URL")
					return
				}
				fmt.Printf("Found %v products at url %v\n", len(productItems), aceUrl)
		
				// Create Product Objects and append to main list
				for i := 1; i < len(productItems); i++ {
					if len(productItems[i].Children[2].Children[3].Children) == 0 {
						continue
					}
		
					productImage := fmt.Sprintf("https:%v", strings.TrimSpace(productItems[i].Children[1].Children[0].Children[0].AttributeValue("src")))
					productURL := strings.TrimSpace(productItems[i].Children[2].Children[0].AttributeValue("href")) 
					productName := strings.TrimSpace(productItems[i].Children[2].Children[0].Children[0].NodeValue)
					productPrice:= formatPrice(strings.TrimSpace(productItems[i].Children[2].Children[3].Children[1].Children[0].NodeValue))
					
		
					// fmt.Printf("Found product title #%v: %v\n", i, productName)
					// fmt.Printf("Found product image: #%v, %v\n", i, productImage)
					// fmt.Printf("Found product url #%v: %v\n", i, productURL)
					// fmt.Printf("Found product price #%v: %v\n", i, productPrice)
					// fmt.Println("---------------------------------------------")
		
					newProduct := Product{
						StoreName: storeName,
						ClosingHour: closingHour,
						PhoneNumber: phoneNumber,
						Address: address,
						StoreCoordinates: storeCoordinates,
						StoreCoordinatesString: storeCoordinatesString,
						ProductName: productName,
						ProductURL: productURL,
						ProductImageUrl: productImage,
						InStorePrice: productPrice,
						OriginalPrice: productPrice,
					}
		
					productsChan <- newProduct
				}
			}(searchKeyy, categoryy, productsChan)
		}
	}
	localWg.Wait()
	
	log.Info("Finished Scraping ACE")

	// close(productsChan)
}