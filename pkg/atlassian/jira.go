package atlassian

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"

	jira "github.com/andygrunwald/go-jira"
	"github.com/sethgrid/pester"
)

func GetJIRAClient(config *Config) *jira.Client {
	pesterClient := pester.New()
	pesterClient.Backoff = pester.ExponentialJitterBackoff
	pesterClient.MaxRetries = 4
	pesterClient.KeepLog = false // Cannot both retain logs and have loghook

	tp := jira.BasicAuthTransport{
		Username: config.User,
		Password: config.Token,
	}
	pesterClient.Transport = tp.Transport
	jiraClient, err := jira.NewClient(pesterClient, config.Host)
	if err != nil {
		log.Fatalf("unable to create new JIRA client. %v", err)
	}
	return jiraClient
}

// GetIssue checks if issue exists in the JIRA instance.
// If not an error will be returned.
func GetIssue(jiraClient *jira.Client, issue string) (*jira.Issue, error) {
	jiraIssue, resp, err := jiraClient.Issue.Get(issue, nil)
	if c := resp.StatusCode; err != nil || (c < 200 || c > 299) {
		return nil, fmt.Errorf("JIRA Request for issue %s returned %s (%d)", issue, resp.Status, resp.StatusCode)
	}
	return jiraIssue, nil
}

// Jiration - convenience for Jira Markup to Github Markdown translation rule
type Jiration struct {
	re   *regexp.Regexp
	repl interface{}
}

// JiraToMD - This uses some regular expressions to make a reasonable translation
// from Jira Markup to Github Markdown. It is not a complete PEG so it will break down
// especially for more complicated nested formatting (lists inside of lists)
func JiraToMD(str string) string {
	jirations := []Jiration{
		{ // UnOrdered Lists
			re: regexp.MustCompile(`(?m)^[ \t]*(\*+)\s+`),
			repl: func(groups []string) string {
				_, stars := groups[0], groups[1]
				return strings.Repeat("  ", len(stars)-1) + "* "
			},
		},
		{ // Ordered Lists
			re: regexp.MustCompile(`(?m)^[ \t]*(#+)\s+`),
			repl: func(groups []string) string {
				_, nums := groups[0], groups[1]
				return strings.Repeat("  ", len(nums)-1) + "1. "
			},
		},
		{ // Headers 1-6
			re: regexp.MustCompile(`(?m)^h([0-6])\.(.*)$`),
			repl: func(groups []string) string {
				_, level, content := groups[0], groups[1], groups[2]
				i, _ := strconv.Atoi(level)
				return strings.Repeat("#", i) + content
			},
		},
		{ // Bold
			re:   regexp.MustCompile(`\*(\S.*)\*`),
			repl: "**$1**",
		},
		{ // Italic
			re:   regexp.MustCompile(`\_(\S.*)\_`),
			repl: "*$1*",
		},
		{ // Monospaced text
			re:   regexp.MustCompile(`\{\{([^}]+)\}\}`),
			repl: "`$1`",
		},
		{ // Citations (buggy)
			re:   regexp.MustCompile(`\?\?((?:.[^?]|[^?].)+)\?\?`),
			repl: "<cite>$1</cite>",
		},
		{ // Inserts
			re:   regexp.MustCompile(`\+([^+]*)\+`),
			repl: "<ins>$1</ins>",
		},
		{ // Superscript
			re:   regexp.MustCompile(`\^([^^]*)\^`),
			repl: "<sup>$1</sup>",
		},
		{ // Subscript
			re:   regexp.MustCompile(`~([^~]*)~`),
			repl: "<sub>$1</sub>",
		},
		{ // Strikethrough
			re:   regexp.MustCompile(`(\s+)-(\S+.*?\S)-(\s+)`),
			repl: "$1~~$2~~$3",
		},
		{ // Code Block
			re: regexp.MustCompile(
				`\{code(:([a-z]+))?([:|]?(title|borderStyle|borderColor|borderWidth|bgColor|titleBGColor)=.+?)*\}`,
			),
			repl: "```$2",
		},
		{ // Code Block End
			re:   regexp.MustCompile(`{code}`),
			repl: "```",
		},
		{ // Pre-formatted text
			re:   regexp.MustCompile(`{noformat}`),
			repl: "```",
		},
		{ // Un-named Links
			re:   regexp.MustCompile(`(?U)\[([^|]+)\]`),
			repl: "<$1>",
		},
		{ // Images
			re:   regexp.MustCompile(`!(.+)!`),
			repl: "![]($1)",
		},
		{ // Named Links
			re:   regexp.MustCompile(`\[(.+?)\|(.+)\]`),
			repl: "[$1]($2)",
		},
		{ // Single Paragraph Blockquote
			re:   regexp.MustCompile(`(?m)^bq\.\s+`),
			repl: "> ",
		},
		{ // Remove color: unsupported in md
			re:   regexp.MustCompile(`(?m)\{color:[^}]+\}(.*)\{color\}`),
			repl: "$1",
		},
		{ // panel into table
			re: regexp.MustCompile(
				`(?m)\{panel:title=([^}]*)\}\n?(.*?)\n?\{panel\}`,
			),
			repl: "\n| $1 |\n| --- |\n| $2 |",
		},
		{ // table header
			re: regexp.MustCompile(`(?m)^[ \t]*((?:\|\|.*?)+\|\|)[ \t]*$`),
			repl: func(groups []string) string {
				_, headers := groups[0], groups[1]
				reBarred := regexp.MustCompile(`\|\|`)

				singleBarred := reBarred.ReplaceAllString(headers, "|")
				fillerRe := regexp.MustCompile(`\|[^|]+`)
				return "\n" + singleBarred + "\n" + fillerRe.ReplaceAllString(
					singleBarred,
					"| --- ",
				)
			},
		},
		{ // remove leading-space of table headers and rows
			re:   regexp.MustCompile(`(?m)^[ \t]*\|`),
			repl: "|",
		},
	}
	for _, jiration := range jirations {
		switch v := jiration.repl.(type) {
		case string:
			str = jiration.re.ReplaceAllString(str, v)
		case func([]string) string:
			str = ReplaceAllStringSubmatchFunc(jiration.re, str, v)
		default:
			fmt.Printf("I don't know about type %T!\n", v)
		}
	}
	return str
}

