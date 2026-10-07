// Package web embeds the calculator's browser front end.
package web

import (
	"embed"
	"io/fs"
)

//go:embed static
var files embed.FS

// Static holds index.html, style.css and app.js.
var Static, _ = fs.Sub(files, "static")
