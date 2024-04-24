package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	out := os.Stdout
	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}
	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)
	if err != nil {
		panic(err.Error())
	}
}

func dirTree(out io.StringWriter, path string, printFiles bool) error {
	if printFiles {
		return dirTreeWithFiles(out, path, 0, "")
	}
	return dirTreeWithoutFiles(out, path, 0, "")
}

func dirTreeWithoutFiles(out io.StringWriter, path string, depth int, tmp string) error {
	listFiles, err := os.ReadDir(path)
	if err != nil {
		panic("Error read dir: " + path)
	}
	countDir := 0
	for _, v := range listFiles {
		if v.IsDir() {
			countDir++
		}
	}
	curDirCount := 0
	for index, file := range listFiles {
		if file.IsDir() {
			curDirCount++
		}
		if strings.HasPrefix(file.Name(), ".") || strings.Contains(file.Name(), "_debug") || !file.IsDir() {
			continue
		}

		out.WriteString(tmp)
		if index == len(listFiles)-1 || curDirCount == countDir {
			out.WriteString("└───")
		} else {
			out.WriteString("├───")
		}

		out.WriteString(file.Name())
		if file.IsDir() {
			out.WriteString("\n")
			if index != len(listFiles)-1 && countDir > curDirCount {
				tmp += "│\t"
			} else {
				tmp += "\t"
			}
			depth++

			dirTreeWithoutFiles(out, path+string(os.PathSeparator)+file.Name(), depth, tmp)
			if strings.HasSuffix(tmp, "│\t") {
				tmp, _ = strings.CutSuffix(tmp, "│\t")
			} else {
				tmp, _ = strings.CutSuffix(tmp, "\t")
			}
			depth--
		}
	}
	return nil
}

func dirTreeWithFiles(out io.StringWriter, path string, depth int, tmp string) error {
	listFiles, err := os.ReadDir(path)
	if err != nil {
		panic("Error read dir: " + path)
	}

	for index, file := range listFiles {
		if strings.HasPrefix(file.Name(), ".") || strings.Contains(file.Name(), "_debug") {
			continue
		}

		out.WriteString(tmp)
		if index == len(listFiles)-1 {
			out.WriteString("└───")
		} else {
			out.WriteString("├───")
		}

		out.WriteString(file.Name())
		if file.IsDir() {
			out.WriteString("\n")
			if index != len(listFiles)-1 {
				tmp += "│\t"
			} else {
				tmp += "\t"
			}
			depth++

			dirTreeWithFiles(out, path+string(os.PathSeparator)+file.Name(), depth, tmp)
			if strings.HasSuffix(tmp, "│\t") {
				tmp, _ = strings.CutSuffix(tmp, "│\t")
			} else {
				tmp, _ = strings.CutSuffix(tmp, "\t")
			}
			depth--
		} else {
			infoFile, err := file.Info()
			if err != nil {
				panic("Error get info file" + file.Name())
			}
			size := ""
			if infoFile.Size() == 0 {
				size = "empty"
			} else {
				size = fmt.Sprint(infoFile.Size()) + "b"
			}
			out.WriteString(" (" + size + ")\n")
		}
	}
	return nil
}
