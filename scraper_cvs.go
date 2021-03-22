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

func ScrapeCvs(wg *sync.WaitGroup, productsChan chan Product) {
	defer wg.Done()

	// Constants
	storeName := "CVS Pharmacy"
	phoneNumber := "(925) 933-8353"
	address := "1123 S California Blvd, Walnut Creek, CA 94596"
	storeCoordinates := []float32{-122.06446582963069, 37.896474574278315}
	closingHour := "Open 24 hours"

	dataPoints := map[string][]string {
		"household": {"batteries-electronics", "hardware", "paper-plastic", "school-office-supplies"},
		"personal-care": {"shaving", "deodrant"},
	}

	localWg := sync.WaitGroup{}

	for categoryy, subcategories := range dataPoints {
		log.Info("Searching CVS category: ", categoryy)

		for _, searchKeyy := range subcategories {
			log.Info("Searching CVS category: ", searchKeyy)
			maxPageNum := 10

			localWg.Add(1)
			go func(category string, searchKey string, productsChan chan Product){
				defer localWg.Done()
				
				for i := 0; i < maxPageNum; i++ {

					// Initialize Context
					log.Info("Initializing Context for CVS")
					ctx, cancel := chromedp.NewContext(context.Background())
					defer cancel()

					ctx, cancel = context.WithTimeout(ctx, 20*time.Second)
					defer cancel()
		
					cvsUrl := fmt.Sprintf("https://www.cvs.com/shop/%v/%v?page=%v", category, searchKey, i+1)	
					log.Info(fmt.Sprintf("Set url for page #%v: %v", i+1, cvsUrl))
			
					
					log.Info("Looking for product data")
					
					// Look for product data
					var productTitles []*cdp.Node
					var productPrices []*cdp.Node
					var productUrls []*cdp.Node
					
					if err := chromedp.Run(ctx, chromedp.Navigate(cvsUrl),
												chromedp.Sleep(3*time.Second),
												chromedp.Nodes(".r-ubezar", &productTitles),
												chromedp.Nodes(".r-ttdzmv", &productPrices)); err != nil {
													log.Warn("Timed out, continuing onto the next page")
													continue 
												}
					
					if err := chromedp.Run(ctx, chromedp.Nodes(".r-1lz4bg0", &productUrls)); err != nil {
						log.Warn("Was not able to find any URL's for this page")
					}							

					log.Info(fmt.Sprintf("Found %v Product Titles", len(productTitles)))
					log.Info(fmt.Sprintf("Found %v Product Prices", len(productPrices)))
					log.Info(fmt.Sprintf("Found %v Product Urls", len(productUrls)))
			
					// Have to start at 2 because of inconsistency in the loaded page
					for i := 2; i < len(productTitles); i++ {
						productName := strings.TrimSpace(productTitles[i].Children[0].NodeValue)
						productPrice := formatPrice(strings.TrimSpace(productPrices[i-1].Children[0].NodeValue))

						productURL := "NA"
						
						if len(productUrls) > 0 {
							productURL = fmt.Sprintf("https://www.cvs.com%v", strings.TrimSpace(productUrls[i-1].AttributeValue("href")))
						}
						
			
						// fmt.Printf("Product Name #%v: %v\n", i, productName)
						// fmt.Printf("Product Price #%v: %v\n", i, productPrice)
						// fmt.Printf("Product Url #%v: %v\n", i, productURL)
						// fmt.Println("------------------------------------")
			
						newProduct := Product{
							StoreName: storeName,
							ClosingHour: closingHour,
							PhoneNumber: phoneNumber,
							Address: address,
							StoreCoordinates: storeCoordinates,
							ProductName: productName,
							ProductURL: productURL,
							ProductImageUrl: "productImageHere",
							InStorePrice: productPrice,
							OriginalPrice: productPrice,
						}
			
						productsChan <- newProduct
					}
				}
			}(categoryy, searchKeyy, productsChan)
		}		
	}

	localWg.Wait()

	log.Info("Done scraping CVS website")
}