package elastigo

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/nsf/jsondiff"
)

type MyInnerStruct struct {
	Bar string `json:"bar" es:"text"`
}
type MyStruct struct {
	Foo   int           `json:"foo,omitempty"`
	Inner MyInnerStruct `json:"inner"`
}

func compare(t *testing.T, mapping, expectedDTO Mapping) {
	expectedBytes, err := json.Marshal(expectedDTO)
	if err != nil {
		t.Fatal(err)
	}
	actual, err := json.Marshal(mapping)
	if err != nil {
		t.Fatal(err)
	}
	opts := jsondiff.Options{}
	if diff, s := jsondiff.Compare(expectedBytes, actual, &opts); diff != jsondiff.FullMatch {
		t.Logf("Expected:%s", string(expectedBytes))
		t.Logf("Actual:%s", string(actual))
		t.Fatalf(s)
	}
}

func TestMappingGen(t *testing.T) {
	var m MyStruct
	mapping, err := GenerateMapping(reflect.TypeOf(m))
	if err != nil {
		t.Fatal(err)
	}
	var nested = "nested"
	var intType = "integer"
	var textType = "text"
	expectedDTO := Mapping{
		Properties: map[string]Mapping{
			"foo": Mapping{
				Type: &intType,
			},
			"inner": Mapping{
				Type: &nested,
				Properties: map[string]Mapping{
					"bar": Mapping{
						Type: &textType,
					},
				},
			},
		},
	}

	compare(t, *mapping, expectedDTO)
}

type Dog struct {
	Name     string `json:"name" es:",eager_global_ordinals"`
	UniqueID string `json:"unique_id" es:",indexignore"`
}

func TestCustomESProps(t *testing.T) {
	var dog Dog
	mapping, err := GenerateMapping(reflect.TypeOf(dog))
	if err != nil {
		t.Fatal(err)
	}
	var keywordType = "keyword"
	var falsy = false
	var truthy = true
	expectedDTO := Mapping{
		Properties: map[string]Mapping{
			"name": Mapping{
				Type:                &keywordType,
				EagerGlobalOrdinals: &truthy,
			},
			"unique_id": Mapping{
				Type:  &keywordType,
				Index: &falsy,
			},
		},
	}
	compare(t, *mapping, expectedDTO)
}
