// Package wiki generates HTML from markdown for a given wiki.
package wiki

import (
	"embed"
	"html/template"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var markdown goldmark.Markdown
var defaultHtmlHeaderTemplate *template.Template
var githubHtmlHeaderTemplate *template.Template

//go:embed static/style.css static/github-style.css
var embeddedFileSystem embed.FS

// defaultHtmlHeaderTemplateText is the text used to create the HTML template that
// generates the start of each HTML file that uses default styles.
const defaultHtmlHeaderTemplateText = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<meta name=generator content="gomarkwiki {{.Version}}">
<title>{{.Title}}</title>
<link rel="icon" href="{{.RootRelPath}}favicon.ico" sizes="any" />
<link href="{{.RootRelPath}}style.css" rel="stylesheet" />
<link href="{{.RootRelPath}}local.css" rel="stylesheet" />
</head>
<body>
`

// githubHtmlHeaderTemplateText is the text used to create the HTML template that
// generates the start of each HTML file that uses GitHub styles.
const githubHtmlHeaderTemplateText = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8" />
<meta name="viewport" content="width=device-width, initial-scale=1" />
<meta name=generator content="gomarkwiki {{.Version}}">
<title>{{.Title}}</title>
<link rel="icon" href="{{.RootRelPath}}favicon.ico" sizes="any" />
<link href="{{.RootRelPath}}github-style.css" rel="stylesheet" />
<link href="{{.RootRelPath}}github-local.css" rel="stylesheet" />
<style>
	.markdown-body {
		box-sizing: border-box;
		min-width: 200px;
		max-width: 980px;
		margin: 0 auto;
		padding: 45px;
	}

	@media (max-width: 767px) {
		.markdown-body {
			padding: 15px;
		}
	}
</style>
</head>
<body>
<article class="markdown-body">
`

// templateData holds the values used to instantiate HTML from the HTML header template.
type templateData struct {
	Title       string
	Version     string
	RootRelPath string
}

func init() {
	// Create markdown converter.
	markdown = goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithAttribute(),
		),
		goldmark.WithRendererOptions(
			html.WithUnsafe(),
		),
	)

	// Create HTML header templates.
	defaultHtmlHeaderTemplate = template.Must(template.New("defaultHtml").Parse(defaultHtmlHeaderTemplateText))
	githubHtmlHeaderTemplate = template.Must(template.New("githubHtml").Parse(githubHtmlHeaderTemplateText))
}
