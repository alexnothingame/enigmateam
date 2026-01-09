
package config

import (
  "encoding/json"
  "fmt"
  "os"
)

type CommandSpec struct {
  Name   string   `json:"name"`
  Slash  string   `json:"slash"`
  Method string   `json:"method"`
  Path   string   `json:"path"`
  Desc   string   `json:"desc"`
  Args   []string `json:"args"`
}

type GroupSpec struct {
  Title    string   `json:"title"`
  Prefix   string   `json:"prefix"`
  Commands []string `json:"commands"`
}

type Registry struct {
  Groups   []GroupSpec   `json:"groups"`
  Commands []CommandSpec `json:"commands"`
  bySlash  map[string]CommandSpec
  byName   map[string]CommandSpec
}

func LoadRegistry(path string) (*Registry, error) {
  b, err := os.ReadFile(path)
  if err != nil { return nil, err }
  var r Registry
  if err := json.Unmarshal(b, &r); err != nil {
    return nil, fmt.Errorf("не удалось прочитать %s: %w", path, err)
  }
  r.bySlash = map[string]CommandSpec{}
  r.byName = map[string]CommandSpec{}
  for _, c := range r.Commands {
    r.bySlash[c.Slash] = c
    r.byName[c.Name] = c
  }
  return &r, nil
}

func (r *Registry) FindBySlash(slash string) (CommandSpec, bool) {
  c, ok := r.bySlash[slash]
  return c, ok
}

func (r *Registry) FindByName(name string) (CommandSpec, bool) {
  c, ok := r.byName[name]
  return c, ok
}
