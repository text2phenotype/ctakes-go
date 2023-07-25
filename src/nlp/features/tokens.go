package features

import (
	"text2phenotype.com/fdl/nlp/model"
	"strconv"
	"unicode"
)

type TokenFeaturesBuilder = Builder

func NewTokenFeaturesBuilder(m *model.Model) *TokenFeaturesBuilder {
	return NewFeaturesBuilder(m)
}

func (builder *TokenFeaturesBuilder) AppendTokenFeatures(prevToken string, nextToken string) {
	builder.AppendFeature(TOKEN_PREV_IDENTITY, prevToken)

	if len(nextToken) != 0 {
		builder.AppendFeature(TOKEN_NEXT_IDENTITY, nextToken)
	} else {
		builder.AppendFeature(TOKEN_NEXT_IDENTITY)
	}
	builder.AppendFeature(TOKEN_NEXT_LEN, strconv.Itoa(len(nextToken)), SUFFIX_TRUE)
	builder.AppendFeature(TOKEN_PREV_LEN, strconv.Itoa(len(prevToken)), SUFFIX_TRUE)

	nextTokenCap := len(nextToken) > 0 && unicode.IsUpper(rune(nextToken[0]))

	if nextTokenCap {
		builder.AppendFeature(TOKEN_CAPITALIZED, SUFFIX_TRUE)
		builder.AppendFeature(LEFT_WORD_RIGHT_CAP, prevToken, SUFFIX_TRUE)
	} else {
		builder.AppendFeature(TOKEN_CAPITALIZED, SUFFIX_FALSE)
		builder.AppendFeature(LEFT_WORD_RIGHT_CAP, prevToken, SUFFIX_FALSE)
	}
	builder.AppendFeature(TOKEN_CONTEXT_CAT, prevToken, nextToken)
}
