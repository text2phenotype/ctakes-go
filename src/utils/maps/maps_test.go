package maps

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

type testSlicesStruct struct {
	SrtSliceField           []string                     `json:"str_slice"`
	StructSliceField        []testStructsStructDblInner  `json:"structs_slice"`
	PointerSliceField       []*testStructsStructDblInner `json:"pointers_slice"`
	StringPointerSliceField []*string                    `json:"string_pointers_slice"`
}

type testMapsStruct struct {
	SrtMapField           map[string]string                     `json:"str_map"`
	StructMapField        map[string]testStructsStructDblInner  `json:"structs_map"`
	PointerMapField       map[string]*testStructsStructDblInner `json:"pointers_map"`
	StringPointerMapField map[string]*string                    `json:"string_pointers_map"`
}

type testPrimitivesStruct struct {
	StrField   string  `json:"str_field"`
	IntField   int     `json:"int_field"`
	FloatField float64 `json:"float_field"`
	BoolField  bool    `json:"bool_field"`
}

type testCorruptedPrimitiveStruct struct {
	StrField   int     `json:"str_field"`
	IntField   int     `json:"int_field"`
	FloatField float64 `json:"float_field"`
	BoolField  bool    `json:"bool_field"`
}

type testStructsStructDblInner struct {
	DblInnerField bool `json:"inner_bool_field"`
}

type testStructsStructInner struct {
	InnerField  string                    `json:"inner_field"`
	InnerStruct testStructsStructDblInner `json:"dblInner_struct_field"`
}

type testStructsStructInnerPtr struct {
	InnerField  string                     `json:"inner_field"`
	InnerStruct *testStructsStructDblInner `json:"dblInner_struct_field"`
}

type testStructsPtrStruct struct {
	StructField testStructsStructInnerPtr `json:"inner_struct_field"`
}

type testStructsStruct struct {
	StructField testStructsStructInner `json:"inner_struct_field"`
}

func preparePrimitives() *testPrimitivesStruct {
	return &testPrimitivesStruct{
		StrField:   "str value",
		IntField:   1243,
		FloatField: 434,
		BoolField:  false,
	}
}

func prepareStructs() *testStructsStruct {
	return &testStructsStruct{
		StructField: testStructsStructInner{
			InnerField: "inner field",
			InnerStruct: testStructsStructDblInner{
				DblInnerField: true,
			},
		}}
}

func prepareStructsWithNilPointer() *testStructsPtrStruct {
	return &testStructsPtrStruct{
		StructField: testStructsStructInnerPtr{
			InnerField: "inner field",
		}}
}

func prepareStructsWithPointer() *testStructsPtrStruct {
	return &testStructsPtrStruct{
		StructField: testStructsStructInnerPtr{
			InnerField: "inner field",
			InnerStruct: &testStructsStructDblInner{
				DblInnerField: true,
			},
		}}
}

func prepareSlices() *testSlicesStruct {
	var s1, s2, s3 = "hi", "there", "world"
	return &testSlicesStruct{
		SrtSliceField: []string{"Hello", "world"},
		StructSliceField: []testStructsStructDblInner{
			{DblInnerField: true},
		},
		PointerSliceField: []*testStructsStructDblInner{
			{DblInnerField: true},
			{DblInnerField: false},
			nil,
		},
		StringPointerSliceField: []*string{
			&s1, nil, &s2, &s3,
		},
	}
}

func prepareMaps() *testMapsStruct {
	var s1, s2, s3 = "hi", "there", "world"
	return &testMapsStruct{
		SrtMapField: map[string]string{"1": "Hello", "2": "world"},
		StructMapField: map[string]testStructsStructDblInner{
			"1": {DblInnerField: true},
		},
		PointerMapField: map[string]*testStructsStructDblInner{
			"1": {DblInnerField: true},
			"2": {DblInnerField: false},
			"3": nil,
		},
		StringPointerMapField: map[string]*string{
			"1": &s1, "2": nil, "3": &s2, "4": &s3,
		},
	}
}

