package main

type CSVPayload struct {
	Filename  string `json:"fileName"`
	InputDir  string `json:"inputDir"`
	OutputDir string `json:"outputDir"`
}
