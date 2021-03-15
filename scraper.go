package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
	log "github.com/sirupsen/logrus"
)

type Product struct {
	ProductName string `json:"product_name"`
	ProductURL string `json:"product_url"`
	ProductImageUrl string `json:"product_image_url"`
	InStorePrice string `json:"in_store_price"`
	OriginalPrice string `json:"original_price"`
	StoreName string `json:"store_name"`
	PhoneNumber string `json:"phone_number"`
	Address string `json:"address"`
	StoreCoordinates []float32 `json:"store_coordinates"`
	ClosingHour string `json:"closing_hour"`
}

func RunScraper() []Product{
	log.Info("Running Scraper Module")

	var allProducts []Product

	aceProducts := scrapeAce();

	allProducts = append(allProducts, aceProducts...)

	return allProducts
}


func scrapeAce() []Product{

	aceProducts := []Product{}

	// Constants
	storeName := "Ace Hardware"
	phoneNumber := "(925) 705-7500"
	address := "2044 Mt Diablo Blvd Walnut Creek, CA 94596"
	storeCoordinates := []float32{37.89759212473446, -122.06905777332277}
	closingHour := "19:00"

	departments := []string{"tools"}
	brands := []string{"dewalt", "craftsman", "milwaukee"}

	// TODO - Figure out best way to do this!
	maxPageSize := 300

	aceUrl := fmt.Sprintf("https://www.acehardware.com/departments/%v/%v?pageSize=%v", departments[0], brands[0], maxPageSize)

	// create context
	log.Info("Initializing Context for Ace Scraper")
	ctx, cancel := chromedp.NewContext(context.Background())
	ctx, cancel = context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Navigate to page
	if err := chromedp.Run(ctx, chromedp.Navigate(aceUrl)); err != nil {
		panic("Error navigating to Ace url")
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
		panic("Error getting product data from ace store")
	}
	fmt.Printf("Found %v products\n", len(productItems))

	// Create Product Objects and append to main list
	for i := 1; i < len(productItems); i++ {
		if len(productItems[i].Children[2].Children[3].Children) == 0 {
			continue
		}

		productImage := fmt.Sprintf("https:%v", strings.TrimSpace(productItems[i].Children[1].Children[0].Children[0].AttributeValue("src")))
		productURL := strings.TrimSpace(productItems[i].Children[2].Children[0].AttributeValue("href")) 
		productName := strings.TrimSpace(productItems[i].Children[2].Children[0].Children[0].NodeValue)
		productPrice:= strings.TrimSpace(productItems[i].Children[2].Children[3].Children[1].Children[0].NodeValue)
		

		fmt.Printf("Found product title #%v: %v\n", i, productName)
		fmt.Printf("Found product image: #%v, %v\n", i, productImage)
		fmt.Printf("Found product url #%v: %v\n", i, productURL)
		fmt.Printf("Found product price #%v: %v\n", i, productPrice)
		fmt.Println("---------------------------------------------")

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

		aceProducts = append(aceProducts, newProduct)
	}

	return aceProducts
}

func scrapeCvs() {
	log.Info("Starting to scrape CVS")
}


// // Scroll to bottom
	// var res *runtime.RemoteObject
	// if err := chromedp.Run(ctx, chromedp.Evaluate(`window.scrollTo(0,document.body.scrollHeight);`, &res)); err != nil {
	// 	panic(err)
	// }