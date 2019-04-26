package core
import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Pack stores the modpack metadata, usually in pack.toml
type Pack struct {
	Name  string `toml:"name"`
	Index struct {
		File       string `toml:"file"`
		HashFormat string `toml:"hash-format"`
		Hash       string `toml:"hash"`
	} `toml:"index"`
	Versions map[string]string         `toml:"versions"`
	Client   map[string]toml.Primitive `toml:"client"`
	Server   map[string]toml.Primitive `toml:"server"`
	flags    Flags
}

// LoadPack loads the modpack metadata to a Pack struct
func LoadPack(flags Flags) (Pack, error) {
	data, err := ioutil.ReadFile(flags.PackFile)
	if err != nil {
		return Pack{}, err
	}
	var modpack Pack
	if _, err := toml.Decode(string(data), &modpack); err != nil {
		return Pack{}, err
	}

	if len(modpack.Index.File) == 0 {
		modpack.Index.File = "index.toml"
	}
	modpack.flags = flags
	return modpack, nil
}

// LoadIndex attempts to load the index file of this modpack
func (pack Pack) LoadIndex() (Index, error) {
	if filepath.IsAbs(pack.Index.File) {
		return LoadIndex(pack.Index.File, pack.flags)
	}
	return LoadIndex(filepath.Join(filepath.Dir(pack.flags.PackFile), pack.Index.File), pack.flags)
}

// UpdateIndexHash recalculates the hash of the index file of this modpack
func (pack *Pack) UpdateIndexHash() error {
	indexFile := filepath.Join(filepath.Dir(pack.flags.PackFile), pack.Index.File)
	if filepath.IsAbs(pack.Index.File) {
		indexFile = pack.Index.File
	}

	f, err := os.Open(indexFile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Hash usage strategy (may change):
	// Just use SHA256, overwrite existing hash regardless of what it is
	// May update later to continue using the same hash that was already being used
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return err
	}
	hashString := hex.EncodeToString(h.Sum(nil))

	pack.Index.HashFormat = "sha256"
	pack.Index.Hash = hashString
	return nil
}

// Write saves the pack file
func (pack Pack) Write() error {
	f, err := os.Create(pack.flags.PackFile)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := toml.NewEncoder(f)
	// Disable indentation
	enc.Indent = ""
	return enc.Encode(pack)
}

