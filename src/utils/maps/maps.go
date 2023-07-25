package maps

import (
	"text2phenotype.com/fdl/utils"
	"fmt"
	"reflect"
)

func mapToStruct(fromMap *map[string]interface{}, toPtr interface{}) error {
	value := reflect.ValueOf(toPtr)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("%v is not a pointer", toPtr)
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		v := value.Kind()
		_ = v
		return fmt.Errorf("%v is not a struct pointer", toPtr)
	}
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		fInfo := valueType.Field(i)
		mapKey, ok := fInfo.Tag.Lookup("json")
		if !ok {
			continue
		}
		rawFieldContents, ok := (*fromMap)[mapKey]
		if !ok {
			continue
		}
		err := readValue(&rawFieldContents, fieldValue)
		if err != nil {
			return fmt.Errorf("got error at field %s: %w", fInfo.Name, err)
		}
	}
	return nil
}

func readValue(rawFieldContents *interface{}, fieldValue reflect.Value) error {
	switch fieldValue.Kind() {
	case reflect.Struct:
		value := asStruct(fieldValue)
		innerMap, ok := (*rawFieldContents).(map[string]interface{})
		if !ok {
			return nil
		}
		err := mapToStruct(&innerMap, value.pointer())
		if err != nil {
			return err
		}
	case reflect.Slice:
		err := readSlice(rawFieldContents, fieldValue)
		if err != nil {
			return err
		}
	case reflect.Map:
		err := readMap(rawFieldContents, fieldValue)
		if err != nil {
			return err
		}
	case reflect.Ptr:
		if *rawFieldContents == nil {
			return nil
		}
		value := asPointer(fieldValue)
		value.init()
		err := readValue(rawFieldContents, value.pointedValue())
		if err != nil {
			return err
		}
	default:
		// Most probably fieldValue is a primitive
		err := readPrimitive(rawFieldContents, fieldValue)
		if err != nil {
			return err
		}
	}
	return nil
}

func readPrimitive(rawValue *interface{}, fieldValue reflect.Value) (err error) {
	defer utils.RecoverWithError(&err)
	if rawValue == nil || *rawValue == nil {
		return nil
	}
	fieldValue.Set(reflect.ValueOf(*rawValue).Convert(fieldValue.Type()))
	return nil
}

func readSlice(fromValue *interface{}, sliceValue reflect.Value) error {
	value, ok := (*fromValue).([]interface{})
	if !ok {
		return fmt.Errorf("expected slice, got %v type", reflect.TypeOf(*fromValue))
	}
	elemType := sliceValue.Type().Elem()

	values := make([]reflect.Value, len(value))

	for index, elem := range value {
		elem := elem
		elemValue := reflect.New(elemType).Elem()
		err := readValue(&elem, elemValue)
		if err != nil {
			return err
		}
		values[index] = elemValue
	}
	sliceValue.Set(reflect.Append(sliceValue, values...))
	return nil
}

func readMap(fromValue *interface{}, mapValue reflect.Value) error {
	value, ok := (*fromValue).(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map, got %v type", reflect.TypeOf(*fromValue))
	}
	// mapValue is currently a nil map, we need to explicitly create nonnil map with reflect.MakeMap
	mv := reflect.MakeMap(mapValue.Type())
	elemType := mapValue.Type().Elem()
	for key, elem := range value {
		elem := elem
		elemValue := reflect.New(elemType).Elem()
		err := readValue(&elem, elemValue)
		if err != nil {
			return err
		}
		mv.SetMapIndex(reflect.ValueOf(key), elemValue)
	}
	mapValue.Set(mv)
	return nil
}

func updateMapFromStruct(mapToUpdate *map[string]interface{}, v interface{}) error {
	value := reflect.ValueOf(v)
	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("%v is not a pointer", v)
	}
	value = value.Elem()
	if value.Kind() != reflect.Struct {
		return fmt.Errorf("%v is not a struct pointer", v)
	}
	valueType := value.Type()
	for i := 0; i < value.NumField(); i++ {
		fieldValue := value.Field(i)
		fieldInfo := valueType.Field(i)
		mapKey, ok := fieldInfo.Tag.Lookup("json")
		if !ok {
			continue
		}
		rawFieldContents := (*mapToUpdate)[mapKey]
		updated, err := makeUpdatedValue(&rawFieldContents, &fieldValue)
		if err != nil {
			return fmt.Errorf("got error at field %s: %w", fieldInfo.Name, err)
		}
		if updated != nil {
			(*mapToUpdate)[mapKey] = *updated
		} else {
			(*mapToUpdate)[mapKey] = nil
		}
	}
	return nil
}

func makeUpdatedValue(current *interface{}, valuePtr *reflect.Value) (*interface{}, error) {
	v := *valuePtr
	switch v.Kind() {
	case reflect.Struct:
		value := asStruct(v)
		return makeMapFromStruct(current, &value)
	case reflect.Ptr:
		value := asPointer(v)
		if value.IsNil() {
			return nil, nil
		}
		pointed := value.pointedValue()
		return makeUpdatedValue(current, &pointed)
	case reflect.Slice:
		slice, err := makeSlice(v)
		if err != nil {
			return nil, err
		}
		r := interface{}(*slice)
		return &r, nil
	case reflect.Map:
		m, err := makeMap(v)
		if err != nil {
			return nil, err
		}
		r := interface{}(*m)
		return &r, nil
	default:
		// Most probably fieldValue is a primitive
		r := v.Interface()
		return &r, nil
	}
}

func makeMapFromStruct(current *interface{}, value *structValue) (*interface{}, error) {
	var innerMap map[string]interface{}
	if current == nil || *current == nil {
		// Map doesn't exist for specified key, creating one
		innerMap = map[string]interface{}{}
	} else if m, ok := (*current).(map[string]interface{}); ok {
		innerMap = m
	} else {
		return nil, fmt.Errorf(
			"expected inner structire to be map, got %v",
			reflect.TypeOf(*current),
		)
	}
	err := updateMapFromStruct(&innerMap, value.pointer())
	if err != nil {
		return nil, err
	}
	r := interface{}(innerMap)
	return &r, nil
}

func makeSlice(sliceField reflect.Value) (*[]interface{}, error) {
	slice := make([]interface{}, sliceField.Len())
	for index := 0; index < sliceField.Len(); index++ {
		elemValue := sliceField.Index(index)
		updatedValue, err := makeUpdatedValue(nil, &elemValue)
		if err != nil {
			return nil, err
		}
		if updatedValue == nil {
			slice[index] = nil
			continue
		}
		slice[index] = *updatedValue
	}
	return &slice, nil
}

func makeMap(mapField reflect.Value) (*map[string]interface{}, error) {
	m := map[string]interface{}{}
	iter := mapField.MapRange()
	for iter.Next() {
		keyValue, elemValue := iter.Key(), iter.Value()
		/*  We need this conversion because `elemValue` returned by `iter.Value()` is not addressable
			even if `elemValue.Kind() == reflect.Struct`
		More info here https://utcc.utoronto.ca/~cks/space/blog/programming/GoAddressableValues */
		elem := elemValue.Interface()
		elemValue = reflect.ValueOf(&elem).Elem()

		updatedValue, err := makeUpdatedValue(nil, &elemValue)
		if err != nil {
			return nil, err
		}
		if updatedValue == nil {
			m[keyValue.String()] = nil
			continue
		}
		m[keyValue.String()] = *updatedValue
	}
	return &m, nil
}
