package scanner

import (
	"embed"
	"io/fs"
	"path/filepath"

	"github.com/pkg/errors"
)

//go:embed templates templates/*/.dockerignore templates/*/*/.dockerignore templates/**/.fly
var content embed.FS

type InitCommand struct {
	Command     string
	Args        []string
	Description string
	Condition   bool
}

type Secret struct {
	Key      string
	Help     string
	Value    string
	Generate func() (string, error)
}

type SourceInfo struct {
	Family                       string
	Version                      string
	DockerfilePath               string
	BuildArgs                    map[string]string
	Builder                      string
	ReleaseCmd                   string
	DockerCommand                string
	DockerEntrypoint             string
	KillSignal                   string
	Buildpacks                   []string
	Secrets                      []Secret
	Files                        []SourceFile
	Port                         int
	Env                          map[string]string
	Statics                      []Static
	Processes                    map[string]string
	DeployDocs                   string
	Notice                       string
	SkipDeploy                   bool
	SkipDatabase                 bool
	Volumes                      []Volume
	DockerfileAppendix           []string
	InitCommands                 []InitCommand
	PostgresInitCommands         []InitCommand
	PostgresInitCommandCondition bool
}

type SourceFile struct {
	Path     string
	Contents []byte
}
type Static struct {
	GuestPath string `toml:"guest_path" json:"guest_path"`
	UrlPrefix string `toml:"url_prefix" json:"url_prefix"`
}
type Volume struct {
	Source      string `toml:"source" json:"source"`
	Destination string `toml:"destination" json:"destination"`
}

func Scan(sourceDir string) (*SourceInfo, error) {
	scanners := []sourceScanner{
		configureDjango,
		configureLaravel,
		configurePhoenix,
		configureRails,
		configureRedwood,
		/* frameworks scanners are placed before generic scanners,
		   since they might mix languages or have a Dockerfile that
			 doesn't work with Fly */
		configureDockerfile,
		configureLucky,
		configureRuby,
		configureGo,
		configureElixir,
		configurePython,
		configureDeno,
		configureRemix,
		configureNuxt,
		configureNextJs,
		configureNode,
		configureStatic,
	}

	for _, scanner := range scanners {
		si, err := scanner(sourceDir)
		if err != nil {
			return nil, err
		}
		if si != nil {
			return si, nil
		}
	}

	return nil, nil
}

type sourceScanner func(sourceDir string) (*SourceInfo, error)

// templates recursively returns files from the templates directory within the named directory
// will panic on errors since these files are embedded and should work
func templates(name string) (files []SourceFile) {
	err := fs.WalkDir(content, name, func(path string, d fs.DirEntry, e error) error {
		if d.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(name, path)
		if err != nil {
			return errors.Wrap(err, "error removing template prefix")
		}

		data, err := fs.ReadFile(content, path)
		if err != nil {
			return err
		}

		if err != nil {
			return err
		}

		f := SourceFile{
			Path:     relPath,
			Contents: data,
		}

		files = append(files, f)
		return nil
	})
	if err != nil {
		panic(err)
	}

	return
}
