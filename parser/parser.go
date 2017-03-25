package parser

import "github.com/bfontaine/quinoa/language"

func Parse(code string) (*language.AST, error) {

	p := &Parser{}
	p.Init()

	err := p.Parse()

	// dummy parsing
	return &language.AST{}, err
}
