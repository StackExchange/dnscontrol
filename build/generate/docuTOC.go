package main

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func generateDocuTOC(folderPath string, targetFile string, onlyThisPath string, startMarker string, endMarker string) error {
	if folderPath == "" {
		return fmt.Errorf("empty docutoc path")
	}

	var exclusivePath string
	if onlyThisPath != "" {
		// if onlyThisPath is provided, build a list of files exclusively in this (sub)folder
		exclusivePath = string(filepath.Separator) + onlyThisPath
	}
	// Find all the markdown files in the specified folder and its subfolders.
	markdownFiles, err := findMarkdownFiles(folderPath + exclusivePath)
	if err != nil {
		return err
	}

	//First sort by folders, then by filename.
	sort.SliceStable(markdownFiles, func(i, j int) bool {
		if filepath.Dir(markdownFiles[i]) == filepath.Dir(markdownFiles[j]) {
			return strings.ToLower(markdownFiles[i]) < strings.ToLower(markdownFiles[j])
		} else {
			return filepath.Dir(strings.ToLower(markdownFiles[i])) < filepath.Dir(strings.ToLower(markdownFiles[j]))
		}
	})

	// Create the table of contents.
	toc := generateTableOfContents(folderPath, markdownFiles)

	err = replaceTextBetweenMarkers(filepath.Join(folderPath, targetFile), startMarker, endMarker, toc)
	if err != nil {
		return err
	}

	return nil
}

// func stringInSlice(a string, list []string) bool {
//     for _, b := range list {
//         if b == a {
//             return true
//         }
//     }
//     return false
// }

func generateTableOfContents(folderPath string, markdownFiles []string) string {
	var toc strings.Builder
	currentFolder := ""

	// skip over these root entries (dont print these "#" headings)
	rootFolderExceptions := []string{
		"documentation",
	}

	// dont print these folder names as bullets (which lack a link)
	folderExceptions := []string{
		"providers",
	}
	// dont print these file names as bullets
	fileExceptions := []string{
		"index.md",
		"summary.md",
		"ignore-me.md",
	}

	caser := cases.Title(language.Und, cases.NoLower)

	for _, file := range markdownFiles {
		//depthCount is folder depth for toc indentation, minus one for docu folder
		depthCount := strings.Count(file, string(filepath.Separator)) - 1
		filename := filepath.Base(file)

		fileFolder := filepath.Dir(file)
		if fileFolder != currentFolder {
			// we are in a new folder

			// hop over these entries altogether
			if stringInSlice(strings.ToLower(fileFolder), rootFolderExceptions) {
				continue
			}
			currentFolder = fileFolder
			folderName := filepath.Base(currentFolder)

			// if we're in an "exception" folder, deeper than a heading "#", skip printing it
			// this has the effect of putting subentries under an entry that is already a link,
			// without printing an entry to represent the folder name that is not a link
			// e.g. provider md files are all links, under the provider.md file which is a link
			if stringInSlice(strings.ToLower(folderName), folderExceptions) && depthCount > 1 {
				continue
			} else {
				if depthCount > 1 {
					// if we're deeper in heirarchy, print an indented bullet "*" to add to bullet heirarchy
					toc.WriteString(strings.Repeat("  ", depthCount-1) + "* ")
				} else {
					// If we're in folder root, just print an unindented heading "#"
					toc.WriteString("\n## ")
				}
				// Captalize folder names, replace underscores with spaces for # headings
				toc.WriteString(strings.TrimSpace(caser.String(strings.ReplaceAll(folderName, "_", " "))) + "\n")
			}
		}
		//if the file is an exception listed above, skip it.
		if stringInSlice(strings.ToLower(filename), fileExceptions) {
			continue
		}

		// naming exceptions - function names shall retain "_"
		displayfilename := strings.TrimSuffix(filename, filepath.Ext(filename))
		if !strings.Contains(file, "functions") {
			displayfilename = strings.TrimSpace(caser.String(strings.ReplaceAll(displayfilename, "_", " ")))
		}

		// print the filename as a bullet, and as a [link](hyperlink)
		toc.WriteString(strings.Repeat("  ", depthCount))
		toc.WriteString("* [" + displayfilename + "](" + filepath.Join(".", strings.ReplaceAll(file, folderPath, "")) + ")\n")
	}
	return toc.String()
}

// replaceTextBetweenMarkers inserts the generated table of contents between the two markers in the specified markdown file.
func replaceTextBetweenMarkers(targetFile, startMarker, endMarker, newcontent string) error {
	// Read the contents of the markdown file into memory.
	input, err := os.ReadFile(targetFile)
	if err != nil {
		return err
	}

	// Find the starting and ending positions of the table of contents markers.
	startPos := strings.Index(string(input), startMarker)
	if startPos == -1 {
		return fmt.Errorf("could not find start marker %q in file %q", startMarker, targetFile)
	}
	endPos := strings.Index(string(input), endMarker)
	if endPos == -1 {
		return fmt.Errorf("could not find end marker %q in file %q", endMarker, targetFile)
	}

	// Construct the new contents of the markdown file with the updated table of contents.
	output := string(input[:startPos+len(startMarker)]) + newcontent + string(input[endPos:])

	// Write the updated contents to the markdown file.
	err = os.WriteFile(targetFile, []byte(output), 0644)
	if err != nil {
		return err
	}

	return nil
}

// findMarkdownFiles returns a list of all the markdown files in the specified folder and its subfolders.
func findMarkdownFiles(folderPath string) ([]string, error) {
	markdownFiles := make([]string, 0)
	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".md" {
			markdownFiles = append(markdownFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return markdownFiles, err
}
