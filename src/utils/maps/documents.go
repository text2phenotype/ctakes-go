package maps

import (
	"text2phenotype.com/fdl/utils"
	"encoding/json"
	"reflect"
)

type PartialDocument interface {
	getRaw() *map[string]interface{}
	setRaw(*map[string]interface{})
	MarshalJSON() ([]byte, error)
}

type BaseDocument struct {
	rawMap *map[string]interface{}
}

func (doc *BaseDocument) getRaw() *map[string]interface{} {
	return doc.rawMap
}

func (doc *BaseDocument) setRaw(raw *map[string]interface{}) {
	doc.rawMap = raw
}

func (doc *BaseDocument) MarshalJSON() ([]byte, error) {
	return json.Marshal(doc.getRaw())
}

func FillFromMap(doc PartialDocument, from *map[string]interface{}) error {
	err := mapToStruct(from, doc)
	if err != nil {
		return err
	}
	doc.setRaw(from)
	return nil
}

func CopyValues(from PartialDocument, to PartialDocument) error {
	raw := from.getRaw()
	err := mapToStruct(raw, to)
	if err != nil {
		return err
	}
	cachedMap := map[string]interface{}{}
	err = updateMapFromStruct(&cachedMap, to)
	if err != nil {
		return err
	}
	to.setRaw(&cachedMap)
	return nil
}

func ApplyUpdates(doc PartialDocument, updateFunc interface{}) (err error) {
	if updateFunc == nil {
		return nil
	}
	defer utils.RecoverWithError(&err)
	funcValue := reflect.ValueOf(updateFunc)
	docValue := reflect.ValueOf(doc)
	funcValue.Call([]reflect.Value{docValue})
	err = updateMapFromStruct(doc.getRaw(), doc)
	return
}
