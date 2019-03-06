package model

import "flag"

type LocaleConverterActions string
type Platforms string

const (
	LocaleToTable LocaleConverterActions = "L2T"
	TableToLocale LocaleConverterActions = "T2L"

	Flex Platforms = "Flex"
	Java Platforms = "Java"
)

type LocaleConverterParameters struct {
	Action          LocaleConverterActions
	SourcePath      string
	DestinationPath string
	Platform        Platforms
}

func parseAction(action string) LocaleConverterActions {
	switch action {
	case "L2T":
		return LocaleToTable
	case "T2L":
		return TableToLocale
	}
	return LocaleToTable
}

func parsePlatform(platform string) Platforms {
	switch platform {
	case "Flex":
		return Flex
	case "Java":
		return Java
	}
	return Flex
}

func ParseParameters() *LocaleConverterParameters {

	params := new(LocaleConverterParameters)
	sourceParam := flag.String("source", "", "source path locale folder to process")
	destParam := flag.String("dest", "", "destination spreadsheet file to save result")
	actionParam := flag.String("action", "L2T | T2L", "action to execute, L2T = locale to table, T2L = table to locale")
	platformParam := flag.String("platform", "Flex | Java", "locale folder and .properties files configuration")

	flag.Parse()

	params.SourcePath = *sourceParam
	params.DestinationPath = *destParam
	params.Action = parseAction(*actionParam)
	params.Platform = parsePlatform(*platformParam)

	return params
}
