package main

import "strings"

const (
	mapReturnType           = "map"
	sliceReturnType         = "slice"
	atPlaceholderType       = "@"
	questionPlaceholderType = "?"
	dollarPlaceholderType   = "$"
)

type GenFunction struct {
	Name             string
	Args             string
	ReturnValueItems string
}

type StmtItem struct {
	Stmt     string
	Function GenFunction
}

type TemplateData struct {
	StmtItems       map[string]StmtItem
	ReturnValueType string
	GenPackage      string
	ImportPackages  []string
}

type GenFuncReturnType struct {
	Signature string
	Type      string
}

func (rt *GenFuncReturnType) IsMap() bool {
	return rt.Type == "map"
}

func (rt *GenFuncReturnType) IsSlice() bool {
	return rt.Type == "slice"
}

func NewGenFuncReturnType(inp string) *GenFuncReturnType {
	formattedInput := strings.ToLower(strings.TrimSpace(inp))
	if formattedInput == "map" {
		return &GenFuncReturnType{
			Signature: "map[string]any",
			Type:      "map",
		}
	} else {
		return &GenFuncReturnType{
			Signature: "[]any",
			Type:      "slice",
		}
	}
}
