package main

type CSVMergePayload struct {
	Filename  string `json:"fileName"`
	InputDir  string `json:"inputDir"`
	OutputDir string `json:"outputDir"`
}

type CSVSplitPayload struct {
	Filename   string `json:"fileName"`
	PerRow     int    `json:"perRow"`
	InputFile  string `json:"inputFile"`
	OutputFile string `json:"OutputFile"`
}
