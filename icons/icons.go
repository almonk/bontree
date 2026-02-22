package icons

import "strings"

// Default icons
const (
	FolderClosed = "\uf07b" // 
	FolderOpen   = "\uf07c" // 
	FileDefault  = "\uf15b" // 
	ChevronRight = "\ue5fe" // 
	ChevronDown  = "\ue5ff" // 
	GitIcon      = "\ue702" // 
	Indent       = "│"
	Branch       = "├"
	LastBranch   = "└"
	Dash         = "─"
)

// Extension-based icons mapping (Nerd Font)
var extIcons = map[string]string{
	// Programming languages
	".go":     "\ue65e", // 
	".py":     "\ue73c", // 
	".js":     "\ue74e", // 
	".ts":     "\ue628", // 
	".tsx":    "\ue7ba", // 
	".jsx":    "\ue7ba", // 
	".rs":     "\ue7a8", // 
	".rb":     "\ue739", // 
	".java":   "\ue738", // 
	".c":      "\ue61e", // 
	".cpp":    "\ue61d", // 
	".cc":     "\ue61d", // 
	".h":      "\ue61e", // 
	".hpp":    "\ue61d", // 
	".cs":     "\uf81a", // 
	".swift":  "\ue755", // 
	".kt":     "\ue634", // 
	".lua":    "\ue620", // 
	".php":    "\ue73d", // 
	".r":      "\uf25d", // 
	".scala":  "\ue737", // 
	".zig":    "\ue6a9", // 
	".nim":    "\ue677", // 
	".hs":     "\ue777", // 
	".ex":     "\ue62d", // 
	".exs":    "\ue62d", // 
	".erl":    "\ue7b1", // 
	".clj":    "\ue768", // 
	".dart":   "\ue798", // 
	".v":      "\ue6ac", // 
	".ml":     "\ue67a", // 

	// Shell / scripting
	".sh":   "\ue795", // 
	".bash": "\ue795", // 
	".zsh":  "\ue795", // 
	".fish": "\ue795", // 
	".ps1":  "\ue795", // 
	".bat":  "\ue795", // 

	// Web
	".html": "\ue736", // 
	".htm":  "\ue736", // 
	".css":  "\ue749", // 
	".scss": "\ue749", // 
	".sass": "\ue749", // 
	".less": "\ue749", // 
	".vue":  "\ue6a0", // 
	".svelte": "\ue697", // 

	// Data / Config
	".json":  "\ue60b", // 
	".yaml":  "\ue60b", // 
	".yml":   "\ue60b", // 
	".toml":  "\ue60b", // 
	".xml":   "\ue619", // 
	".csv":   "\uf1c3", // 
	".sql":   "\ue706", // 
	".graphql": "\ue662", // 

	// Docs
	".md":   "\ue73e", // 
	".txt":  "\uf15c", // 
	".pdf":  "\uf1c1", // 
	".doc":  "\uf1c2", // 
	".docx": "\uf1c2", // 
	".rst":  "\uf15c", // 
	".tex":  "\uf15c", // 

	// Images
	".png":  "\uf1c5", // 
	".jpg":  "\uf1c5", // 
	".jpeg": "\uf1c5", // 
	".gif":  "\uf1c5", // 
	".svg":  "\uf1c5", // 
	".ico":  "\uf1c5", // 
	".webp": "\uf1c5", // 
	".bmp":  "\uf1c5", // 

	// Archives
	".zip":  "\uf1c6", // 
	".tar":  "\uf1c6", // 
	".gz":   "\uf1c6", // 
	".bz2":  "\uf1c6", // 
	".xz":   "\uf1c6", // 
	".rar":  "\uf1c6", // 
	".7z":   "\uf1c6", // 

	// Build / package
	".lock": "\uf023", // 
	".sum":  "\uf023", // 

	// Docker
	".dockerfile": "\ue7b0", // 

	// Config
	".env":  "\uf462", // 
	".ini":  "\ue615", // 
	".cfg":  "\ue615", // 
	".conf": "\ue615", // 

	// Video / Audio
	".mp3":  "\uf001", // 
	".wav":  "\uf001", // 
	".mp4":  "\uf03d", // 
	".avi":  "\uf03d", // 
	".mkv":  "\uf03d", // 
	".mov":  "\uf03d", // 

	// Binary / Compiled
	".o":    "\uf471", // 
	".so":   "\uf471", // 
	".dll":  "\uf471", // 
	".exe":  "\uf471", // 
	".wasm": "\ue6a1", // 

	// Misc
	".log": "\uf18d", // 
}

// Filename-based icons
var nameIcons = map[string]string{
	"Makefile":       "\ue615", // 
	"makefile":       "\ue615",
	"Dockerfile":     "\ue7b0", // 
	"dockerfile":     "\ue7b0",
	"docker-compose.yml":  "\ue7b0",
	"docker-compose.yaml": "\ue7b0",
	".gitignore":     "\ue702", // 
	".gitmodules":    "\ue702",
	".gitattributes": "\ue702",
	"go.mod":         "\ue65e", // 
	"go.sum":         "\ue65e",
	"Cargo.toml":     "\ue7a8",
	"Cargo.lock":     "\ue7a8",
	"package.json":   "\ue71e", // 
	"package-lock.json": "\ue71e",
	"tsconfig.json":  "\ue628",
	"webpack.config.js": "\ue74e",
	"LICENSE":        "\uf15c", // 
	"license":        "\uf15c",
	"README.md":      "\ue73e",
	"readme.md":      "\ue73e",
	".env":           "\uf462",
	".env.local":     "\uf462",
	".env.development": "\uf462",
	".env.production":  "\uf462",
	"Gemfile":        "\ue739",
	"Rakefile":       "\ue739",
	"requirements.txt": "\ue73c",
	"setup.py":       "\ue73c",
	"Pipfile":        "\ue73c",
	"CMakeLists.txt": "\ue615",
	"justfile":       "\ue615",
	"Justfile":       "\ue615",
	".editorconfig":  "\ue615",
	".prettierrc":    "\ue615",
	".eslintrc":      "\ue615",
	".eslintrc.js":   "\ue615",
	".eslintrc.json": "\ue615",
	"flake.nix":      "\uf313",
	"flake.lock":     "\uf313",
	"default.nix":    "\uf313",
	"shell.nix":      "\uf313",
}

// GetIcon returns the appropriate nerd font icon for a file
func GetIcon(name string, isDir bool, isOpen bool) string {
	if isDir {
		if isOpen {
			return FolderOpen
		}
		return FolderClosed
	}

	// Check exact filename first
	if icon, ok := nameIcons[name]; ok {
		return icon
	}

	// Check extension
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		if icon, ok := extIcons[strings.ToLower(name[i:])]; ok {
			return icon
		}
	}

	// Check if hidden/dot file
	if strings.HasPrefix(name, ".") {
		return "\ue615" // 
	}

	return FileDefault
}
