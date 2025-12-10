package ttml

import (
	"errors"
	"strings"

	"github.com/beevik/etree"
)

func ExtractTextFromTTML(ttml string) (string, error) {
	document := etree.NewDocument()
	err := document.ReadFromString(ttml)
	if err != nil {
		return ``, err
	}
	element := document.FindElement("/tt/body")
	if element == nil {
		return ``, errors.New("invalid ttml")
	}

	result := strings.Builder{}
	for _, div := range element.ChildElements() {
		if div.Tag != "div" {
			return ``, errors.New("invalid ttml: unknown tag in body")
		}
		for _, p := range div.ChildElements() {
			if p.Tag != "p" {
				return ``, errors.New("invalid ttml: unknown tag in body > div")
			}
			result.WriteString(p.Text())
			result.WriteByte('\n')
		}
		result.WriteByte('\n')
	}

	return strings.TrimRight(result.String(), "\n"), nil
}