type JiraResolver struct {
	JiraClient *jira.Client
}

// JiraMarkupMentionToEmail will replace JiraMarkup account mentions
// with Display Name followed by parenthetical email addresses
func (j *JiraResolver) JiraMarkupMentionToEmail(str string) string {
	re := regexp.MustCompile(`(?m)(\[~accountid:)([a-zA-Z0-9-:]+)(\])`)
	rfunc := func(groups []string) string {
		// groups[0] is initial match
		accountID := groups[2]

		jiraUser, resp, err := j.JiraClient.User.Get(accountID)

		if c := resp.StatusCode; err != nil || (c < 200 || c > 299) {
			// we cannot resolve it, so just leave it as it was
			return groups[0]
		}

		return jiraUser.DisplayName + " (" + jiraUser.EmailAddress + ")"
	}
	return ReplaceAllStringSubmatchFunc(re, str, rfunc)
}

// ReplaceAllStringSubmatchFunc - Invokes Callback for Regex Replacement
// The repl function takes an unusual string slice argument:
// - The 0th element is the complete match
// - The following slice elements are the nth string found
// by a parenthesized capture group (including named capturing groups)
//
// This is a Go implementation to match other languages:
// PHP: preg_replace_callback($pattern, $callback, $subject)
// Ruby: subject.gsub(pattern) {|match| callback}
// Python: re.sub(pattern, callback, subject)
// JavaScript: subject.replace(pattern, callback)
// See https://gist.github.com/elliotchance/d419395aa776d632d897
func ReplaceAllStringSubmatchFunc(
	re *regexp.Regexp,
	str string,
	repl func([]string) string,
) string {
	result := ""
	lastIndex := 0

	for _, v := range re.FindAllSubmatchIndex([]byte(str), -1) {
		groups := []string{}
		for i := 0; i < len(v); i += 2 {
			if v[i] == -1 || v[i+1] == -1 {
				// if the group is not found, avoid possible error
				groups = append(groups, "")
			} else {
				groups = append(groups, str[v[i]:v[i+1]])
			}
		}

		result += str[lastIndex:v[0]] + repl(groups)
		lastIndex = v[1]
	}

	return result + str[lastIndex:]
}

func JiraMarkupToGithubMarkdown(jiraClient *jira.Client, str string) string {
	jiraAccountResolver := JiraResolver{
		JiraClient: jiraClient,
	}
	resolved := jiraAccountResolver.JiraMarkupMentionToEmail(str)
	return JiraToMD(resolved)
}
