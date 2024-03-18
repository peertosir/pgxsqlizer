package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	sqlSuffix         = ".sql"
	titleSplitter     = "title:"
	addImportSplitter = "addimport:"
	genFilePostfix    = "_actions_gen.go"
	pkgForGenFiles    = "actionsgen"
)

var (
	stmtTitleRegExp    = regexp.MustCompile(fmt.Sprintf(`^--\s*%s.*`, titleSplitter))
	addImportRegExp    = regexp.MustCompile(fmt.Sprintf(`^--\s*%s.*`, addImportSplitter))
	stmtArgValueRegExp = regexp.MustCompile(`@\S*:\S*@`)

	AvailablePlaceholders = []string{atPlaceholderType, questionPlaceholderType, dollarPlaceholderType}
	AvailableReturnTypes  = []string{mapReturnType, sliceReturnType}
)

type config struct {
	inFolder        string
	outFolder       string
	genPkg          string
	returnType      string
	placeholderType string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.inFolder, "input", ".", "folder to scan for sql files")
	flag.StringVar(&cfg.outFolder, "output", ".", "folder to put files with generated functions")
	flag.StringVar(
		&cfg.placeholderType, "placeholder", "$",
		"placeholder which will be user in sql queries: [@|?|$]. @ => @<string>; ? => ?; $ => $<int>."+
			"'@' is only suitable for 'map' RT. Others are only suitable for 'slice' RT.",
	)
	flag.StringVar(&cfg.genPkg, "genPkg", pkgForGenFiles, "folder/package for generated files")
	flag.StringVar(
		&cfg.returnType, "returnType", "slice",
		"(RT). Generated function return type for stmt params [map|slice]. Default is slice",
	)
	flag.Parse()

	checkValidOptions(cfg)

	result := map[string]TemplateData{}

	err := filepath.Walk(cfg.inFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(info.Name(), sqlSuffix) {
			formattedName := strings.TrimSuffix(info.Name(), sqlSuffix)
			if _, ok := result[formattedName]; !ok {
				data := parseSQLFile(formattedName, path, cfg.returnType, cfg.placeholderType, cfg.genPkg)
				result[formattedName] = data
			}
		}
		return nil
	})

	if err != nil {
		log.Fatalf("cannot walk input dir: error = %v\n", err)
	}

	outputDir := filepath.Join(cfg.outFolder, cfg.genPkg)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		log.Println("directory for actions will be automatically created")
		err = os.Mkdir(outputDir, 0777)
		if err != nil {
			log.Fatalf("cannot create folder for generated files: %v\n", err)
		}
	} else if err != nil {
		log.Fatalf("unexpected os error: %v\n", err)
	}

	for name, data := range result {
		var dataBuffer bytes.Buffer
		data.GenPackage = cfg.genPkg
		err = generatedActionsFileTmpl.Execute(&dataBuffer, data)
		if err != nil {
			log.Fatal(err)
		}
		filePath := filepath.Join(outputDir, name+genFilePostfix)
		err = os.WriteFile(filePath, dataBuffer.Bytes(), 0777)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func parseSQLFile(name, path, returnType, placeholderType, genPkg string) TemplateData {
	genReturnType := NewGenFuncReturnType(returnType)
	result := TemplateData{
		StmtItems:       map[string]StmtItem{},
		ReturnValueType: genReturnType.Signature,
		GenPackage:      genPkg,
	}
	readFile, err := os.Open(path)
	defer readFile.Close()

	if err != nil {
		log.Fatalf("cannot open sql file with path: %v", path)
	}
	scn := bufio.NewScanner(readFile)
	scn.Split(bufio.ScanLines)

	var sqlStmtAccum []string
	var sqlStmtTitle string

	for scn.Scan() {
		scannedText := strings.TrimSpace(scn.Text())

		if len(scannedText) == 0 {
			continue
		} else if addImportRegExp.MatchString(scannedText) {
			result.ImportPackages = append(result.ImportPackages, getImportStmt(scannedText))
		} else if stmtTitleRegExp.MatchString(scannedText) {
			if len(sqlStmtTitle) != 0 {
				result.StmtItems[firstLetterToLower(sqlStmtTitle)] = getStmtItem(
					sqlStmtTitle, strings.Join(sqlStmtAccum, " "),
					placeholderType, genReturnType,
				)
			}
			sqlStmtTitle = getStmtTitle(scannedText)
			sqlStmtAccum = []string{}
		} else {
			sqlStmtAccum = append(sqlStmtAccum, scannedText)
		}
	}
	result.StmtItems[firstLetterToLower(sqlStmtTitle)] = getStmtItem(
		sqlStmtTitle, strings.Join(sqlStmtAccum, " "),
		placeholderType, genReturnType,
	)
	return result
}

func getStmtTitle(inp string) string {
	res := strings.Split(inp, titleSplitter)
	if len(res) != 2 {
		log.Fatalf("sql stmt parse error: %v\n", inp)
	}
	title := strings.TrimSpace(res[1])
	return title
}

func getImportStmt(inp string) string {
	res := strings.Split(inp, addImportSplitter)
	if len(res) != 2 {
		log.Fatalf("import stmt parse error: %v\n", inp)
	}
	return strings.TrimSpace(res[1])
}

func getStmtItem(
	name, stmt, placeholderType string,
	genReturnType *GenFuncReturnType,
) StmtItem {
	result := StmtItem{
		Stmt: stmt,
		Function: GenFunction{
			Name: name,
		},
	}

	funcArgs := map[string]string{}
	returnValueArgs := []string{}

	valuesToReplace := stmtArgValueRegExp.FindAllString(stmt, -1)
	count := 1
	for _, val := range valuesToReplace {
		argName, argType := getArgumentData(val)
		if _, ok := funcArgs[argName]; !ok {
			funcArgs[argName] = fmt.Sprintf(funcArgTmpl, argName, argType)
			if genReturnType.IsMap() {
				returnValueArgs = append(returnValueArgs, fmt.Sprintf(mapArgTmpl, argName, argName))
			} else {
				returnValueArgs = append(returnValueArgs, fmt.Sprintf(sliceArgTmpl, argName))
			}
		} else {
			data := fmt.Sprintf(funcArgTmpl, argName, argType)
			if data != funcArgs[argName] {
				log.Printf("WARNING: different arg types for one var name in stmt: %s\n", name)
			}
		}
		result.Stmt, count = insertPlaceholders(result.Stmt, argName, val, placeholderType, count)
	}
	for _, val := range funcArgs {
		result.Function.Args += " " + val
	}
	result.Function.ReturnValueItems = strings.Join(returnValueArgs, " ")
	return result
}

func getArgumentData(input string) (string, string) {
	argData := strings.Split(strings.ReplaceAll(input, "@", ""), ":")
	if len(argData) != 2 {
		log.Fatalf("argument formatting is incorrect: %s", input)
	}
	return argData[0], argData[1]
}

func insertPlaceholders(stmt, name, value, placeholderType string, count int) (string, int) {
	switch placeholderType {
	case atPlaceholderType:
		return strings.Replace(stmt, value, atPlaceholderType+name, 1), 0
	case dollarPlaceholderType:
		return strings.Replace(stmt, value, dollarPlaceholderType+strconv.Itoa(count), 1), count + 1
	case questionPlaceholderType:
		return strings.Replace(stmt, value, questionPlaceholderType, 1), 0
	default:
		panic("cannot parse placeholder type")
	}
}

func checkValidOptions(cfg config) {
	if !slices.Contains(AvailablePlaceholders, cfg.placeholderType) {
		log.Fatal("placeholder value should be valid => [@|?|$]")
	}

	if !slices.Contains(AvailableReturnTypes, cfg.returnType) {
		log.Fatal("return types value should be valid => [map|slice]")
	}

	if (cfg.returnType == mapReturnType &&
		(cfg.placeholderType == questionPlaceholderType || cfg.returnType == dollarPlaceholderType)) ||
		(cfg.returnType == sliceReturnType && cfg.placeholderType == atPlaceholderType) {
		log.Fatal("incompatible return type and placeholder type")
	}
}

func firstLetterToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}
