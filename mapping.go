package elastigo

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/emirpasic/gods/sets/hashset"
)

const esBOOL = "boolean"
const esBYTE = "byte"
const esSHORT = "short"
const esINT = "integer"
const esLONG = "long"
const esFLOAT = "float"
const esDOUBLE = "double"
const esKEYWORD = "keyword"
const esTEXT = "text"
const esDATE = "date"
const esIP = "ip"

var esPrimitives = []string{
	esBOOL,
	esINT,
	esLONG,
	esFLOAT,
	esDOUBLE,
	esKEYWORD,
	esTEXT,
	esDATE,
	esIP,
}

const esGEOPOINT = "geo_point"
const esGEOSHAPE = "geo_shape"
const esCOMPLETION = "completion"

var esSpecials = []string{
	esGEOPOINT,
	esGEOSHAPE,
	esCOMPLETION,
}

// primitives
var primitiveKinds = []reflect.Kind{
	reflect.Bool,
	reflect.Int,
	reflect.Int8,
	reflect.Int16,
	reflect.Int32,
	reflect.Int64,
	reflect.Float32,
	reflect.Float64,
	reflect.String,
}

// default mappings
var defaultElasticTypeMap = map[reflect.Kind]string{
	reflect.Bool:    esBOOL,
	reflect.Int:     esINT,
	reflect.Int8:    esBYTE,
	reflect.Int16:   esSHORT,
	reflect.Int32:   esINT,
	reflect.Int64:   esLONG,
	reflect.Float32: esFLOAT,
	reflect.Float64: esDOUBLE,
	reflect.String:  esKEYWORD,
}

// createPrimitiveSet https://emersion.fr/blog/2017/sets-in-go/
func createPrimitiveSet(primitivesList []string) map[string]struct{} {
	hs := make(map[string]struct{})
	for _, s := range primitivesList {
		hs[s] = struct{}{}
	}
	return hs
}

var esPrimitiveSet = createPrimitiveSet(esPrimitives)

// isPrimitiveESType determines if the type is
// not an object
func isPrimitiveESType(tName string) bool {
	_, ok := esPrimitiveSet[tName]
	return ok
}

// getTagValues returns the comma-separated values of
// a given tag name
func getTagValues(tag reflect.StructTag, tagName string) []string {
	values := tag.Get(tagName)
	if len(values) > 0 {
		strSplit := strings.Split(values, ",")
		return strSplit
	}
	return []string{}
}

func getESTypeAndProps(tag reflect.StructTag) (string, *hashset.Set) {
	name := ""
	var hs *hashset.Set
	esValues := getTagValues(tag, "es")
	if len(esValues) == 0 {
		return name, hs
	}
	if len(esValues) == 1 {
		return esValues[0], hs
	}
	hs = hashset.New()
	for _, prop := range esValues[1:] {
		hs.Add(prop)
	}
	name = esValues[0]
	return name, hs
}

func getJSONNameFromTag(tag reflect.StructTag) string {
	jsonValues := getTagValues(tag, "json")
	if len(jsonValues) > 0 {
		// in addition to the tag's field
		// name, it could have additional metadata,
		// comma-separated
		// e.g. `json:"fee,omitempty"`
		return jsonValues[0]
	}
	return ""
}

// getUnderlyingType returns the type referenced by
// a pointer or array
func getUnderlyingType(t reflect.Type) reflect.Type {
	// Elasticsearch's schema doesn't have an array
	// type. Instead, the schema just specifies the
	// type of the elements in an array
	if t.Kind() == reflect.Array ||
		t.Kind() == reflect.Slice ||
		t.Kind() == reflect.Ptr ||
		t.Kind() == reflect.UnsafePointer {

		// could be a pointer to pointer, for example
		return getUnderlyingType(t.Elem())
	}
	// otherwise, just return t
	return t
}

// isPrimitiveGoType determines if the type
// is not an object
func isPrimitiveGoType(t reflect.Type) bool {
	// getUnderlyingType should be idempotent
	underlyingT := getUnderlyingType(t)
	return underlyingT.Kind() != reflect.Struct
}

