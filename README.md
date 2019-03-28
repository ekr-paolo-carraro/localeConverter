# localeConverter
Utility to convert locale resource in Excel spreadsheet for better compare and compile lang translations.
Currently Java and Flex locale resource are managed.

![locale converter image](/localeConverter-image.png)

## Usage
To convert a locale folder source in a spreadsheet
```
./localeConverter -action=L2T -platform=Flex -source=C:/Users/name/workspace/app/locale -dest=C:/Users/name/Documents/appLocale.xlsx
```

To convert a previous spreadsheet created with localeConverter in locale sources
```
./localeConverter -action=T2L -platform=Flex -source=C:/Users/name/Documents/appLocale.xlsx -dest=C:/Users/name/workspace/app/locale
```
To print help
```
./localeConverter --help
```
