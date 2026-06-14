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
}
