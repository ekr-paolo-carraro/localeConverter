package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/kpango/glg"

	"github.com/ekr-paolo-carraro/localeConverter/apputils"
	"github.com/ekr-paolo-carraro/localeConverter/model"
)

func main() {

	apputils.InitLoggin()

	params := model.ParseParameters()

	if params.Action == model.LocaleToTable {
		processL2T(params)
	} else if params.Action == model.TableToLocale {
		processT2L(params)
	}
}

func processL2T(params *model.LocaleConverterParameters) {
	var appModelData model.AppModelData
	var tempError error

	//source folder to convert
	localeDirRef, tempError := os.Open(params.SourcePath)
	if apputils.CheckError(tempError) {
		apputils.StopRunning()
	}

	//source folder listing
	sourceFolderItems, tempError := localeDirRef.Readdir(0)
	if apputils.CheckError(tempError) {
		apputils.StopRunning()
	}
	localeDirRef.Close()

	appModelData.SourceToParse = []string{}

	if params.Platform == model.Flex {
		parseFlexSources(sourceFolderItems, &appModelData, params)
	} else if params.Platform == model.Java {
		parseJavaSources(sourceFolderItems, &appModelData, params)
	}

	appModelData.ResultSheet = make(map[string]model.LocaleItem)

	for _, tempLang := range appModelData.LangsToManage {
		for _, tempSourceToParse := range appModelData.SourceToParse {
			var tempPathSource string

			if params.Platform == model.Flex {
				tempPathSource = params.SourcePath + "/" + tempLang + "/" + tempSourceToParse
			} else if params.Platform == model.Java {
				tempPathSource = params.SourcePath + "/" + tempSourceToParse + "_" + tempLang + model.ResourceFileExt
			}

			apputils.WriteLog("Parsing "+tempPathSource, glg.INFO)
			tempSource, tempErr := os.Open(tempPathSource)

			if apputils.CheckError(tempErr) {
				continue
			} else {
				localePropertyLineScanner := bufio.NewScanner(tempSource)
				localePropertyLineScanner.Split(bufio.ScanLines)

				for localePropertyLineScanner.Scan() {
					candidateLine := localePropertyLineScanner.Text()

					var candidateProp = ""
					candidateProp, value := parseSingleLine(candidateLine, "=")
					if candidateProp == "" {
						candidateProp, value = parseSingleLine(candidateLine, "\t")
					}

					if candidateProp != "" {
						candidateLocaleItem, exists := appModelData.ResultSheet[candidateProp+tempSourceToParse]
						if exists {
							candidateLocaleItem.AddTranslation(tempLang, value, appModelData.LangsToManage)
						} else {
							localeItem := model.LocaleItem{PropertyName: candidateProp, Group: tempSourceToParse}
							localeItem.AddTranslation(tempLang, value, appModelData.LangsToManage)
							appModelData.ResultSheet[candidateProp+tempSourceToParse] = localeItem
						}
					}
				}
			}
			tempSource.Close()
		}
	}

	var lenResult = len(appModelData.ResultSheet)

	var resultSlice []model.LocaleItem = make([]model.LocaleItem, lenResult, lenResult)
	var c int = 0
	for _, lc := range appModelData.ResultSheet {
		resultSlice[c] = lc
		c++
	}

	sort.Slice(resultSlice, func(i int, j int) bool {
		if resultSlice[i].Group != resultSlice[j].Group {
			return resultSlice[i].Group < resultSlice[j].Group
		}
		return resultSlice[i].PropertyName < resultSlice[j].PropertyName
	})

	xlsx := excelize.NewFile()

	var colsLetter rune = 'A'
	var rowIndex uint64 = 1
	var sheetName = "Sheet1"
	xlsx.SetCellValue(sheetName, "A1", "Group")
	xlsx.SetCellValue(sheetName, "B1", "Property")
	colsLetter = 'C'
	for _, tempLang := range appModelData.LangsToManage {
		xlsx.SetCellValue(sheetName, string(colsLetter)+"1", tempLang)
		colsLetter++
	}

	for _, lc := range resultSlice {
		colsLetter = 'A'
		rowIndex++
		xlsx.SetCellValue(sheetName, string(colsLetter)+strconv.FormatUint(rowIndex, 10), lc.Group)
		colsLetter++
		xlsx.SetCellValue(sheetName, string(colsLetter)+strconv.FormatUint(rowIndex, 10), lc.PropertyName)
		colsLetter++

		for _, tempLang := range appModelData.LangsToManage {
			candidateTranslation, _ := lc.GetTranslation(tempLang)
			xlsx.SetCellValue(sheetName, string(colsLetter)+strconv.FormatUint(rowIndex, 10), candidateTranslation)
			colsLetter++
		}
	}

	err := xlsx.SaveAs(params.DestinationPath)
	if apputils.CheckError(err) {
		apputils.StopRunning()
	} else {
		apputils.WriteLog("Excel file saved: "+params.DestinationPath, glg.INFO)
	}
}

