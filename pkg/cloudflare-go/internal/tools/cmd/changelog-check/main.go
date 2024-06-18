package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const (
	changelogEntryFileFormat      = ".changelog/%d.txt"
	changelogProcessDocumentation = "https://github.com/cloudflare/cloudflare-go/blob/master/docs/changelog-process.md"
	changelogDetectedMessage      = "changelog detected :white_check_mark:"
)

var (
	changelogEntryPresent        = false
	successMessageAlreadyPresent = false
)

func getSkipLabels() []string {
	return []string{"workflow/skip-changelog-entry", "dependencies"}
}

func main() {
	ctx := context.Background()
	if len(os.Args) < 2 {
		log.Fatalf("Usage: changelog-check PR#\n")
	}
	pr := os.Args[1]
	prNo, err := strconv.Atoi(pr)
	if err != nil {
		log.Fatalf("error parsing PR %q as a number: %s", pr, err)
	}

	owner := os.Getenv("GITHUB_OWNER")
	repo := os.Getenv("GITHUB_REPO")
	token := os.Getenv("GITHUB_TOKEN")

	if owner == "" {
		log.Fatalf("GITHUB_OWNER not set")
	}

	if repo == "" {
		log.Fatalf("GITHUB_REPO not set")
	}

	if token == "" {
		log.Fatalf("GITHUB_TOKEN not set")
	}

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	pullRequest, _, err := client.PullRequests.Get(ctx, owner, repo, prNo)
	if err != nil {
		log.Fatalf("error retrieving pull request %s/%s#%d: %s", owner, repo, prNo, err)
	}

	for _, label := range pullRequest.Labels {
		for _, skipLabel := range getSkipLabels() {
			if label.GetName() == skipLabel {
				log.Printf("%s label found, exiting as changelog is not required\n", label.GetName())
				os.Exit(0)
			}
		}
	}

	files, _, _ := client.PullRequests.ListFiles(ctx, owner, repo, prNo, &github.ListOptions{})
	if err != nil {
		log.Fatalf("error retrieving files on pull request %s/%s#%d: %s", owner, repo, prNo, err)
	}

	for _, file := range files {
		if file.GetFilename() == fmt.Sprintf(changelogEntryFileFormat, prNo) {
			changelogEntryPresent = true
		}
	}

	comments, _, _ := client.Issues.ListComments(ctx, owner, repo, prNo, &github.IssueListCommentsOptions{})
	for _, comment := range comments {
		if strings.Contains(comment.GetBody(), "no changelog entry is attached to") {
			if changelogEntryPresent {
				client.Issues.EditComment(ctx, owner, repo, *comment.ID, &github.IssueComment{
					Body: cloudflare.StringPtr(changelogDetectedMessage),
				})
				os.Exit(0)
			}
			log.Println("no change in status of changelog checks; exiting")
			os.Exit(1)
		}

		if strings.Contains(comment.GetBody(), changelogDetectedMessage) {
			successMessageAlreadyPresent = true
		}
	}

	if changelogEntryPresent {
		if !successMessageAlreadyPresent {
			_, _, _ = client.Issues.CreateComment(ctx, owner, repo, prNo, &github.IssueComment{
				Body: cloudflare.StringPtr(changelogDetectedMessage),
			})
		}
		log.Printf("changelog found for %d, skipping remainder of checks\n", prNo)
		os.Exit(0)
	}

	body := "Oops! It looks like no changelog entry is attached to" +
		" this PR. Please include a release note as described in " +
		changelogProcessDocumentation + ".\n\nExample: " +
		"\n\n~~~\n```release-note:TYPE\nRelease note" +
		"\n```\n~~~\n\n" +
		"If you do not require a release note to be included, please add the `workflow/skip-changelog-entry` label."

	_, _, err = client.Issues.CreateComment(ctx, owner, repo, prNo, &github.IssueComment{
		Body: &body,
	})

	if err != nil {
		log.Fatalf("failed to comment on pull request %s/%s#%d: %s", owner, repo, prNo, err)
	}

	os.Exit(1)
}
