package features

import (
	"unicode"
)

const (
	// prefixes
	CHARACTER   = "Character"
	CHAR_OFFSET = "CharOffset"
	TOKEN       = "Token"

	TOKEN_PREV_IDENTITY = "TokenPrevIdentity"
	TOKEN_NEXT_IDENTITY = "TokenNextIdentity"
	TOKEN_PREV_LEN      = "TokenPrevLength"
	TOKEN_NEXT_LEN      = "TokenNextLength"

	TOKEN_CAPITALIZED = "Tokencap"

	TOKEN_CONTEXT_CAT   = "TokenContextCat"
	LEFT_WORD_RIGHT_CAP = "LeftWordRightCap"

	RIGHT_LOWER  = "RightLower"
	LEFT_DOTLESS = "LeftDotless"

	PREV_OUTCOME = "PrevOutcome"

	SUFFIX_ID    = "Id"
	SUFFIX_UPPER = "Upper"
	SUFFIX_LOWER = "Lower"
	SUFFIX_DIGIT = "Digit"
	SUFFIX_SPACE = "Space"

	SUFFIX_FALSE = "false"
	SUFFIX_TRUE  = "true"
)

type charRangeType struct {
	charRange  *unicode.RangeTable
	typeString string
}

var charRangesNames = []charRangeType{
	{unicode.Ll, "Type2"},
	{unicode.Lu, "Type1"},
	{unicode.Nd, "Type9"},
	{unicode.Po, "Type24"},
	{unicode.Zs, "Type12"},
	{unicode.Cc, "Type15"},
	{unicode.Pd, "Type20"},
	{unicode.Pe, "Type22"},
	{unicode.Ps, "Type21"},
	{unicode.Sm, "Type25"},
	{unicode.Pc, "Type23"},
	{unicode.Lt, "Type3"},
	{unicode.Zl, "Type13"},
	{unicode.Mc, "Type8"},
	{unicode.Me, "Type7"},
	{unicode.Nl, "Type10"},
	{unicode.Sc, "Type26"},
	{unicode.Pf, "Type30"},
	{unicode.Cf, "Type16"},
	{unicode.Pi, "Type29"},
	{unicode.Lm, "Type4"},
	{unicode.Sk, "Type27"},
	{unicode.Mn, "Type6"},
	{unicode.Lo, "Type5"},
	{unicode.No, "Type11"},
	{unicode.So, "Type28"},
	{unicode.Zp, "Type14"},
	{unicode.Co, "Type18"},
	{unicode.Cs, "Type19"},
}

func getCharTypeString(curChar rune) string {
	for _, charRange := range charRangesNames {
		if unicode.Is(charRange.charRange, curChar) {
			return charRange.typeString
		}
	}
	return "Type0"
}

var CharOffsetFeatureValues = map[int]string{
	-3: "-3",
	-2: "-2",
	-1: "-1",
	0:  "0",
	1:  "1",
	2:  "2",
	3:  "3",
}
