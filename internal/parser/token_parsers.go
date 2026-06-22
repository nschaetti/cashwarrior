package parser

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/nschaetti/cashwarrior/internal/config"
	"github.com/nschaetti/cashwarrior/internal/domain"
	"github.com/nschaetti/cashwarrior/internal/utils"
)

func trimFlagPrefix(raw string) string {
	if strings.HasPrefix(raw, "--") {
		raw = strings.TrimPrefix(raw, "--")
	} else {
		raw = strings.TrimPrefix(raw, "-")
	}
	return raw
}

func isRange(raw string) bool {
	return strings.Contains(raw, "..")
}

func isList(raw string) bool {
	return strings.Contains(raw, ",")
}

func checkIntegerValue(raw string) bool {
	_, err := strconv.Atoi(raw)
	return err == nil
}

func checkFloatValue(raw string) bool {
	_, err := strconv.ParseFloat(raw, 64)
	return err == nil
}

func checkDateValue(raw string, config config.Config) bool {
	_, err := time.Parse(config.Display.DateFormat, raw)
	isTS := domain.IsTimeShortcut(raw)
	return err == nil || isTS
}

func parseDateValue(raw string, config config.Config) (time.Time, error) {
	parsedDate, err := time.Parse(config.Display.DateFormat, raw)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}

func IsValidPath(path string) bool {
	return path != "" && !strings.ContainsRune(path, '\x00')
}

func canonicalAttributeKey(key string) string {
	switch key {
	case "id", "T":
		return "identifier"
	default:
		return key
	}
}

// ParseArgFlag parses a flag argument.
func ParseArgFlag(raw string, config config.Config) (Arg, *ParseError) {
	var flagKey string
	var flagValue string
	var singleFlag bool

	if strings.Contains(raw, "=") {
		flagValue = strings.SplitN(raw, "=", 2)[1]
		singleFlag = false
	} else {
		flagValue = "true"
		singleFlag = true
	}

	if singleFlag {
		flagKey = trimFlagPrefix(raw)
	} else {
		flagKey = trimFlagPrefix(strings.SplitN(raw, "=", 2)[0])
	}

	// Get spec
	flagSpec, ok := FlagSpecs[flagKey]
	if !ok {
		return ArgFlag{}, &ParseError{
			Code:    ParseErrorUnknownFlag,
			Message: fmt.Sprintf("unknown flag key: %s", flagKey),
		}
	}

	if singleFlag {
		if flagSpec.Type == FlagValueTypeBool {
			flagValue = "true"
		} else {
			return ArgFlag{}, &ParseError{
				Code:    ParseErrorInvalidFlagValue,
				Message: fmt.Sprintf("invalid value given for %s flag: %s", flagKey, flagValue),
			}
		}
	}

	var flagItem FlagValueItem
	switch flagSpec.Type {
	case FlagValueTypeFloat:
		flagItemValue, err := strconv.ParseFloat(flagValue, 64)
		flagItem = FloatItem{Raw: flagValue, Value: flagItemValue}
		if err != nil {
			return ArgFlag{}, &ParseError{
				Code:    ParseErrorInvalidFlagValue,
				Message: fmt.Sprintf("invalid value given for %s flag: %s", flagKey, flagValue),
			}
		}
	case FlagValueTypeInteger:
		flagItemValue, err := strconv.Atoi(flagValue)
		flagItem = IntItem{Raw: flagValue, Value: int64(flagItemValue)}
		if err != nil {
			return ArgFlag{}, &ParseError{
				Code: ParseErrorInvalidFlagValue,
			}
		}
	case FlagValueTypeString:
		flagItem = StringItem{Raw: flagValue, Value: flagValue}
		break
	case FlagValueTypeBool:
		flagItem = BoolItem{Raw: flagValue, Value: true}
		break
	}

	return ArgFlag{Raw: raw, Key: flagKey, Value: flagItem}, nil
}

/*
 * Parse text arguments
 */

func ParseArgText(raw string, config config.Config) (Arg, *ParseError) {
	return ArgText{Raw: raw, Text: raw}, nil
}

/*
 * Parse tag arguments
 */

func ParseArgTag(raw string, config config.Config) (Arg, *ParseError) {
	return ArgTag{Raw: raw, Tag: raw, Negative: false}, nil
}

/*
 * Parse tag negative arguments
 */

func ParseArgTagNegative(raw string, config config.Config) (Arg, *ParseError) {
	return ArgTag{Raw: raw, Tag: raw, Negative: true}, nil
}

/*
 * Parse attribute arguments
 */

type TokenAttributeValueParser func(raw string, config config.Config) (AttributeValue, *ParseError)

var tokenAttributeParsers = map[AttributeValueType]TokenAttributeValueParser{
	AttributeValueTypeString:  parseTokenAttributeStringValue,
	AttributeValueTypeInteger: parseTokenAttributeIntegerValue,
	AttributeValueTypeFloat:   parseTokenAttributeFloatValue,
	AttributeValueTypeFile:    parseTokenAttributeFileValue,
	AttributeValueTypeDate:    parseTokenAttributeDateValue,
	AttributeValueTypeBool:    parseTokenAttributeBoolValue,
}

