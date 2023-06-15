package main

type CSVPayload struct {
	Filename  string `json:"fileName"`
	InputDir  string `json:"inputDir"`
	OutputDir string `json:"outputDir"`
}

type StructInput struct {
	Index            string
	CustomerId       string
	FirstName        string
	LastName         string
	Company          string
	City             string
	Country          string
	Phone1           string
	Phone2           string
	Email            string
	SubscriptionDate string
	Website          string
}
