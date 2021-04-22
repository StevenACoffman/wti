package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/StevenACoffman/wti/pkg/atlassian"
)

const (
	// exitFail is the exit code if the program
	// fails.
	exitFail = 1
	// exitSuccess is the exit code if the program succeeds
	exitSuccess = 0
)

func main() {
	errlog := log.New(os.Stderr, "", 0)

	// pass all arguments without the executable name
	if err := wti(); err != nil {
		errlog.Printf("%+v\n", err)
		os.Exit(exitFail)
	} else {
		os.Exit(exitSuccess)
	}
}

func wti() error {
	outlog := log.New(os.Stdout, "", 0)

	ticket := ""
	// getArgs will exclude the binary name and flags
	args := getArgs()
	ticket = getTicketFromArgs(args)
	var omitTitle, omitDescription bool
	flag.BoolVar(&omitTitle, "no-title", false, "Do Not Print Title")
	flag.BoolVar(&omitDescription, "no-description", false, "Do Not Print Description")

	flagErr := flag.CommandLine.Parse(getFlags())
	// check if we get a flag parse Error (e.g. missing required or
	// unrecognized)
	if flagErr != nil {
		fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		return errors.New("invalid or missing flag")
	}
	jiraConfig := atlassian.ConfigureJira()
	jiraClient := atlassian.GetJIRAClient(jiraConfig)
	jiraIssue, issueErr := atlassian.GetIssue(jiraClient, ticket)

	if issueErr == nil {
		if !omitTitle {
			outlog.Printf("%s - %s\n\n", jiraIssue.Key, jiraIssue.Fields.Summary)
		}
		if !omitDescription {
			outlog.Println(atlassian.JiraMarkupToGithubMarkdown(jiraClient, jiraIssue.Fields.Description))
		}
	}

	return issueErr
}

func getTicketFromArgs(args []string) string {
	if len(args) < 1 {
		return ""
	}
	return args[0]
}

// flags but no args
func getFlags() []string {
	var args []string
	for _, arg := range os.Args[1:] {
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			args = append(args, arg)
		}
	}
	return args
}

// getArgs skips binary name, and no flags please
func getArgs() []string {
	argSkip := 1 // skip binary name if there was one
	if len(os.Args) <= argSkip {
		return []string{}
	}

	var args []string
	for _, arg := range os.Args[argSkip:] {
		// skip flags
		if strings.HasPrefix(arg, "--") || strings.HasPrefix(arg, "-") {
			continue
		}
		args = append(args, arg)
	}
	return args
}