// extractFields takes a struct type, reads its fields, and
// stores it in the slice that `fieldsPtr` references,
// flattening any embedded structs
func extractFields(t reflect.Type, fieldsPtr *[]reflect.StructField) {
	if fieldsPtr == nil {
		panic(fmt.Sprintf("Passed in nil slice for flattening %s", t.Name()))
	}
	tActual := getUnderlyingType(t)
	if tActual.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < tActual.NumField(); i++ {
		memberField := tActual.Field(i)
		typeOfField := getUnderlyingType(memberField.Type)
		// if this is an anonymous type, we want to flatten it
		if typeOfField.Kind() == reflect.Struct && memberField.Anonymous {
			extractFields(typeOfField, fieldsPtr)
		} else {
			*fieldsPtr = append(*fieldsPtr, memberField)
		}
	}
}

// generateMappingRecur generates an Elasticsearch schema
// from a given type
func generateMappingRecur(t reflect.Type, customESType string) (Mapping, error) {
	var goType reflect.Type
	// in Elasticsearch, there is no dedicated
	// 'array' type--the type is the type of the
	// element
	goType = getUnderlyingType(t)
	// check if primitive type
	if isPrimitiveGoType(goType) {
		var esPrimType string
		if len(customESType) > 0 {
			if !isPrimitiveESType(customESType) {
				panic(fmt.Sprintf(
					"%s is an invalid Elasticsearch type", customESType))
			}
			esPrimType = customESType
		} else {
			esPrimType = defaultElasticTypeMap[goType.Kind()]
		}
		return Mapping{
			Type: &esPrimType,
		}, nil
	}

	// this is a struct
	// get nested type properties
	var props = make(map[string]Mapping)
	// for embedded structs, we want to
	// flatten their members
	var allFields []reflect.StructField
	extractFields(goType, &allFields)
	for _, goField := range allFields {
		goFieldType := goField.Type

		// get the name of the member
		var memberName string
		if name := getJSONNameFromTag(goField.Tag); len(name) > 0 {
			memberName = name
		} else {
			memberName = goField.Name
		}

		// read the custom Elasticsearch type from the tag if it exists
		esType, customProps := getESTypeAndProps(goField.Tag)

		fmt.Printf("field name: %s\n\ttype:%s\n", memberName, goFieldType.Kind())
		innerMapping, err := generateMappingRecur(goFieldType, esType)
		if err != nil {
			panic(err)
		}
		// get the generated type
		var generatedESType = ""
		if innerMapping.Type != nil {
			generatedESType = *innerMapping.Type
		}
		if customProps != nil {
			// check for indexignore
			if customProps.Contains("indexignore") {
				var shouldIndex = false
				innerMapping.Index = &shouldIndex
			}
			var epochSecond = "epoch_second"
			if customProps.Contains(epochSecond) {
				if generatedESType != esDATE {
					panic(
						fmt.Sprintf(
							"%s not a valid format for Elasticsearch type %s",
							epochSecond,
							esType,
						))
				}
				innerMapping.Format = &epochSecond
			} else if customProps.Contains("epoch_ms") {
				var epochMS = "epoch_millis"
				if generatedESType != esDATE {
					panic(
						fmt.Sprintf(
							"%s not a valid format for Elasticsearch type %s",
							epochMS,
							esType,
						))
				}
				innerMapping.Format = &epochMS
			}
			const eagerGlobalOrdinalsProp = "eager_global_ordinals"
			if customProps.Contains(eagerGlobalOrdinalsProp) {
				if generatedESType != esKEYWORD && generatedESType != esTEXT {
					panic(
						fmt.Sprintf("%s can only be set for %s and %s fields",
							eagerGlobalOrdinalsProp,
							esKEYWORD,
							esTEXT,
						))
				}
				eagerGlobalOrdinals := true
				innerMapping.EagerGlobalOrdinals = &eagerGlobalOrdinals
			}
		}
		props[memberName] = innerMapping
	}
	var nested = "nested"
	mapping := Mapping{
		Type:       &nested,
		Properties: props,
	}
	return mapping, nil
}

// GenerateMapping generates an Elasticsearch mapping
// from a given Go type
func GenerateMapping(t reflect.Type) (*Mapping, error) {
	m, err := generateMappingRecur(t, "")
	if err != nil {
		return nil, err
	}
	// delete the extra "type" member if it exists (it should)
	m.Type = nil
	return &m, nil
}
