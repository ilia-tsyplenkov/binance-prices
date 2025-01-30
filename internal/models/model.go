package models

type Price struct {
	Symbol string `json:"symbol"`
	Ask    string `json:"ask"`
	Bid    string `json:"bid"`
}
