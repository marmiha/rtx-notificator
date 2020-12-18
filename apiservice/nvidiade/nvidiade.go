package nvidiade

import (
	"encoding/json"
	"net/http"
	"rtx-notificator/apiservice"
)

const Name = "nvidia.com/de-de"
const apiUrl = "https://api.nvidia.partners/edge/product/search?page=1&limit=9&locale=de-de"

// These are nvidia.com/de-de specific IDs for these graphics cards.
var TranslatorTable = apiservice.GpuLookupTable{
	apiservice.Rtx3080:   30044,
	apiservice.Rtx3070:   30052,
	apiservice.Rtx3060Ti: 33824,
}

type StockClient struct {
	httpClient http.Client
}

func (s StockClient) Name() string {
	return Name
}

func NewStockClient() *StockClient {
	client := StockClient{
		httpClient: http.Client{},
	}
	return &client
}

func (s StockClient) CheckStock(gpus ...apiservice.Gpu) ([]apiservice.StockCheckResult, error) {
	var results []apiservice.StockCheckResult
	var stockResp StockResponse

	// Call the endpoint.
	resp, err := http.Get(apiUrl)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Decode the body.
	err = json.NewDecoder(resp.Body).Decode(&stockResp)
	if err != nil {
		return nil, err
	}

	for _, gpu := range gpus {
		checkResult := apiservice.StockCheckResult{
			Gpu:   gpu,
		}

		for _, detail := range stockResp.SearchedProducts.ProductDetails {
			if detail.ProductId == TranslatorTable[gpu] {
				for _, retailer := range detail.Retailers {

					// For each retailer
					r := apiservice.RetailerResult{
						Stock: retailer.Stock,
						Name: retailer.Name,
						Link:  retailer.DirectPurchaseLink,
						Msg: apiservice.StockUnavailable,
						Alert: false,
					}

					// Alert only if available
					if retailer.Stock > 0 {
						r.Alert = true
						r.Msg = apiservice.StockAvailable
					}

					// Append
					checkResult.Retailers = append(checkResult.Retailers, r)
				}
			}
		}

		results = append(results, checkResult)
	}

	return results, nil
}