func parseSingleLine(line string, sep string) (prop, value string) {
	prop = ""
	value = ""
	if containsSep := strings.Index(line, sep); containsSep > -1 {
		parts := strings.Split(line, sep)
		if len(parts) == 2 {
			prop = strings.TrimSpace(parts[0])
			value = strings.TrimSpace(parts[1])
			return
		}
	}
	return
}

func processT2L(params *model.LocaleConverterParameters) {

	if params.SourcePath != "" {
		excelLangSource, err := excelize.OpenFile(params.SourcePath)
		if apputils.CheckError(err) {
			apputils.StopRunning()
		}

		//retrive langs from col C of first row
		langs := make([]string, 0, 0)

		var sheet string = "Sheet1"
		var colLetter rune = 'C'
		for {
			candidateLang := string(excelLangSource.GetCellValue(sheet, string(colLetter)+"1"))
			if candidateLang == "" {
				break
			}
			langs = append(langs, candidateLang)
			colLetter++
		}

		if len(langs) == 0 {
			apputils.WriteLog("No langs found", glg.ERR)
			apputils.StopRunning()
			return
		}

		rows, err := excelLangSource.Rows(sheet)
		if err != nil {
			fmt.Println(err)
			apputils.StopRunning()
		}

		groupsByLang := make(map[string][]model.LocaleItem)

		for rows.Next() {
			candidateRow := rows.Columns()
			var group string = candidateRow[0]
			var prop string = candidateRow[1]
			translations := make(map[string]string)
			if group != "" && prop != "" {
				for li, lang := range langs {
					candidateTranslation := candidateRow[li+2]
					if candidateTranslation != "" && candidateTranslation != model.NoTranslationPlaceholder {
						translations[lang] = candidateTranslation
					}
				}
				localeItem := model.LocaleItem{prop, group, translations}
				addInGroup(localeItem, groupsByLang)
			}
		}

		for k, m := range groupsByLang {
			var pathLocaleResource string = k
			var lang string = strings.Split(pathLocaleResource, "/")[0]
			var resourceName string = strings.Split(pathLocaleResource, "/")[1]
			var buffer strings.Builder
			for _, litem := range m {
				translation, _ := litem.GetTranslation(lang)
				buffer.WriteString(litem.PropertyName + " = " + translation + "\r")
			}
			var pathLocaleResourceComplete string
			if params.Platform == model.Flex {
				pathLocaleResourceComplete = params.DestinationPath + "/" + pathLocaleResource
			} else if params.Platform == model.Java {
				pathLocaleResourceComplete = params.DestinationPath + "/" + resourceName + "_" + lang + model.ResourceFileExt
			}

			if params.Platform == model.Flex {
				pathError := os.MkdirAll(params.DestinationPath+"/"+lang, os.ModePerm)
				if pathError != nil {
					apputils.WriteLog("Error on creating path "+params.DestinationPath+"/"+lang+": "+err.Error(), glg.WARN)
					continue
				}
			}

			if params.Platform == model.Java {
				pathError := os.MkdirAll(params.DestinationPath, os.ModePerm)
				if pathError != nil {
					apputils.WriteLog("Error on creating path "+params.DestinationPath+": "+err.Error(), glg.WARN)
					continue
				}
			}

			localeFile, err := os.Create(pathLocaleResourceComplete)
			if err != nil {
				apputils.WriteLog("Error on creating "+pathLocaleResourceComplete+": "+err.Error(), glg.WARN)
			} else {
				localeFile.WriteString(buffer.String())
				localeFile.Sync()
				localeFile.Close()
				apputils.WriteLog("Created "+pathLocaleResourceComplete, glg.INFO)
			}
		}
	} else {
		apputils.StopRunning()
	}

}

