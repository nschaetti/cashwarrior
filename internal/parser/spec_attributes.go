package parser

import "fmt"

// getAttributeSpec returns the AttributeSpec for the given name.
func getAttributeSpec(name string) (AttributeSpec, error) {
	attrSpec, ok := AttributeSpecs[name]
	if !ok {
		return AttributeSpec{}, fmt.Errorf("unknown attribute %q", name)
	}
	return attrSpec, nil
}

func buildCompleteAttributeSpec(name string, shapes AttributeValueShape, settable bool, clearable bool) AttributeSpec {
	attrSpec, err := getAttributeSpec(name)
	if err != nil {
		panic(err)
	}
	if settable && !attrSpec.AllowSettable {
		panic(fmt.Errorf("attribute %s is not settable", name))
	}
	if clearable && !attrSpec.AllowClear {
		panic(fmt.Errorf("attribute %s is not clearable", name))
	}
	attrSpec.Settable = settable
	attrSpec.Clearable = clearable
	attrSpec.Shapes = shapes
	return attrSpec
}

func buildAttributeSpec(name string) AttributeSpec {
	attrSpec, err := getAttributeSpec(name)
	if err != nil {
		panic(err)
	}
	return attrSpec
}

// settableOnlyAttribute creates an AttributeSpec with the given name and shapes and sets the attribute as settable.
func settableOnlyAttribute(name string) AttributeSpec {
	attrSpec, err := getAttributeSpec(name)
	if err != nil {
		panic(err)
	}
	if !attrSpec.AllowSettable {
		panic(fmt.Errorf("attribute %s is not settable", name))
	}
	attrSpec.Settable = true
	return attrSpec
}

// settableOnlyAttribute creates an AttributeSpec with the given name and shapes and sets the attribute as clearable.
func setOrClearAttribute(name string) AttributeSpec {
	attrSpec := settableOnlyAttribute(name)
	if !attrSpec.AllowClear {
		panic(fmt.Errorf("attribute %s is not clearable", name))
	}
	attrSpec.Clearable = true
	return attrSpec
}

// AttributeSpecs is a map of all the attributes that can be used in a transaction.
var AttributeSpecs = map[string]AttributeSpec{
	"amount":          {Name: "amount", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange, Type: AttributeValueTypeFloat, AllowSettable: true},
	"account":         {Name: "account", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true},
	"balance":         {Name: "balance", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange, Type: AttributeValueTypeFloat, AllowSettable: true},
	"category":        {Name: "category", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true, AllowClear: true},
	"currency":        {Name: "currency", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true},
	"date":            {Name: "date", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange | AttributeValueShapeShortcut, Type: AttributeValueTypeDate, AllowSettable: true},
	"desc":            {Name: "desc", AllowedShapes: AttributeValueShapeSingle, Type: AttributeValueTypeBool, AllowSettable: true},
	"description":     {Name: "description", AllowedShapes: AttributeValueShapeSingle, Type: AttributeValueTypeString, AllowSettable: true},
	"from":            {Name: "from", AllowedShapes: AttributeValueShapeSingle, Type: AttributeValueTypeString, AllowSettable: true},
	"group":           {Name: "group", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true, AllowClear: true},
	"identifier":      {Name: "identifier", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange, Type: AttributeValueTypeString, AllowSettable: true},
	"id":              {Name: "id", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange, Type: AttributeValueTypeString, AllowSettable: true},
	"T":               {Name: "T", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange, Type: AttributeValueTypeString, AllowSettable: true},
	"initial-balance": {Name: "initial-balance", AllowedShapes: AttributeValueShapeSingle, Type: AttributeValueTypeFloat, AllowSettable: true},
	"month":           {Name: "month", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true},
	"name":            {Name: "name", AllowedShapes: AttributeValueShapeSingle, Type: AttributeValueTypeString, AllowSettable: true},
	"output":          {Name: "output", AllowedShapes: AttributeValueShapeSingle, Type: AttributeValueTypeFile, AllowSettable: true},
	"order":           {Name: "order", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true},
	"parent":          {Name: "parent", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true},
	"size":            {Name: "size", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange, Type: AttributeValueTypeInteger, AllowSettable: true, AllowClear: false},
	"store":           {Name: "store", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true},
	"tag":             {Name: "tag", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList, Type: AttributeValueTypeString, AllowSettable: true},
	"to":              {Name: "to", AllowedShapes: AttributeValueShapeSingle, Type: AttributeValueTypeString, AllowSettable: true},
	"year":            {Name: "year", AllowedShapes: AttributeValueShapeSingle | AttributeValueShapeList | AttributeValueShapeRange, Type: AttributeValueTypeInteger, AllowSettable: true},
}

var transactionFilterAttributes = []string{"account", "currency", "store", "desc", "date", "group", "identifier"}
