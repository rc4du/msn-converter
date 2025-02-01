package main

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
	"strings"

	"github.com/rrc4du/msn-converter/models"
)

const (
	inputDir      = "files/input/"
	inputFileName = "example.xml"
	outputDir     = "files/output/"
)

func openInputFile() (file *os.File, closeFunc func()) {
	f, err := os.Open(inputDir + inputFileName)
	if err != nil {
		panic(fmt.Errorf("error opening file: %w", err))
	}
	return f, func() {
		if err = f.Close(); err != nil {
			panic(fmt.Errorf("error closing input file: %w", err))
		}
	}
}

func openOutputFile(dateTime, receiver string) (file *os.File, closeFunc func()) {
	f, err := os.Create(outputDir + dateTime + "-" + receiver + ".txt")
	if err != nil {
		panic(fmt.Errorf("error creating output file: %w", err))
	}
	return f, func() {
		if err = f.Close(); err != nil {
			panic(fmt.Errorf("error closing file: %w", err))
		}
	}
}

func main() {
	inputFile, closeInputFile := openInputFile()
	defer closeInputFile()
	log := models.Log{}
	if err := xml.NewDecoder(inputFile).Decode(&log); err != nil {
		panic(fmt.Errorf("error decoding xml: %w", err))
	}
	dateTime := strings.Replace(log.Messages[0].Date, "/", "_", -1) + "_" + log.Messages[0].Time
	outputFile, closeOutputFile := openOutputFile(dateTime, log.Messages[0].To.User.FriendlyName)
	defer closeOutputFile()
	tmpl, err := template.ParseFiles("templates/output.txt")
	if err != nil {
		panic(fmt.Errorf("error parsing template: %w", err))
	}
	if err = tmpl.Execute(outputFile, log); err != nil {
		panic(fmt.Errorf("error executing template: %w", err))
	}
}
