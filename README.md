# wti - (W)hat (T)he (I)ssue?
Utility that tries to get a description of the specified ISSUE from JIRA Cloud. 

```
git wti TEAM-1234
```
If the ISSUE is found, by default it will output on one line 
the Issue Key (TEAM-1234), a dash separator, then the issue summary, followed 
by the **Github Markdown translation** of the multi-line issue description. 

For example:
```
ISSUE-1234 — Feature: Make wobulator frobnobicate frabjously.
This is a very important feature that we really want done soon.
It should be extremely well crafted though.
```
You may suppress output of the first line with `--no-title`, or suppress output of the multi-line description,
with `--no-description`

### Installation
For now, you can run `make build` from this directory.

### Usage

WTI (or `git-wti`) will use an existing JSON `~/.config/jira` config file, or environment variables:
```
ATLASSIAN_API_TOKEN
ATLASSIAN_HOST
ATLASSIAN_API_USER
```
The `~/.config/jira` file format is:
```json
{"host":"https://khanacademy.atlassian.net","user":"username@khanacademy.org","token":"sOmEtOkEn"}
```
The token must be a [JIRA Service Token](https://confluence.atlassian.com/cloud/api-tokens-938839638.html?_ga=2.71928019.1521673145.1568996908-1205092387.1568813216) (see below for directions to obtain one).

##### Create an API token
Per the [Create a JIRA Service Token](https://confluence.atlassian.com/cloud/api-tokens-938839638.html?_ga=2.71928019.1521673145.1568996908-1205092387.1568813216) documentation, create an API token from your Atlassian account:

1. Log in to https://id.atlassian.com/manage/api-tokens.
2. Click Create API token.
3. From the dialog that appears, enter a memorable and concise Label for your token and click Create.
4. Click Copy to clipboard, then paste the token to your script, or elsewhere to save:

<table><tr><td>:bulb: <b>NOTE:</b> For security reasons it isn't possible to view the token after closing the creation dialog; if necessary, create a new token.<br/>
You should store the token securely, just as for any password.
</td></tr></table>

##### Test an API token
A primary use case for API tokens is to allow scripts to access REST APIs for Atlassian Cloud applications using HTTP basic authentication.

Depending on the details of the HTTP library you use, simply replace your password with the token. For example, when using curl, you could do something like this:
```shell script
curl -v -L \
https://mycompany.com/rest/api/2/issue/JIRA-116 \
--user $(whoami)@khanacademy.org:${JIRA_API_TOKEN} | jq '{"key": .key, "summary": .fields.summary, "description": .fields.description}'
```
Note that `$(whoami)@khanacademy.org` here is intended to be the email address for the Atlassian account you're using to create the token

### TODO?

Currently, a JIRA account reference `[~accountid:5b4ce9b2dc42b22bc590bd17]` will be translated as `Bob Froptoppitt (bob@khanacademy.org)`. It is unclear if
further translating these to Github account mentions would be desirable, since not all JIRA users have Github accounts.

Similarly, the JIRA translation is a little half-baked. If users have pandoc installed, we could shell out to that and use that if it turned out to be better. Making a robust complete translation would require a PEG ([Parsing expression grammar](https://en.wikipedia.org/wiki/Parsing_expression_grammar)).

But that would involve more time than **this** simple toy maker can find in their weekend. 