// parseTokenAttributeStringValue parses a string attribute value.
func parseTokenAttributeStringValue(raw string, config config.Config) (AttributeValue, *ParseError) {
	if isRange(raw) {
		parts := strings.Split(raw, "..")
		return AttributeValue{
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: StringItem{Raw: parts[0], Value: parts[0]},
				End:   StringItem{Raw: parts[1], Value: parts[1]},
			},
		}, nil
	} else if isList(raw) {
		sitems := strings.Split(raw, ",")
		items := make([]AttributeItem, 0, len(sitems))
		for _, item := range sitems {
			items = append(items, StringItem{Raw: item, Value: item})
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeList,
			Items:      items,
		}, nil
	}
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
		Value:      StringItem{Raw: raw, Value: raw},
	}, nil
}

// parseTokenAttributeIntegerValue parses an integer attribute value.
func parseTokenAttributeIntegerValue(raw string, config config.Config) (AttributeValue, *ParseError) {
	if isRange(raw) {
		parts := strings.Split(raw, "..")
		startValue := parts[0]
		endValue := parts[1]
		if !checkIntegerValue(startValue) {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidAttributeValue,
				Message: fmt.Sprintf("invalid integer range start: %s", startValue),
			}
		}
		if !checkIntegerValue(endValue) {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidAttributeValue,
				Message: fmt.Sprintf("invalid integer range end: %s", endValue),
			}
		}
		startInt, _ := strconv.Atoi(startValue)
		endInt, _ := strconv.Atoi(endValue)
		if startInt > endInt {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidRange,
				Message: fmt.Sprintf("invalid integer range: %s > %s", startValue, endValue),
			}
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: IntItem{Raw: parts[0], Value: int64(startInt)},
				End:   IntItem{Raw: parts[1], Value: int64(endInt)},
			},
		}, nil
	} else if isList(raw) {
		sitems := strings.Split(raw, ",")
		for _, item := range sitems {
			if !checkIntegerValue(item) {
				return AttributeValue{}, &ParseError{
					Code:    ParseErrorInvalidAttributeValue,
					Message: fmt.Sprintf("invalid integer list item: %s", item),
				}
			}
		}
		items := make([]AttributeItem, 0, len(sitems))
		for _, item := range sitems {
			intVal, _ := strconv.Atoi(item)
			items = append(items, IntItem{Raw: item, Value: int64(intVal)})
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeList,
			Items:      items,
		}, nil
	}
	if !checkIntegerValue(raw) {
		return AttributeValue{}, &ParseError{
			Code:    ParseErrorInvalidAttributeValue,
			Message: fmt.Sprintf("invalid integer: %s", raw),
		}
	}
	intVal, _ := strconv.Atoi(raw)
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
		Value:      IntItem{Raw: raw, Value: int64(intVal)},
	}, nil
}

// parseTokenAttributeFloatValue parses a float attribute value.
func parseTokenAttributeFloatValue(raw string, config config.Config) (AttributeValue, *ParseError) {
	if isRange(raw) {
		parts := strings.Split(raw, "..")
		startValue := parts[0]
		endValue := parts[1]
		if !checkFloatValue(startValue) {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidAttributeValue,
				Message: fmt.Sprintf("invalid float range start: %s", startValue),
			}
		}
		if !checkFloatValue(endValue) {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidAttributeValue,
				Message: fmt.Sprintf("invalid float range end: %s", endValue),
			}
		}
		floatStart, _ := strconv.ParseFloat(startValue, 64)
		floatEnd, _ := strconv.ParseFloat(endValue, 64)
		if floatStart > floatEnd {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidAttributeValue,
				Message: "invalid float range",
			}
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: FloatItem{Raw: parts[0], Value: floatStart},
				End:   FloatItem{Raw: parts[1], Value: floatEnd},
			},
		}, nil
	} else if isList(raw) {
		items := strings.Split(raw, ",")
		for _, item := range items {
			if !checkFloatValue(item) {
				return AttributeValue{}, &ParseError{
					Code:    ParseErrorInvalidAttributeValue,
					Message: fmt.Sprintf("invalid float list item: %s", item),
				}
			}
		}
		floatItems := make([]AttributeItem, 0, len(items))
		for _, item := range items {
			floatItem, _ := strconv.ParseFloat(item, 64)
			floatItems = append(floatItems, FloatItem{Raw: item, Value: floatItem})
		}
		return AttributeValue{
			ValueShape: AttributeValueShapeList,
			Items:      floatItems,
		}, nil
	}
	if !checkFloatValue(raw) {
		return AttributeValue{}, &ParseError{
			Code:    ParseErrorInvalidAttributeValue,
			Message: "invalid float",
		}
	}
	floatValue, _ := strconv.ParseFloat(raw, 64)
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
		Value:      FloatItem{Raw: raw, Value: floatValue},
	}, nil
}

