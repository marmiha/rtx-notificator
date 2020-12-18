// Package includes the interface which should be implemented for the GPU availability checks.
package apiservice

import "fmt"

// Lookup table defined for each api.
type GpuLookupTable map[Gpu]int

// Constants which represent the GPUs that the stocks can be polled.
type Gpu string

const (
	Rtx3090   Gpu = "3090"
	Rtx3080   Gpu = "3080"
	Rtx3070       = "3070"
	Rtx3060Ti     = "3060Ti"
)

type StockMsg string

const (
	StockAvailable   StockMsg = "Stock Available"
	StockUnavailable StockMsg = "Stock Unavailable"
)

type GpuStockClient interface {
	CheckStock(gpus ...Gpu) ([]StockCheckResult, error)
	Name() string
}

type StockCheckResult struct {
	Gpu       Gpu
	Retailers []RetailerResult
}

// If the program should alert the user via msg service about this result.
func (s StockCheckResult) ShouldAlert() bool {
	for _, r := range s.Retailers {
		if r.Alert {
			return true
		}
	}
	return false
}

func (r RetailerResult) String() string {
	return fmt.Sprintf("Retailer: %v Stock: %v Link: %v", r.Name, r.Stock, r.Link)
}

// The alert string msg that will be send via the msg service.
func (s StockCheckResult) AlertString() string {
	res := fmt.Sprintf("[%v] ", s.Gpu)
	for _, r := range s.Retailers {
		if r.Alert {
			res += fmt.Sprint("\n||", r, "||\n")
		}
	}
	return res
}

type RetailerResult struct {
	Stock int
	Name  string
	Msg   StockMsg
	Link  string
	Alert bool
}
