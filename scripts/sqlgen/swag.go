package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type apiEntry struct {
	OperationID string `json:"operationId" yaml:"operationId"`
	Summary     string `json:"summary" yaml:"summary"`
}

type methods map[string]apiEntry

type paths map[string]methods

type swagDoc struct {
	Swagger string `json:"swagger" yaml:"swagger"`
	Info    struct {
		Title string `json:"title" yaml:"title"`
	}
	Paths paths `json:"paths" yaml:"paths"`
}

func loadDoc(docfile string) (*swagDoc, error) {
	yf, err := os.Open(docfile)
	if err != nil {
		return nil, err
	}
	doc := new(swagDoc)
	err = yaml.NewDecoder(yf).Decode(doc)
	if err != nil {
		return nil, err
	}

	return doc, nil
}
