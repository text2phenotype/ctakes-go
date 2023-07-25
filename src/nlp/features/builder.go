package features

import (
	"text2phenotype.com/fdl/nlp/model"
)

type Builder struct {
	namesParts *featuresNamesParts
	features   [60]model.FeatureValue
	model      *model.Model
	nextIndex  byte
}

func NewFeaturesBuilder(m *model.Model) *Builder {
	feats := Builder{
		features:   [60]model.FeatureValue{},
		model:      m,
		namesParts: &featuresNamesParts{},
	}
	return &feats
}

func (builder *Builder) Merge(other *Builder) {
	for _, featValue := range other.features[:other.nextIndex] {
		builder.appendFeatureAtIndex(featValue.Index)
	}
}

func (builder *Builder) appendFeatureAtIndex(index int) {
	builder.features[builder.nextIndex] = model.FeatureValue{Index: index, Value: 1.0}
	builder.nextIndex++
}

func (builder *Builder) setFeaturePrefix(prefixParts ...string) {
	builder.namesParts.setPrefixes(prefixParts...)
}

func (builder *Builder) removePrefixes() {
	builder.namesParts.clear()
}

func (builder *Builder) AppendFeature(nameParts ...string) {
	builder.namesParts.fill(nameParts...)
	fIdx, ok := builder.model.GetFeatureIndex(builder.namesParts.slice())
	if !ok {
		return
	}
	builder.appendFeatureAtIndex(fIdx)
}

func (builder *Builder) Values() []model.FeatureValue {
	builder.appendFeatureAtIndex(1)
	return builder.features[:builder.nextIndex]
}

func (builder *Builder) Cleanup() {
	builder.nextIndex = 0
	builder.namesParts.clear()
}

type featuresNamesParts struct {
	inner       [6]string
	nextInd     byte
	prefixesLen byte
}

func (namesParts *featuresNamesParts) fill(appendedParts ...string) {
	namesParts.nextInd = namesParts.prefixesLen
	for _, part := range appendedParts {
		namesParts.append(part)
	}
}

func (namesParts *featuresNamesParts) setPrefixes(prefixesParts ...string) {
	namesParts.clear()
	for _, part := range prefixesParts {
		namesParts.append(part)
	}
	namesParts.prefixesLen = namesParts.nextInd
}

func (namesParts *featuresNamesParts) slice() []string {
	return namesParts.inner[:namesParts.nextInd]
}

func (namesParts *featuresNamesParts) append(part string) {
	namesParts.inner[namesParts.nextInd] = part
	namesParts.nextInd++
}

func (namesParts *featuresNamesParts) clear() {
	namesParts.prefixesLen = 0
	namesParts.nextInd = 0
}
