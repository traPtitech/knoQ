package utils

import "github.com/cheekybits/genny/generic"

//go:generate genny -in=$GOFILE -out=gen-$GOFILE gen "Type=uuid.UUID"

type Type generic.Type

func TypeIn(element Type, set []Type) bool {
	for _, v := range set {
		if element == v {
			// exist!
			return true
		}
	}
	return false
}
