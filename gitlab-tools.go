package main

import (
	"flag"
	"log"

	"github.com/xanzy/go-gitlab"
)

var info = log.Println

type issue = gitlab.CreateIssueOptions
type snippet = gitlab.CreateSnippetOptions

// OpenGitlab ...
func OpenGitlab(base, token string) *gitlab.Client {
	git := gitlab.NewClient(nil, token)
	git.SetBaseURL(base)
	return git
}

func copyIssues() {
	exists := TargetIssues(config.target, config.tToken)

	target := OpenGitlab(config.target, config.tToken)
	for is := range Issues(config.source, config.sToken) {
		if _, ok := exists[is.Title]; !ok {
			nissue := issue{
				Title:        &is.Title,
				Description:  &is.Description,
				Confidential: &is.Confidential,
				Labels:       is.Labels,
			}
			is, _, err := target.Issues.CreateIssue(config.pid, &nissue)
			panice(err)
			info(is.ID, is.Title)
		}
	}
}

func copySnippets() {
	exists := TargetSnippetFileNames(config.target, config.tToken)

	source := OpenGitlab(config.source, config.sToken)
	target := OpenGitlab(config.target, config.tToken)
	for s := range Snippets(config.source, config.sToken) {
		if _, ok := exists[s.FileName]; !ok {
			if cont, _, err := source.Snippets.SnippetContent(s.ID); err == nil {
				cont := string(cont)
				cs := snippet{
					Title:       &s.Title,
					FileName:    &s.FileName,
					Description: &s.Description,
					Visibility:  gitlab.Visibility(gitlab.PublicVisibility),
					Content:     &cont,
				}
				s, _, err = target.Snippets.CreateSnippet(&cs)
				panice(err)
				info(s.ID, s.FileName)
			}
		}

	}
}

func main() {
	if config.copyIssues {
		copyIssues()
	}
	if config.copySnippets {
		copySnippets()
	}
}

// TargetSnippetFileNames ...
func TargetSnippetFileNames(base, token string) map[string]struct{} {
	rets := map[string]struct{}{}
	for s := range Snippets(base, token) {
		rets[s.FileName] = struct{}{}
	}
	return rets
}

// Snippets ...
func Snippets(base, token string) chan *gitlab.Snippet {
	rets := make(chan *gitlab.Snippet, 100)

	go func() {
		defer close(rets)
		git := OpenGitlab(base, token)
		opt := gitlab.ListSnippetsOptions{PerPage: 100}
		snippets, _, err := git.Snippets.ListSnippets(&opt)
		for err == nil && len(snippets) > 0 {
			for _, s := range snippets {
				rets <- s
			}
			opt.Page = opt.Page + 1
			snippets, _, err = git.Snippets.ListSnippets(&opt)
		}

	}()
	return rets
}

// TargetIssues ...
func TargetIssues(base, token string) map[string]struct{} {
	rets := map[string]struct{}{}
	for issue := range Issues(base, token) {
		rets[issue.Title] = struct{}{}
	}
	return rets
}

// Issues ...
func Issues(base, token string) chan *gitlab.Issue {

	rets := make(chan *gitlab.Issue, 100)
	go func() {
		defer close(rets)
		git := OpenGitlab(base, token)
		opt := gitlab.ListOptions{PerPage: 1000}
		issues, _, err := git.Issues.ListIssues(&gitlab.ListIssuesOptions{ListOptions: opt})
		for err == nil && len(issues) > 0 {
			for _, issue := range issues {
				rets <- issue
			}
			opt.Page = opt.Page + 1
			issues, _, err = git.Issues.ListIssues(&gitlab.ListIssuesOptions{ListOptions: opt})
		}

	}()
	return rets
}

func init() {
	flag.StringVar(&config.source, "source", "http://172.17.12.184/", "")
	flag.StringVar(&config.target, "target", "http://10.1.7.196/", "")
	flag.StringVar(&config.sToken, "source-token", "6z6AdxxA2rg1puyNYQsw", "")
	flag.StringVar(&config.tToken, "target-token", "j9s3YZXnamsft5Khqdeq", "")
	flag.StringVar(&config.pid, "project-id", "hunks/xiwangzuang", "target project id")
	flag.BoolVar(&config.copyIssues, "copy-issues", false, "")
	flag.BoolVar(&config.copySnippets, "copy-snippets", false, "")

	flag.Parse()
}

var config struct {
	source       string
	target       string
	sToken       string
	tToken       string
	pid          string
	copyIssues   bool
	copySnippets bool
}

func panice(err error) {
	if err != nil {
		panic(err)
	}
}
