package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	sqlSuffix         = ".sql"
	titleSplitter     = "title:"
	funcGenTmpl       = "\nfunc %s(%s) (string, pgx.NamesArgs) {\n\treturn %sStmtMap[\"%s\"], pgx.NamedArgs{\n%s\n\t}\n}\n"
	packageHeaderTmpl = "package pgxsqlizer\n\n"
	genFilePostfix    = "_actions.go"
	stmtMapTmpl       = "\t\"%s\": \"%s\",\n"
	funcArgTmpl       = "%s %s"
	namedArgTmpl      = "\t\t\"%s\": %s,"
	mapStmtTmpl       = "var %sStmtMap = map[string]string{\n%s}\n"
	importPgxStmt     = "import \"github.com/jackc/pgx/v5\"\n\n"
)

var (
	stmtTitleRegExp = regexp.MustCompile(fmt.Sprintf(`^--\s*%s.*`, titleSplitter))
	stmtArgValue    = regexp.MustCompile(`@\S*:\S*@`)
)

type config struct {
	inFolder  string
	outFolder string
}

func main() {
	var cfg config
	flag.StringVar(&cfg.inFolder, "input", "", "folder to scan for sql files")
	flag.StringVar(&cfg.outFolder, "output", "", "folder to put files with generated functions")
	flag.Parse()

	if len(strings.TrimSpace(cfg.inFolder)) == 0 {
		panic("cannot find folder for sql scan")
	}

	result := map[string]map[string]string{}

	err := filepath.Walk(cfg.inFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return err
		}

		if strings.HasSuffix(info.Name(), sqlSuffix) {
			formattedName := strings.TrimSuffix(info.Name(), sqlSuffix)
			if _, ok := result[formattedName]; !ok {
				data, err := parseSQLFile(path)
				if err != nil {
					fmt.Printf("cannot parse SQL File with path = %s: %v\n", path, err)
					return err
				}
				result[formattedName] = data
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("cannot walk input dir: error = %v\n", err)
		panic(err)
	}

	outputDir := filepath.Join(cfg.outFolder, "pgxsqlizer")
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0777)
		if err != nil {
			fmt.Printf("cannot create folder for generated files: %v\n", err)
			panic(err)
		}
	} else if err != nil {
		fmt.Printf("unexpected os error: %v\n", err)
		panic(err)
	}

	for fileName, queries := range result {
		generateFile(outputDir, fileName, queries)
	}
}

func parseSQLFile(path string) (map[string]string, error) {
	result := map[string]string{}
	readFile, err := os.Open(path)

	if err != nil {
		fmt.Printf("cannot open sql file with path: %v", path)
		return nil, err
	}
	scn := bufio.NewScanner(readFile)
	scn.Split(bufio.ScanLines)

	var sqlStmtAccum []string
	var sqlStmtTitle string

	for scn.Scan() {
		scannedText := strings.TrimSpace(scn.Text())

		if len(scannedText) == 0 {
			continue
		} else if stmtTitleRegExp.MatchString(scannedText) {
			if len(sqlStmtTitle) != 0 {
				result[sqlStmtTitle] = strings.Join(sqlStmtAccum, " ")
			}
			sqlStmtTitle = getStmtTitle(scannedText)
			sqlStmtAccum = []string{}
		} else {
			sqlStmtAccum = append(sqlStmtAccum, scannedText)
		}
	}

	result[sqlStmtTitle] = strings.Join(sqlStmtAccum, " ")
	readFile.Close()
	return result, nil
}

func getStmtTitle(inp string) string {
	res := strings.Split(inp, titleSplitter)
	if len(res) != 2 {
		panic("sql stmt parse error: " + inp)
	}
	return strings.TrimSpace(res[1])
}

func generateFile(folderPath, filePrefix string, queries map[string]string) {
	var stmtMapInnerData string
	var functionsCode string
	// package header
	resultData := packageHeaderTmpl + importPgxStmt

	// traverse queries for sql file
	for key, val := range queries {
		funcName := key
		funcQuery := val
		funcArgsWithTypes := []string{}
		namedArgsItems := []string{}

		valuesToReplace := stmtArgValue.FindAllString(val, -1)

		for _, v := range valuesToReplace {
			argWithType := strings.Split(strings.ReplaceAll(v, "@", ""), ":")
			funcArgsWithTypes = append(funcArgsWithTypes, fmt.Sprintf(funcArgTmpl, argWithType[0], argWithType[1]))
			namedArgsItems = append(namedArgsItems, fmt.Sprintf(namedArgTmpl, argWithType[0], argWithType[0]))
			funcQuery = strings.Replace(funcQuery, v, "@"+argWithType[0], 1)
		}

		stmtMapInnerData += fmt.Sprintf(stmtMapTmpl, funcName, funcQuery)
		functionsCode += fmt.Sprintf(funcGenTmpl, key, strings.Join(funcArgsWithTypes, ", "), filePrefix, funcName, strings.Join(namedArgsItems, "\n"))
	}

	resultData += fmt.Sprintf(mapStmtTmpl, filePrefix, stmtMapInnerData)
	resultData += functionsCode
	filePath := filepath.Join(folderPath, filePrefix+genFilePostfix)
	err := os.WriteFile(filePath, []byte(resultData), 0777)
	if err != nil {
		fmt.Printf("cannot write to file with path: %s", filePath)
		panic(err)
	}
}