// TODO Slice of pointers
func TestFillMap(t *testing.T) {
	for name, prepared := range preparedStructs {
		t.Run(fmt.Sprintf("Correct %s", name), testCorrectStructType(prepared))
	}
	t.Run("Corrupted primitive", testCorruptedPrimitives)
	t.Run("Struct and pointer", testStructToPointer)
}

var preparedStructs = map[string]interface{}{
	"primitives":                  preparePrimitives(),
	"structs":                     prepareStructs(),
	"slices":                      prepareSlices(),
	"maps":                        prepareMaps(),
	"structs with nil pointer":    prepareStructsWithNilPointer(),
	"structs with filled pointer": prepareStructsWithPointer(),
}

func TestUpdateMap(t *testing.T) {
	for name, prepared := range preparedStructs {
		t.Run(fmt.Sprintf("Correct %s", name), testUpdateCorrectStructType(prepared))
	}
}

func testCorrectStructType(prepared interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		b, err := json.Marshal(prepared)
		if err != nil {
			t.Error("Failed to marshall prepared struct", prepared, err)
			return
		}
		structType := reflect.ValueOf(prepared).Elem().Type()
		newPrt := reflect.New(structType).Interface()

		var raw map[string]interface{}
		err = json.Unmarshal(b, &raw)
		if err != nil {
			t.Error("Failed to unmarshall prepared struct", prepared, err)
			return
		}
		err = mapToStruct(&raw, newPrt)
		if err != nil {
			t.Error("Failed to fill from map", prepared, err)
			return
		}
		if !reflect.DeepEqual(prepared, newPrt) {
			t.Error("Got unequal structs after parsing", prepared, newPrt)
		}
	}
}

func testCorruptedPrimitives(t *testing.T) {
	correct := preparePrimitives()
	var corrupt testCorruptedPrimitiveStruct

	b, err := json.Marshal(correct)
	if err != nil {
		t.Error("Failed to marshall correct struct", correct, err)
		return
	}
	var raw map[string]interface{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		t.Error("Failed to unmarshall correct struct", correct, err)
		return
	}
	err = mapToStruct(&raw, &corrupt)
	if err == nil {
		t.Error("mapToStruct should return error when types could not be converted")
	}
}

func testUpdateCorrectStructType(prepared interface{}) func(t *testing.T) {
	return func(t *testing.T) {
		updatedMap := make(map[string]interface{})
		err := updateMapFromStruct(&updatedMap, prepared)
		if err != nil {
			t.Error("Failed to update map", prepared, err)
		}
		preparedBytes, err := json.Marshal(prepared)
		if err != nil {
			t.Error("Failed to marshal prepared struct", updatedMap, err)
		}
		var preparedMap map[string]interface{}
		err = json.Unmarshal(preparedBytes, &preparedMap)
		if err != nil {
			t.Error("Failed to unmarshal prepared bytes", updatedMap, err)
		}
		preparedMapBytes, err := json.Marshal(preparedMap)
		if err != nil {
			t.Error("Failed to marshal prepared map", updatedMap, err)
		}
		updatedMapBytes, err := json.Marshal(updatedMap)
		if err != nil {
			t.Error("Failed to marshal updated map", updatedMap, err)
		}
		if string(preparedMapBytes) != string(updatedMapBytes) {
			t.Error(
				"updateMapFromStruct should create correct copy of object",
				string(preparedMapBytes),
				string(updatedMapBytes),
			)
		}
	}
}

func testStructToPointer(t *testing.T) {
	simple := prepareStructs()
	var ptrStruct testStructsPtrStruct

	b, err := json.Marshal(simple)
	if err != nil {
		t.Error("Failed to marshall primary struct", simple, err)
		return
	}
	var raw map[string]interface{}
	err = json.Unmarshal(b, &raw)
	if err != nil {
		t.Error("Failed to unmarshall primary struct", simple, err)
		return
	}
	err = mapToStruct(&raw, &ptrStruct)
	if err != nil {
		t.Error("Failed to fill from map", err)
	}
	if !reflect.DeepEqual(simple.StructField.InnerStruct, *ptrStruct.StructField.InnerStruct) {
		t.Error("Got unequal structs after parsing", simple, ptrStruct)
	}
}