func addInGroup(localeItem model.LocaleItem, groups map[string][]model.LocaleItem) {
	if localeItem.Group == "Group" {
		return
	}
	for lang, _ := range localeItem.Translations {
		var candidateKey string = lang + "/" + localeItem.Group

		if groups[candidateKey] == nil {
			groups[candidateKey] = make([]model.LocaleItem, 0, 0)
		}

		localeitems := groups[candidateKey]
		localeitems = append(localeitems, localeItem)
		groups[candidateKey] = localeitems
	}
}

func parseFlexSources(sourceFolderItems []os.FileInfo, appModelData *model.AppModelData, params *model.LocaleConverterParameters) {
	//mapping langs reading from 1st level folder
	for i := 0; i < len(sourceFolderItems); i++ {

		var sources os.FileInfo = sourceFolderItems[i]

		if sources.IsDir() {
			//add lang to langs list for build table
			appModelData.LangsToManage = append(appModelData.LangsToManage, sources.Name())
			//open the lang folder
			langDirRef, tempError := os.Open(params.SourcePath + "/" + appModelData.LangsToManage[i])
			if apputils.CheckError(tempError) {
				continue
			}
			//analize how many propoerties boundle lang folder contains
			sourcesForLang, tempError := langDirRef.Readdir(0)
			if apputils.CheckError(tempError) {
				continue
			}

			//analize source inside lang folder
			for ii := 0; ii < len(sourcesForLang); ii++ {
				candidateSourceFileInfo := sourcesForLang[ii]
				//no dir allowed
				if candidateSourceFileInfo.IsDir() == true {
					continue
				}
				//only .properties files
				if propIndex := strings.Index(candidateSourceFileInfo.Name(), ".properties"); propIndex == -1 {
					continue
				}

				//map sources commons in lang folder
				candidateSource := candidateSourceFileInfo.Name()
				addCandidate := true
				for _, prevSourceForLang := range appModelData.SourceToParse {
					if prevSourceForLang == candidateSource {
						addCandidate = false
						break
					}
				}
				if addCandidate {
					appModelData.SourceToParse = append(appModelData.SourceToParse, candidateSource)
				}
			}

			langDirRef.Close()
		}
	}
}

func parseJavaSources(sourceFolderItems []os.FileInfo, appModelData *model.AppModelData, params *model.LocaleConverterParameters) {

	//find out source
	for i := 0; i < len(sourceFolderItems); i++ {

		var source os.FileInfo = sourceFolderItems[i]

		if source.IsDir() {
			continue
		} else {
			sourceName := source.Name()
			//only .properties files
			if propIndex := strings.Index(sourceName, ".properties"); propIndex == -1 {
				continue
			}
			sourceNameParts := strings.Split(strings.TrimSuffix(sourceName, ".properties"), "_")
			if len(sourceNameParts) < 2 {
				continue
			}
			firstPart := sourceNameParts[0]
			addCandidate := true
			for _, prevSourceForLang := range appModelData.SourceToParse {
				if prevSourceForLang == firstPart {
					addCandidate = false
					break
				}
			}
			if addCandidate {
				appModelData.SourceToParse = append(appModelData.SourceToParse, firstPart)
			}

			candidateLang := ""
			for i := 1; i < len(sourceNameParts); i++ {
				candidateLang += "_" + sourceNameParts[i]
			}
			candidateLang = strings.TrimPrefix(candidateLang, "_")

			addCandidate = true
			for _, lang := range appModelData.LangsToManage {
				if lang == candidateLang {
					addCandidate = false
					break
				}
			}
			if addCandidate {
				appModelData.LangsToManage = append(appModelData.LangsToManage, candidateLang)
			}

		}
	}
}
