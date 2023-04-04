package main

import (
	"bufio"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func walkFolderForFilesToReplaceAllFileLinks(root, ext string) error {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ext {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, file := range files {
		// fmt.Printf("==%s==\n", file)
		err := processFileForLinks(root, file, files)
		if err != nil {
			return err
		}
	}

	return err
}

// processFileForLinks searches and replaces markdown links that refer to a
// file of identical name also in the folder heirarchy somewhere.
func processFileForLinks(root, path string, files []string) error {
	dirtyFlag := false
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	fileinfo, err := file.Stat()
	if fileinfo.Size() == 0 {
		file.Close()
		return nil
	}
	if err != nil {
		return err
	}

	defer file.Close()

	currentFileFolder := filepath.Dir(path)
	scanner := bufio.NewScanner(file)
	newLines := []string{}

	for scanner.Scan() {
		line := scanner.Text()
		// look for markdown link [...](...) pattern:
		linkRegex := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)

		linkMatchPositions := linkRegex.FindAllStringSubmatchIndex(line, -1)
		if linkMatchPositions != nil {
			dirtyFlag = true

			var replacementline string
			matchStart := 0

			for x, match := range linkMatchPositions {
				replacementline += line[matchStart:linkMatchPositions[x][0]]
				linkText := line[match[2]:match[3]] //link ref
				linkPath := line[match[4]:match[5]] //link to .md

				linkFilename := filepath.Base(linkPath)
				linkFilepath := getLinkFilepathFromFileList(files, linkFilename)
				if linkFilepath != "" {
					//if the linked file is found within the list of files from of our folder walk, rewrite the link.

					// build a path to the linked file from the currently open file
					linkPathRelToCurrentFile, err := filepath.Rel(currentFileFolder, linkFilepath)
					if err != nil {
						return err
					}
					if filepath.Dir(linkFilepath) == currentFileFolder {
						//[current file] and [file target of the current link] are in the same folder: link with no path prefix.
						// fmt.Printf("Test result same folder:[%s](%s)\n", linkText, linkFilename)
						replacementline += "[" + linkText + "](" + linkFilename + ")"
					} else {
						//[current file] and [file target of the current link] are in different folders: link with relative path prefix.
						// fmt.Printf("Test result different folder:[%s](%s)\n", linkText, linkPathRelToCurrentFile) //, linkFilename)
						replacementline += "[" + linkText + "](" + linkPathRelToCurrentFile + ")"
					}
				} else {
					replacementline += "[" + linkText + "](" + linkPath + ")"
				}

				// Update the start index for the next match (could be multiple links per line)
				matchStart = match[5] + 1
			}
			if len(line)-matchStart > 0 {
				replacementline += line[matchStart:]
			}
			// fmt.Printf("replacementline: %s\n", replacementline)
			line = replacementline
		}
		newLines = append(newLines, line)
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	newContent := strings.Join(newLines, "\n") + "\n"
	if dirtyFlag {
		err = ioutil.WriteFile(path, []byte(newContent), 0644)
		if err != nil {
			return err
		}
	}

	return nil
}

func getLinkFilepathFromFileList(files []string, filename string) string {
	for _, file := range files {
		if filepath.Base(file) == filepath.Base(filename) {
			return file
		}
	}
	return ""
}
