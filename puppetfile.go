package puppetfile

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type Puppetfile struct {
	Forge        string   `parser:"(\"forge\" @String)?"`
	ModuleDir    string   `parser:"(\"moduledir\" @String)?"`
	ForgeAfterMD string   `parser:"(\"forge\" @String)?"`
	Modules      []Module `parser:"@@*"`
}

type Module struct {
	Name string `parser:"\"mod\" @String \",\"?"`
	Spec Spec   `parser:"(@@)?"`
}

type Spec struct {
	Version    string     `parser:"(@String | @Latest)"`
	GitSpec    GitSpec    `parser:"| @@"`
	SvnSpec    SvnSpec    `parser:"| @@"`
	LegacySpec LegacySpec `parser:"| @@"`
	LocalSpec  LocalSpec  `parser:"| @@"`
}

type GitSpec struct {
	URL     string      `parser:"Git @String \",\"?"`
	Options []GitOption `parser:"@@*"`
}

type GitOption struct {
	Ref           string `parser:"((Ref @String \",\"?)"`
	Branch        string `parser:"| (Branch (@String | @ControlBranch) \",\"?)"`
	DefaultBranch string `parser:"| (DefaultBranch @String \",\"?)"`
	Tag           string `parser:"| (Tag @String \",\"?)"`
	Commit        string `parser:"| (Commit @String \",\"?)"`
	InstallPath   string `parser:"| (InstallPath @String \",\"?)"`
	ExcludeSpec   *bool  `parser:"| (ExcludeSpec @Bool \",\"?))"`
}

type LegacySpec struct {
	Fields []LegacyField `parser:"@@+"`
}

type LegacyField struct {
	Type        string `parser:"((LegacyType @String \",\"?)"`
	Source      string `parser:"| (LegacySource @String \",\"?)"`
	Version     string `parser:"| (LegacyVersion (@String | @Latest) \",\"?)"`
	InstallPath string `parser:"| (LegacyInstallPath @String \",\"?)"`
	ExcludeSpec *bool  `parser:"| (LegacyExcludeSpec @Bool \",\"?))"`
}

type SvnSpec struct {
	URL     string      `parser:"Svn @String \",\"?"`
	Options []SvnOption `parser:"@@*"`
}

type SvnOption struct {
	Revision    string `parser:"((Rev @String \",\"?)"`
	InstallPath string `parser:"| (InstallPath @String \",\"?)"`
	Username    string `parser:"| (Username @String \",\"?)"`
	Password    string `parser:"| (Password @String \",\"?)"`
	ExcludeSpec *bool  `parser:"| (ExcludeSpec @Bool \",\"?))"`
}

type LocalSpec struct {
	Local bool `parser:"Local @Bool"`
}

func (p *Puppetfile) Validate() error {
	if p.Forge != "" && p.ForgeAfterMD != "" && p.Forge != p.ForgeAfterMD {
		return fmt.Errorf("multiple forge declarations with different values")
	}

	forge := p.Forge
	if forge == "" {
		forge = p.ForgeAfterMD
	}
	if forge != "" {
		_, err := url.Parse(forge)
		if err != nil {
			return fmt.Errorf("invalid forge URL: %v", err)
		}
	}
	p.Forge = forge

	seen := make(map[string]struct{}, len(p.Modules))
	for _, m := range p.Modules {
		if _, ok := seen[m.Name]; ok {
			return fmt.Errorf("duplicate module declaration for %q", m.Name)
		}
		seen[m.Name] = struct{}{}

		if err := m.Spec.validate(); err != nil {
			return fmt.Errorf("module %q: %w", m.Name, err)
		}
	}

	return nil
}