// parseTokenAttributeFileValue parses a file attribute value.
func parseTokenAttributeFileValue(raw string, config config.Config) (AttributeValue, *ParseError) {
	if IsValidPath(raw) {
		return AttributeValue{ValueShape: AttributeValueShapeSingle, Raw: utils.ExpandPath(raw)}, nil
	}
	return AttributeValue{}, &ParseError{
		Code:    ParseErrorInvalidAttributeValue,
		Message: fmt.Sprintf("invalid file path: %s", raw),
	}
}

// parseTokenAttributeDateValue parses a date attribute value.
func parseTokenAttributeDateValue(raw string, config config.Config) (AttributeValue, *ParseError) {
	if domain.IsTimeShortcut(raw) {
		//startDate, endDate, err := domain.GetTimeShortcut(raw)
		//if err != nil {
		//	return AttributeValue{}, &ParseError{
		//		Code:    ParseErrorInvalidAttributeValue,
		//		Message: fmt.Sprintf("invalid date shortcut: %s", raw),
		//	}
		//}
		//return AttributeValue{
		//	ValueShape: AttributeValueShapeRange,
		//	Range: AttributeRange{
		//		Start: TimeItem{Raw: raw, Value: startDate},
		//		End:   TimeItem{Raw: raw, Value: endDate},
		//	},
		//}, nil
		return AttributeValue{
			Raw:        raw,
			ValueShape: AttributeValueShapeShortcut,
			Shortcut:   AttributeShortcut{Name: raw},
		}, nil
	}

	if isRange(raw) {
		parts := strings.Split(raw, "..")
		startValue := parts[0]
		endValue := parts[1]
		if !checkDateValue(startValue, config) || domain.IsTimeShortcut(startValue) {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidAttributeValue,
				Message: fmt.Sprintf("invalid date range start: %s", startValue),
			}
		}
		if !checkDateValue(endValue, config) || domain.IsTimeShortcut(endValue) {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidAttributeValue,
				Message: fmt.Sprintf("invalid date range end: %s", endValue),
			}
		}
		startDate, _ := parseDateValue(startValue, config)
		endDate, _ := parseDateValue(endValue, config)
		if startDate.After(endDate) {
			return AttributeValue{}, &ParseError{
				Code:    ParseErrorInvalidRange,
				Message: fmt.Sprintf("invalid date range: %s > %s", startValue, endValue),
			}
		}
		return AttributeValue{
			Raw:        raw,
			ValueShape: AttributeValueShapeRange,
			Range: AttributeRange{
				Start: TimeItem{Raw: parts[0], Value: startDate},
				End:   TimeItem{Raw: parts[1], Value: endDate},
			},
		}, nil
	} else if isList(raw) {
		items := strings.Split(raw, ",")
		for _, item := range items {
			if !checkDateValue(item, config) || domain.IsTimeShortcut(item) {
				return AttributeValue{}, &ParseError{
					Code:    ParseErrorInvalidAttributeValue,
					Message: fmt.Sprintf("invalid date list item: %s", item),
				}
			}
		}
		dateItems := make([]AttributeItem, 0, len(items))
		for _, item := range items {
			dateItem, _ := parseDateValue(item, config)
			dateItems = append(dateItems, TimeItem{Raw: item, Value: dateItem})
		}
		return AttributeValue{
			Raw:        raw,
			ValueShape: AttributeValueShapeList,
			Items:      dateItems,
		}, nil
	}
	if !checkDateValue(raw, config) {
		return AttributeValue{}, &ParseError{
			Code:    ParseErrorInvalidAttributeValue,
			Message: fmt.Sprintf("invalid date: %s", raw),
		}
	}
	dateValue, _ := parseDateValue(raw, config)
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
		Value:      TimeItem{Raw: raw, Value: dateValue},
	}, nil
}

// parseTokenAttributeBoolValue parses a boolean attribute value.
func parseTokenAttributeBoolValue(raw string, config config.Config) (AttributeValue, *ParseError) {
	return AttributeValue{
		ValueShape: AttributeValueShapeSingle,
		Raw:        raw,
	}, nil
}

// ParseArgAttribute parses an attribute argument.
func ParseArgAttribute(raw string, config config.Config) (Arg, *ParseError) {
	var attributeKey string
	var attributeValue AttributeValue
	var err *ParseError

	// Key and values
	attributeKey = strings.SplitN(raw, ":", 2)[0]
	attributeValueRaw := strings.SplitN(raw, ":", 2)[1]

	// Get attribute spec
	attrSpec, ok := AttributeSpecs[attributeKey]
	if !ok {
		return ArgAttribute{}, &ParseError{
			Code:    ParseErrorUnknownAttributeKey,
			Message: fmt.Sprintf("unknown attribute key: %s", attributeKey),
		}
	}

	// Clear attribute
	if attributeValueRaw == "" {
		return ArgAttribute{Raw: raw, Key: canonicalAttributeKey(attributeKey), Clear: true}, nil
	}

	// Parse attribute value
	attributeValue, err = tokenAttributeParsers[attrSpec.Type](attributeValueRaw, config)
	if err != nil {
		return ArgAttribute{}, err
	}

	return ArgAttribute{Raw: raw, Key: canonicalAttributeKey(attributeKey), Value: attributeValue}, nil
}
