package preferences

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Store keeps the profile in a file separate from credential configuration.
type Store struct{ Path string }

// DefaultStore uses the application's per-user configuration directory.
func DefaultStore() (Store, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Store{}, err
	}
	return Store{Path: filepath.Join(home, ".config", "soi-tro", "search-profile.json")}, nil
}

// Load returns nil when no profile has been saved.
func (s Store) Load() (*SearchProfile, error) {
	f, err := os.Open(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()
	dec := json.NewDecoder(f)
	dec.DisallowUnknownFields()
	var p SearchProfile
	if err := dec.Decode(&p); err != nil {
		return nil, fmt.Errorf("không thể đọc hồ sơ tìm phòng: %w", err)
	}
	var trailing any
	if err := dec.Decode(&trailing); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("hồ sơ tìm phòng có dữ liệu thừa")
	}
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("hồ sơ tìm phòng không hợp lệ: %w", err)
	}
	return &p, nil
}

// Save atomically replaces the validated profile with owner-only file mode.
func (s Store) Save(p SearchProfile) error {
	if err := p.Validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.Path), 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(s.Path), ".search-profile-*")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if err := f.Chmod(0o600); err != nil {
		_ = f.Close()
		return err
	}
	if err := json.NewEncoder(f).Encode(p); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), s.Path)
}

// Clear removes the profile without touching credentials or rental history.
func (s Store) Clear() error {
	err := os.Remove(s.Path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