func (s Spec) validate() error {
	if s.Version != "" && s.Version != ":latest" && strings.HasPrefix(s.Version, ":") {
		return fmt.Errorf("unsupported symbolic version %q", s.Version)
	}

	if s.GitSpec.URL != "" {
		var (
			seenRef, seenBranch, seenDefaultBranch, seenTag, seenCommit bool
			seenInstallPath, seenExcludeSpec                           bool
		)
		for _, option := range s.GitSpec.Options {
			if option.Ref != "" {
				if seenRef {
					return fmt.Errorf("duplicate :ref option")
				}
				seenRef = true
			}
			if option.Branch != "" {
				if seenBranch {
					return fmt.Errorf("duplicate :branch option")
				}
				seenBranch = true
			}
			if option.DefaultBranch != "" {
				if seenDefaultBranch {
					return fmt.Errorf("duplicate :default_branch option")
				}
				seenDefaultBranch = true
			}
			if option.Tag != "" {
				if seenTag {
					return fmt.Errorf("duplicate :tag option")
				}
				seenTag = true
			}
			if option.Commit != "" {
				if seenCommit {
					return fmt.Errorf("duplicate :commit option")
				}
				seenCommit = true
			}
			if option.InstallPath != "" {
				if seenInstallPath {
					return fmt.Errorf("duplicate :install_path option")
				}
				seenInstallPath = true
			}
			if option.ExcludeSpec != nil {
				if seenExcludeSpec {
					return fmt.Errorf("duplicate :exclude_spec option")
				}
				seenExcludeSpec = true
			}
		}
	}

	if s.SvnSpec.URL != "" {
		var (
			seenRevision, seenInstallPath, seenUsername, seenPassword, seenExcludeSpec bool
		)
		for _, option := range s.SvnSpec.Options {
			if option.Revision != "" {
				if seenRevision {
					return fmt.Errorf("duplicate revision option")
				}
				seenRevision = true
			}
			if option.InstallPath != "" {
				if seenInstallPath {
					return fmt.Errorf("duplicate :install_path option")
				}
				seenInstallPath = true
			}
			if option.Username != "" {
				if seenUsername {
					return fmt.Errorf("duplicate :username option")
				}
				seenUsername = true
			}
			if option.Password != "" {
				if seenPassword {
					return fmt.Errorf("duplicate :password option")
				}
				seenPassword = true
			}
			if option.ExcludeSpec != nil {
				if seenExcludeSpec {
					return fmt.Errorf("duplicate :exclude_spec option")
				}
				seenExcludeSpec = true
			}
		}
	}

	if len(s.LegacySpec.Fields) > 0 {
		var (
			typeSeen, sourceSeen, versionSeen, installPathSeen, excludeSpecSeen bool
		)
		for _, field := range s.LegacySpec.Fields {
			if field.Type != "" {
				if typeSeen {
					return fmt.Errorf("duplicate type: option")
				}
				typeSeen = true
			}
			if field.Source != "" {
				if sourceSeen {
					return fmt.Errorf("duplicate source: option")
				}
				sourceSeen = true
			}
			if field.Version != "" {
				if versionSeen {
					return fmt.Errorf("duplicate version: option")
				}
				if field.Version != ":latest" && strings.HasPrefix(field.Version, ":") {
					return fmt.Errorf("unsupported symbolic version %q", field.Version)
				}
				versionSeen = true
			}
			if field.InstallPath != "" {
				if installPathSeen {
					return fmt.Errorf("duplicate install_path: option")
				}
				installPathSeen = true
			}
			if field.ExcludeSpec != nil {
				if excludeSpecSeen {
					return fmt.Errorf("duplicate exclude_spec: option")
				}
				excludeSpecSeen = true
			}
		}
	}

	return nil
}

func New() *participle.Parser[Puppetfile] {
	var puppetfileLexer = lexer.MustSimple([]lexer.SimpleRule{
		{Name: "String", Pattern: `'[^']*'|"[^"]*"`},
		{Name: "Latest", Pattern: `:latest`},
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
		{Name: "InstallPath", Pattern: `\:install_path\s*=>`},
		{Name: "ExcludeSpec", Pattern: `\:exclude_spec\s*=>`},
		{Name: "Svn", Pattern: `\:svn\s*=>`},
		{Name: "Rev", Pattern: `\:(rev|revision|version)\s*=>`},
		{Name: "Username", Pattern: `\:username\s*=>`},
		{Name: "Password", Pattern: `\:password\s*=>`},
		{Name: "LegacyType", Pattern: `type\:`},
		{Name: "LegacySource", Pattern: `source\:`},
		{Name: "LegacyVersion", Pattern: `version\:`},
		{Name: "LegacyInstallPath", Pattern: `install_path\:`},
		{Name: "LegacyExcludeSpec", Pattern: `exclude_spec\:`},
		{Name: "Local", Pattern: `:local\s*=>`},
		{Name: "Bool", Pattern: `true|false`},
	})

	return participle.MustBuild[Puppetfile](
		participle.Lexer(puppetfileLexer),
		participle.Unquote("String"),
		participle.Elide("Comment", "Whitespace", "Newline"),
	)
}
