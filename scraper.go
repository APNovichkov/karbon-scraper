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

	wg := sync.WaitGroup{}
	allProducts := []Product{}
	productsChan := make(chan Product, 1000000)

	// Run Ace Scraper
	wg.Add(1)
	go scrapeAce(&wg, productsChan);
	wg.Add(1)
	go scrapeCvs(&wg, productsChan);

	// Defer closing of channel until waitgroup is clear
	go func() {
		defer close(productsChan)
        wg.Wait()
	}()

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


func scrapeAce(wg *sync.WaitGroup, productsChan chan Product) {
	defer wg.Done()

	// Constants
	storeName := "Ace Hardware"
	phoneNumber := "(925) 705-7500"
	address := "2044 Mt Diablo Blvd Walnut Creek, CA 94596"
	storeCoordinates := []float32{37.89759212473446, -122.06905777332277}
	closingHour := "19:00"

	dataPoints := map[string][]string{
		"tools": {"dewalt", "craftsman", "milwaukee"},
		"lighting-and-electrical": {"light-bulbs", "batteries", "circuit-breakers-fuses-and-panels", "electrical-tools", "wire", "home-electronics"},
	}
	for category, subcategories := range dataPoints{

		log.Info("Looking at ACE category: ", category)

		for _, searchKey := range subcategories {
			
			log.Info("Looking at ACE subcategory: ", searchKey)

			// TODO - Figure out best way to do this!
			maxPageSize := 300
	
			aceUrl := fmt.Sprintf("https://www.acehardware.com/departments/%v/%v?pageSize=%v", category, searchKey, maxPageSize)
	
			// create context
			log.Info(fmt.Sprintf("Initializing Context for Ace Scraper: %v", aceUrl))
			ctx, cancel := chromedp.NewContext(context.Background())
			defer cancel()
			ctx, cancel = context.WithTimeout(ctx, 1*time.Hour)
			defer cancel()
	
			// Navigate to page
			if err := chromedp.Run(ctx, chromedp.Navigate(aceUrl)); err != nil {
				panic("Error navigating to ACE url")
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
				continue
			}
			fmt.Printf("Found %v products in that ACE url\n", len(productItems))
	
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
					ProductName: productName,
					ProductURL: productURL,
					ProductImageUrl: productImage,
					InStorePrice: productPrice,
					OriginalPrice: productPrice,
				}
	
				productsChan <- newProduct
			}
		}
	}
	
	log.Info("Finished Scraping ACE")
}

func scrapeCvs(wg *sync.WaitGroup, productsChan chan Product) {
	defer wg.Done()

	// Constants
	storeName := "CVS Pharmacy"
	phoneNumber := "(925) 933-8353"
	address := "1123 S California Blvd, Walnut Creek, CA 94596"
	storeCoordinates := []float32{37.896474574278315, -122.06446582963069}
	closingHour := "Open 24 hours"

	dataPoints := map[string][]string {
		"household": {"batteries-electronics", "hardware", "paper-plastic", "school-office-supplies"},
		"personal-care": {"shaving", "deodrant"},
	}

	for category, subcategories := range dataPoints {
		log.Info("Searching CVS category: ", category)

		for _, searchKey := range subcategories {
			log.Info("Searching CVS category: ", searchKey)
			maxPageNum := 30
	
			for i := 0; i < maxPageNum; i++ {
				// Initialize Context
				log.Info("Initializing Context for CVS")
				ctx, cancel := chromedp.NewContext(context.Background())
				defer cancel()

				ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
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
		}		
	}
	
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