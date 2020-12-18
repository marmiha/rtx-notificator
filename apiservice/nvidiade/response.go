package nvidiade

type Retailer struct {
	Name string `json:"retailerName"`
	ProductId int `json:"productId"`
	DirectPurchaseLink string `json:"directPurchaseLink"`
	Stock int `json:"stock"`
}

type ProductDetail struct {
	ProductId int `json:"productID"`
	DisplayName string `json:"displayName"`
	Retailers []Retailer `json:"retailers"`
}

type SearchedProducts struct {
	ProductDetails []ProductDetail `json:"productDetails"`
}

type StockResponse struct {
	SearchedProducts SearchedProducts `json:"searchedProducts"`
}