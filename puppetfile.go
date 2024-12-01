package puppetfile

import (
	"fmt"
	"net/url"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Puppetfile struct {
	ModuleDir string   `("moduledir" @String)?`
	Forge     string   `("forge" @String)?`
	Modules   []Module `@@*`
}

type Module struct {
	Name string `"mod" @String ","?`
	Spec Spec   `(@@)?`
}

type Spec struct {
	Version    string     `@String`
	GitSpec    GitSpec    `| @@`
	LegacySpec LegacySpec `| @@`
	LocalSpec  LocalSpec  `| @@`
}

type GitSpec struct {
	URL           string `Git @String ","?`
	Ref           string `(Ref @String ","?)?`
	Branch        string `(Branch ( @String | @ControlBranch ) ","?)?`
	DefaultBranch string `(DefaultBranch @String ","?)?`
	Tag           string `(Tag @String ","?)?`
	Commit        string `(Commit @String ","?)?`
}

type LegacySpec struct {
	Type    string `LegacyType @String ","?`
	Source  string `(LegacySource @String ","?)?`
	Version string `(LegacyVersion @String ","?)?`
}

type LocalSpec struct {
	Local bool `Local @Bool`
}

func (p *Puppetfile) Validate() error {
	if p.Forge != "" {
		_, err := url.Parse(p.Forge)
		if err != nil {
			return fmt.Errorf("invalid forge URL: %v", err)
		}
	}
	return nil
}

func New() *participle.Parser[Puppetfile] {
	var puppetfileLexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "String", Pattern: `'[^']*'|"[^"]*"`},
		{Name: "Keyword", Pattern: `\b(moduledir|forge|mod)\b`},
		{Name: "Whitespace", Pattern: `[ \t]+`},
		{Name: `Punct`, Pattern: `,`},
		{Name: "Comment", Pattern: `#.*`},
		{Name: "Newline", Pattern: `\n`},
		{Name: "Git", Pattern: `\:git\s*=>`},
		{Name: "Ref", Pattern: `\:ref\s*=>`},
		{Name: "Branch", Pattern: `\:branch\s*=>`},
		{Name: "Tag", Pattern: `\:tag\s*=>`},
		{Name: "Commit", Pattern: `\:commit\s*=>`},
		{Name: "ControlBranch", Pattern: `\:control_branch`},
		{Name: "DefaultBranch", Pattern: `\:default_branch\s*=>`},
		{Name: "LegacyType", Pattern: `type\:`},
		{Name: "LegacySource", Pattern: `source\:`},
		{Name: "LegacyVersion", Pattern: `version\:`},
		{Name: "Local", Pattern: `:local\s*=>`},
		{Name: "Bool", Pattern: `true|false`},
	})

	return participle.MustBuild[Puppetfile](
		participle.Lexer(puppetfileLexer),
		participle.Unquote("String"),
		participle.Elide("Comment", "Whitespace", "Newline"),
	)
}
