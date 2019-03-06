package model

type AppModelData struct {
	LangsToManage []string
	SourceToParse []string

	ResultSheet map[string]LocaleItem
}

type LocaleItem struct {
	PropertyName string
	Group        string
	Translations map[string]string
}

//AddTranslation to initialize and add a translation for lang to
func (localeItem *LocaleItem) AddTranslation(lang, value string, allLangs []string) {
	if localeItem.Translations == nil {
		localeItem.Translations = make(map[string]string)
		for _, tempLang := range allLangs {
			localeItem.Translations[tempLang] = NoTranslationPlaceholder
		}
	}
	localeItem.Translations[lang] = value
}

//GetTranslation to retrive translation by lang
func (localeItem *LocaleItem) GetTranslation(lang string) (string, bool) {
	if localeItem.Translations != nil {
		return localeItem.Translations[lang], true
	}
	return NoTranslationPlaceholder, false
}

//NoTranslationPlaceholder is a placeholder for missing string translation
const NoTranslationPlaceholder = "**missing-translation**"

//ResourceFileExt ext file for resources
const ResourceFileExt = ".properties"
