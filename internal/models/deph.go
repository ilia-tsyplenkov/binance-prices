package models

type DephMessage struct {
	Stream string `json:"stream"`
	Data   struct {
		LastUpdateID int64      `json:"lastUpdateId"`
		Bids         [][]string `json:"bids"`
		Asks         [][]string `json:"asks"`
	} `json:"data"`
}
