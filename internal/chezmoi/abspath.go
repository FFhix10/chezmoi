package chezmoi

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

var (
	DotAbsPath   = NewAbsPath(".")
	EmptyAbsPath = NewAbsPath("")
	RootAbsPath  = NewAbsPath("/")
)

// An AbsPath is an absolute path.
type AbsPath struct {
	absPath string
}

// NewAbsPath returns a new AbsPath.
func NewAbsPath(absPath string) AbsPath {
	return AbsPath{
		absPath: absPath,
	}
}

// Base returns p's basename.
func (p AbsPath) Base() string {
	return path.Base(p.absPath)
}

// Bytes returns p as a []byte.
func (p AbsPath) Bytes() []byte {
	return []byte(p.absPath)
}

// Dir returns p's directory.
func (p AbsPath) Dir() AbsPath {
	return NewAbsPath(path.Dir(p.absPath))
}

// Empty returns if p is empty.
func (p AbsPath) Empty() bool {
	return p.absPath == ""
}

// Ext returns p's extension.
func (p AbsPath) Ext() string {
	return path.Ext(p.absPath)
}

// Join returns a new AbsPath with relPaths appended.
func (p AbsPath) Join(relPaths ...RelPath) AbsPath {
	relPathStrs := make([]string, 0, len(relPaths)+1)
	relPathStrs = append(relPathStrs, p.absPath)
	for _, relPath := range relPaths {
		relPathStrs = append(relPathStrs, string(relPath))
	}
	return NewAbsPath(path.Join(relPathStrs...))
}

// JoinStr returns a new AbsPath with ss appended.
func (p AbsPath) JoinStr(ss ...string) AbsPath {
	strs := make([]string, 0, len(ss)+1)
	strs = append(strs, p.absPath)
	strs = append(strs, ss...)
	return NewAbsPath(path.Join(strs...))
}

// Len returns the length of p.
func (p AbsPath) Len() int {
	return len(p.absPath)
}

// MarshalText implements encoding.TextMarshaler.MarshalText.
func (p AbsPath) MarshalText() ([]byte, error) {
	return []byte(p.absPath), nil
}

// MustTrimDirPrefix is like TrimPrefix but panics on any error.
func (p AbsPath) MustTrimDirPrefix(dirPrefix AbsPath) RelPath {
	relPath, err := p.TrimDirPrefix(dirPrefix)
	if err != nil {
		panic(err)
	}
	return relPath
}

// Set implements github.com/spf13/pflag.Value.Set.
func (p *AbsPath) Set(s string) error {
	if s == "" {
		p.absPath = ""
		return nil
	}
	homeDirAbsPath, err := homeDirAbsPath()
	if err != nil {
		return err
	}
	absPath, err := NewAbsPathFromExtPath(s, homeDirAbsPath)
	if err != nil {
		return err
	}
	*p = absPath
	return nil
}

// Split returns p's directory and file.
func (p AbsPath) Split() (AbsPath, RelPath) {
	dir, file := path.Split(p.String())
	return NewAbsPath(dir), RelPath(file)
}

func (p AbsPath) String() string {
	return p.absPath
}

// ToSlash calls filepath.ToSlash on p.
func (p AbsPath) ToSlash() AbsPath {
	return NewAbsPath(filepath.ToSlash(p.absPath))
}

// TrimDirPrefix trims prefix from p.
func (p AbsPath) TrimDirPrefix(dirPrefixAbsPath AbsPath) (RelPath, error) {
	if p == dirPrefixAbsPath {
		return "", nil
	}
	dirAbsPath := dirPrefixAbsPath
	if dirAbsPath.absPath != "/" {
		dirAbsPath.absPath += "/"
	}
	if !strings.HasPrefix(p.absPath, dirAbsPath.absPath) {
		return "", &notInAbsDirError{
			pathAbsPath: p,
			dirAbsPath:  dirPrefixAbsPath,
		}
	}
	return RelPath(p.absPath[len(dirAbsPath.absPath):]), nil
}

// Type implements github.com/spf13/pflag.Value.Type.
func (p AbsPath) Type() string {
	return "path"
}

// UnmarshalText implements encoding.TextUnmarshaler.UnmarshalText.
func (p *AbsPath) UnmarshalText(text []byte) error {
	return p.Set(string(text))
}

// StringToAbsPathHookFunc is a github.com/mitchellh/mapstructure.DecodeHookFunc
// that parses an AbsPath from a string.
func StringToAbsPathHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data interface{}) (interface{}, error) {
		if to != reflect.TypeOf(EmptyAbsPath) {
			return data, nil
		}
		s, ok := data.(string)
		if !ok {
			return nil, fmt.Errorf("expected a string, got a %T", data)
		}
		var absPath AbsPath
		if err := absPath.Set(s); err != nil {
			return nil, err
		}
		return absPath, nil
	}
}

// homeDirAbsPath returns the user's home directory as an AbsPath.
func homeDirAbsPath() (AbsPath, error) {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		return EmptyAbsPath, err
	}
	absPath, err := NormalizePath(userHomeDir)
	if err != nil {
		return EmptyAbsPath, err
	}
	return absPath, nil
}
