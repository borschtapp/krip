package model

import (
	"github.com/borschtapp/krip/utils"
)

type Ingredient struct {
	Quantity    float64 `json:"quantity,omitempty"`
	QuantityMax float64 `json:"quantityMax,omitempty"`
	Unit        string  `json:"unit,omitempty"`
	Ingredient  string  `json:"ingredient,omitempty"`
	Description string  `json:"description,omitempty"`
	Annotation  string  `json:"annotation,omitempty"`
	Optional    bool    `json:"optional,omitempty"`
	Grams       float64 `json:"grams,omitempty"` // unified unit to convert all ingredients
}

func (r *Ingredient) String() (s string) {
	s += utils.FormatFraction(r.Quantity)
	if r.QuantityMax > 0 {
		s += "-" + utils.FormatFraction(r.QuantityMax)
	}

	if r.Unit == "" {
		s += " <unit>"
	} else {
		s += " " + r.Unit
	}

	s += " " + r.Ingredient

	if len(r.Annotation) > 0 {
		s += " (" + r.Annotation + ")"
	}

	if len(r.Description) > 0 {
		s += ", " + r.Description
	}

	if r.Optional == true {
		s += ", optional"
	}
	return s
}
