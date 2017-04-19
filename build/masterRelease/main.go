package main

import (
	"context"
	"log"
	"os"
	"strings"

	"fmt"
	"time"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	owner = "StackExchange"
	repo  = "dnscontrol"
	tag   = "latest"
)

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

var bg = context.Background

var files = []string{"dnscontrol.exe", "dnscontrol-Linux", "dnscontrol-Darwin"}

func main() {

	tok := os.Getenv("GITHUB_ACCESS_TOKEN")
	if tok == "" {
		log.Fatal("$GITHUB_ACCESS_TOKEN required")
	}
	c := github.NewClient(oauth2.NewClient(bg(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: tok})))

	log.Println("Getting release info")
	rel, _, err := c.Repositories.GetReleaseByTag(bg(), owner, repo, tag)
	check(err)

	for _, f := range files {
		log.Printf("--- %s", f)

		var found github.ReleaseAsset
		var exists bool
		var foundOld bool
		for _, ass := range rel.Assets {
			if ass.GetName() == f {
				exists = true
				found = ass
			}
			if ass.GetName() == f+".old" {
				foundOld = true
			}
		}

		if foundOld {
			log.Fatalf("%s.old was already found. Previous deploy likely failed. Please check and manually delete.", f)
		}
		if exists {
			oldN := found.GetName()
			n := oldN + ".old"
			found.Name = &n
			log.Printf("Renaming old asset %s(%d) to %s", oldN, found.GetID(), found.GetName())
			_, _, err = c.Repositories.EditReleaseAsset(bg(), owner, repo, found.GetID(), &found)
			check(err)
		}

		log.Printf("Uploading new file %s", f)
		upOpts := &github.UploadOptions{}
		upOpts.Name = f
		f, err := os.Open(f)
		check(err)
		_, _, err = c.Repositories.UploadReleaseAsset(bg(), owner, repo, rel.GetID(), upOpts, f)
		check(err)

		if exists {
			log.Println("Deleting old asset")
			_, err = c.Repositories.DeleteReleaseAsset(bg(), owner, repo, found.GetID())
			check(err)
		}
	}

	log.Println("Editing release body")
	body := strings.TrimSpace(rel.GetBody())
	lines := strings.Split(body, "\n")
	last := lines[len(lines)-1]
	if !strings.HasPrefix(last, "Last updated:") {
		log.Fatal("Release body is not what I expected. Abort!")
	}
	last = fmt.Sprintf("Last updated: %s", time.Now().Format("Mon Jan 2 2006 @15:04 MST"))
	lines[len(lines)-1] = last
	body = strings.Join(lines, "\n")
	rel.Body = &body
	c.Repositories.EditRelease(bg(), owner, repo, rel.GetID(), rel)

	log.Println("DONE")
}
