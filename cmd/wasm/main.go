// WASM entry point for the bontree interactive demo.
// This is a thin JS bridge that delegates all logic to the ui package.
package main

import (
	"strings"
	"syscall/js"

	"github.com/almonk/bontree/config"
	"github.com/almonk/bontree/theme"
	"github.com/almonk/bontree/tree"
	"github.com/almonk/bontree/ui"
)

// ── Build demo tree ──

func buildDemoTree() *tree.Node {
	root := &tree.Node{Name: "my-project", Path: ".", IsDir: true, Expanded: true, Depth: 0}

	add := func(parent *tree.Node, name string, isDir bool, expanded bool) *tree.Node {
		n := &tree.Node{
			Name:     name,
			Path:     strings.TrimPrefix(parent.Path+"/"+name, "./"),
			IsDir:    isDir,
			Expanded: expanded,
			Parent:   parent,
			Depth:    parent.Depth + 1,
			Loaded:   true,
		}
		parent.Children = append(parent.Children, n)
		return n
	}

	// .github
	gh := add(root, ".github", true, false)
	wf := add(gh, "workflows", true, false)
	add(wf, "ci.yml", false, false)

	// cmd
	cmd := add(root, "cmd", true, true)
	srv := add(cmd, "server", true, true)
	add(srv, "main.go", false, false)
	add(srv, "routes.go", false, false)
	add(srv, "middleware.go", false, false)
	cli := add(cmd, "cli", true, false)
	add(cli, "root.go", false, false)
	add(cli, "serve.go", false, false)

	// internal
	internal := add(root, "internal", true, true)
	auth := add(internal, "auth", true, true)
	add(auth, "jwt.go", false, false)
	add(auth, "jwt_test.go", false, false)
	add(auth, "oauth.go", false, false)
	db := add(internal, "db", true, false)
	add(db, "postgres.go", false, false)
	add(db, "migrations.go", false, false)
	add(db, "schema.sql", false, false)
	handlers := add(internal, "handlers", true, false)
	add(handlers, "users.go", false, false)
	add(handlers, "posts.go", false, false)
	add(handlers, "health.go", false, false)
	models := add(internal, "models", true, false)
	add(models, "user.go", false, false)
	add(models, "post.go", false, false)

	// web
	web := add(root, "web", true, false)
	src := add(web, "src", true, false)
	add(src, "App.tsx", false, false)
	add(src, "index.tsx", false, false)
	comps := add(src, "components", true, false)
	add(comps, "Header.tsx", false, false)
	add(comps, "Footer.tsx", false, false)
	add(web, "package.json", false, false)
	add(web, "tsconfig.json", false, false)
	add(web, "vite.config.ts", false, false)

	// docs
	docs := add(root, "docs", true, false)
	add(docs, "architecture.md", false, false)
	add(docs, "api.md", false, false)
	add(docs, "deployment.md", false, false)

	// root files
	add(root, ".env.example", false, false)
	add(root, ".gitignore", false, false)
	add(root, "docker-compose.yml", false, false)
	add(root, "Dockerfile", false, false)
	add(root, "go.mod", false, false)
	add(root, "go.sum", false, false)
	add(root, "Makefile", false, false)
	add(root, "README.md", false, false)

	return root
}

func demoGitFiles() map[string]ui.GitFileStatus {
	return map[string]ui.GitFileStatus{
		"cmd/server/main.go":         ui.GitModified,
		"cmd/server/middleware.go":   ui.GitAdded,
		"cmd/server":                ui.GitModified,
		"cmd":                       ui.GitModified,
		"internal/auth/jwt.go":      ui.GitModified,
		"internal/auth/jwt_test.go": ui.GitModified,
		"internal/auth/oauth.go":    ui.GitAdded,
		"internal/auth":             ui.GitModified,
		"internal/handlers/posts.go": ui.GitModified,
		"internal/handlers":         ui.GitModified,
		"internal":                  ui.GitModified,
		"docs/api.md":               ui.GitAdded,
		"docs":                      ui.GitAdded,
		"Dockerfile":                ui.GitModified,
	}
}

// tokyoNightTheme returns a Theme with Tokyo Night colors for the WASM demo.
func tokyoNightTheme() *theme.Theme {
	return &theme.Theme{
		Name:       "tokyonight",
		Background: "#1a1b26",
		Foreground: "#c0caf5",
		Palette: [16]string{
			0:  "#1a1b26", // black
			1:  "#f7768e", // red
			2:  "#9ece6a", // green
			3:  "#e0af68", // yellow
			4:  "#7aa2f7", // blue
			5:  "#bb9af7", // purple
			6:  "#7dcfff", // cyan
			7:  "#a9b1d6", // white (dim fg)
			8:  "#565f89", // bright black (comment/gutter)
			9:  "#f7768e", // bright red
			10: "#9ece6a", // bright green
			11: "#e0af68", // bright yellow
			12: "#7aa2f7", // bright blue
			13: "#bb9af7", // bright purple
			14: "#7dcfff", // bright cyan
			15: "#c0caf5", // bright white
		},
		SelectionBackground: "#33467c",
	}
}

// ── JS bridge ──

var app *ui.Model

func main() {
	// Apply Tokyo Night theme to the shared ui styles
	ui.ApplyTheme(tokyoNightTheme())

	root := buildDemoTree()
	cfg := config.DefaultConfig()

	m := ui.NewDemo(root, cfg)
	m.SetGitInfo("main", demoGitFiles())
	app = &m

	// Expose functions to JS
	js.Global().Set("bontreeInit", js.FuncOf(bontreeInit))
	js.Global().Set("bontreeKey", js.FuncOf(bontreeKey))
	js.Global().Set("bontreeClick", js.FuncOf(bontreeClick))
	js.Global().Set("bontreeScroll", js.FuncOf(bontreeScroll))
	js.Global().Set("bontreeClearFlash", js.FuncOf(bontreeClearFlash))

	// Keep alive
	select {}
}

func bontreeInit(_ js.Value, args []js.Value) interface{} {
	cols := args[0].Int()
	rows := args[1].Int()
	app.SetSize(cols, rows)
	return app.View()
}

func bontreeKey(_ js.Value, args []js.Value) interface{} {
	key := args[0].String()

	// Determine if this is a printable rune (not a named key like "esc", "enter", etc.)
	isRune := len(key) == 1 && key[0] >= ' '

	result := app.HandleKey(key, isRune)

	if result.Quit {
		// Can't quit in WASM, just ignore
		return app.View()
	}

	if result.CopyPath != "" {
		js.Global().Get("navigator").Get("clipboard").Call("writeText", result.CopyPath)
	}

	if result.FlashMsg != "" {
		app.SetFlash(result.FlashMsg)
	}

	r := js.Global().Get("Object").New()
	r.Set("view", app.View())
	r.Set("flash", app.HasFlash())
	return r
}

func bontreeClick(_ js.Value, args []js.Value) interface{} {
	row := args[0].Int()
	doubleClick := args[1].Bool()
	app.HandleClick(row, doubleClick)
	return app.View()
}

func bontreeScroll(_ js.Value, args []js.Value) interface{} {
	dir := args[0].Int()
	app.HandleScroll(dir)
	return app.View()
}

func bontreeClearFlash(_ js.Value, _ []js.Value) interface{} {
	app.ClearFlash()
	return app.View()
}
