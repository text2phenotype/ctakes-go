package features

import (
	"text2phenotype.com/fdl/nlp/model"
	"unicode"
)

type CharFeaturesBuilder struct {
	*Builder
	cache map[rune]*charProperties
}

type charBoolProperties byte

type charProperties struct {
	boolProperties charBoolProperties
	typeStr        string
	id             string
}

var charChecks = [4]struct {
	featureName string
	checkFunc   func(rune) bool
}{
	{SUFFIX_UPPER, unicode.IsUpper},
	{SUFFIX_LOWER, unicode.IsLower},
	{SUFFIX_DIGIT, unicode.IsDigit},
	{SUFFIX_SPACE, unicode.IsSpace},
}

func NewCharFeaturesBuilder(model *model.Model) *CharFeaturesBuilder {
	return &CharFeaturesBuilder{
		Builder: NewFeaturesBuilder(model),
		cache:   map[rune]*charProperties{},
	}
}

func (builder *CharFeaturesBuilder) AppendCharFeatures(char rune, featureNamePrefixParts ...string) {
	builder.setFeaturePrefix(featureNamePrefixParts...)
	builder.appendCharFeatures(char)
	builder.removePrefixes()
}

func (builder *CharFeaturesBuilder) appendCharFeatures(char rune) {
	cacheItem, hasCache := builder.cache[char]
	if !hasCache {
		cacheItem = getCharProperties(char)
		builder.cache[char] = cacheItem
	}
	builder.appendFromBoolProperties(cacheItem.boolProperties)
	builder.AppendFeature(cacheItem.typeStr, SUFFIX_TRUE)
	builder.AppendFeature(SUFFIX_ID, cacheItem.id)
}

func getCharProperties(char rune) *charProperties {
	var charId string
	if char == '\n' {
		charId = "<LF>"
	} else {
		charId = string(char)
	}
	typeStr := getCharTypeString(char)
	return &charProperties{
		boolProperties: getCharBoolProperties(char),
		typeStr:        typeStr,
		id:             charId,
	}
}

func (builder *CharFeaturesBuilder) appendFromBoolProperties(boolProperties charBoolProperties) {
	properties := boolProperties
	for i := len(charChecks) - 1; i >= 0; i-- {
		check := charChecks[i]
		if properties&1 > 0 {
			builder.AppendFeature(check.featureName, SUFFIX_TRUE)
		} else {
			builder.AppendFeature(check.featureName, SUFFIX_FALSE)
		}
		properties = properties >> 1
	}
}

func getCharBoolProperties(char rune) charBoolProperties {
	var boolProperties charBoolProperties
	for _, check := range charChecks {
		if check.checkFunc(char) {
			boolProperties = boolProperties<<1 + 1
		} else {
			boolProperties = boolProperties << 1
		}
	}
	return boolProperties
}
