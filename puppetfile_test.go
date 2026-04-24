package puppetfile

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParserParsesReferenceSyntaxVariants(t *testing.T) {
	t.Parallel()

	parser := New()

	tests := []struct {
		name    string
		content string
	}{
		{
			name: "forge before moduledir",
			content: "forge 'https://forgeapi.puppet.com'\n" +
				"moduledir 'modules'\n" +
				"mod 'puppetlabs/apache'\n",
		},
		{
			name:    "forge latest symbol",
			content: "mod 'puppetlabs/apache', :latest\n",
		},
		{
			name: "git install_path and exclude_spec",
			content: "mod 'apache',\n" +
				"  :git => 'git@github.com:puppetlabs/puppetlabs-apache.git',\n" +
				"  :install_path => 'external',\n" +
				"  :exclude_spec => false\n",
		},
		{
			name: "svn syntax",
			content: "mod 'apache',\n" +
				"  :svn => 'https://github.com/puppetlabs/puppetlabs-apache/trunk',\n" +
				"  :revision => '154',\n" +
				"  :username => 'user',\n" +
				"  :password => 'hunter2'\n",
		},
		{
			name: "legacy install_path and exclude_spec",
			content: "mod 'puppetlabs-apache',\n" +
				"  type: 'tarball',\n" +
				"  source: 'https://repo.example.com/puppet/modules/puppetlabs-apache-7.0.0.tar.gz',\n" +
				"  install_path: 'external',\n" +
				"  exclude_spec: false\n",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ast, err := parser.ParseString("test", tc.content)
			require.NoError(t, err)
			require.NoError(t, ast.Validate())
		})
	}
}

func TestValidateRejectsDuplicateModuleNames(t *testing.T) {
	t.Parallel()

	parser := New()
	content := "mod 'apache'\nmod 'apache', :git => 'https://github.com/puppetlabs/puppetlabs-apache.git'\n"

	ast, err := parser.ParseString("test", content)
	require.NoError(t, err)
	validateErr := ast.Validate()
	require.Error(t, validateErr)
	require.Contains(t, validateErr.Error(), "duplicate module declaration")
}

func TestValidateRejectsDifferentForgeDeclarations(t *testing.T) {
	t.Parallel()

	parser := New()
	content := "forge 'https://forgeapi.puppet.com'\nmoduledir 'modules'\nforge 'https://forgeapi.puppetlabs.com'\n"

	ast, err := parser.ParseString("test", content)
	require.NoError(t, err)
	validateErr := ast.Validate()
	require.Error(t, validateErr)
	require.Contains(t, validateErr.Error(), "multiple forge declarations")
}
