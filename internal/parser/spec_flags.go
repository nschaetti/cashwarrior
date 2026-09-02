package parser

var FlagSpecs = map[string]FlagSpec{
	"help": {
		Name: "help",
		Type: FlagValueTypeBool,
	},
	"version": {
		Name: "version",
		Type: FlagValueTypeBool,
	},
	"debug": {
		Name: "debug",
		Type: FlagValueTypeBool,
	},
	"format": {
		Name: "format",
		Type: FlagValueTypeString,
	},
	"json": {
		Name: "json",
		Type: FlagValueTypeBool,
	},
	"yes": {
		Name: "yes",
		Type: FlagValueTypeBool,
	},
}
