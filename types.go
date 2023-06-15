package main

type CSVPayload struct {
	Filename  string `json:"fileName"`
	InputDir  string `json:"inputDir"`
	OutputDir string `json:"outputDir"`
}

type StructInput struct {
	Index            string `csv:"Index"`
	CustomerId       string `csv:"CustomerId"`
	FirstName        string `csv:"FirstName"`
	LastName         string `csv:"LastName"`
	Company          string `csv:"Company"`
	City             string `csv:"City"`
	Country          string `csv:"Country"`
	Phone1           string `csv:"Phone1"`
	Phone2           string `csv:"Phone2"`
	Email            string `csv:"Email"`
	SubscriptionDate string `csv:"SubscriptionDate"`
	Website          string `csv:"Website"`
}
