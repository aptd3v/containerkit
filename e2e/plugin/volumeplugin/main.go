// Volume plugin used by the containerkit plugin e2e test.
//
// It implements docker.volumedriver/1.0 over a unix socket at
// /run/docker/plugins/containerkit.sock, as required by
// https://docs.docker.com/engine/extend/#developing-a-plugin.
package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

const (
	sockPath = "/run/docker/plugins/containerkit.sock"
	dataRoot = "/data"
)

type volumePlugin struct {
	mu   sync.Mutex
	vols map[string]string
}

type errResp struct {
	Err string `json:"Err,omitempty"`
}

type nameReq struct {
	Name string `json:"Name"`
}

type mountResp struct {
	Mountpoint string `json:"Mountpoint,omitempty"`
	Err        string `json:"Err,omitempty"`
}

type volumeInfo struct {
	Name       string `json:"Name"`
	Mountpoint string `json:"Mountpoint"`
}

type getResp struct {
	Volume *volumeInfo `json:"Volume,omitempty"`
	Err    string      `json:"Err,omitempty"`
}

type listResp struct {
	Volumes []*volumeInfo `json:"Volumes"`
}

type activateResp struct {
	Implements []string `json:"Implements"`
}

type capabilitiesResp struct {
	Capabilities struct {
		Scope string `json:"Scope"`
	} `json:"Capabilities"`
}

func main() {
	if os.Getenv("DEBUG") == "1" {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	} else {
		log.SetOutput(os.Stderr)
	}
	if err := os.MkdirAll(filepath.Dir(sockPath), 0o755); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(dataRoot, 0o755); err != nil {
		log.Fatal(err)
	}
	_ = os.Remove(sockPath)
	ln, err := net.Listen("unix", sockPath)
	if err != nil {
		log.Fatal(err)
	}
	p := &volumePlugin{vols: map[string]string{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/Plugin.Activate", p.activate)
	mux.HandleFunc("/VolumeDriver.Create", p.create)
	mux.HandleFunc("/VolumeDriver.Remove", p.remove)
	mux.HandleFunc("/VolumeDriver.Mount", p.mount)
	mux.HandleFunc("/VolumeDriver.Path", p.path)
	mux.HandleFunc("/VolumeDriver.Unmount", p.unmount)
	mux.HandleFunc("/VolumeDriver.Get", p.get)
	mux.HandleFunc("/VolumeDriver.List", p.list)
	mux.HandleFunc("/VolumeDriver.Capabilities", p.capabilities)
	log.Printf("listening on %s", sockPath)
	if err := http.Serve(ln, mux); err != nil {
		log.Fatal(err)
	}
}

func (p *volumePlugin) activate(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, activateResp{Implements: []string{"VolumeDriver"}})
}

func (p *volumePlugin) create(w http.ResponseWriter, r *http.Request) {
	var req nameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, errResp{Err: err.Error()})
		return
	}
	mp, err := p.ensure(req.Name)
	if err != nil {
		writeJSON(w, errResp{Err: err.Error()})
		return
	}
	p.mu.Lock()
	p.vols[req.Name] = mp
	p.mu.Unlock()
	writeJSON(w, errResp{})
}

func (p *volumePlugin) remove(w http.ResponseWriter, r *http.Request) {
	var req nameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, errResp{Err: err.Error()})
		return
	}
	p.mu.Lock()
	mp := p.vols[req.Name]
	delete(p.vols, req.Name)
	p.mu.Unlock()
	if mp != "" {
		_ = os.RemoveAll(mp)
	}
	writeJSON(w, errResp{})
}

func (p *volumePlugin) mount(w http.ResponseWriter, r *http.Request) {
	p.respondMount(w, r)
}

func (p *volumePlugin) path(w http.ResponseWriter, r *http.Request) {
	p.respondMount(w, r)
}

func (p *volumePlugin) unmount(w http.ResponseWriter, r *http.Request) {
	var req nameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, errResp{Err: err.Error()})
		return
	}
	writeJSON(w, errResp{})
}

func (p *volumePlugin) get(w http.ResponseWriter, r *http.Request) {
	var req nameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, getResp{Err: err.Error()})
		return
	}
	mp, err := p.lookup(req.Name)
	if err != nil {
		writeJSON(w, getResp{Err: err.Error()})
		return
	}
	writeJSON(w, getResp{Volume: &volumeInfo{Name: req.Name, Mountpoint: mp}})
}

func (p *volumePlugin) list(w http.ResponseWriter, _ *http.Request) {
	p.mu.Lock()
	out := make([]*volumeInfo, 0, len(p.vols))
	for name, mp := range p.vols {
		out = append(out, &volumeInfo{Name: name, Mountpoint: mp})
	}
	p.mu.Unlock()
	writeJSON(w, listResp{Volumes: out})
}

func (p *volumePlugin) capabilities(w http.ResponseWriter, _ *http.Request) {
	var resp capabilitiesResp
	resp.Capabilities.Scope = "local"
	writeJSON(w, resp)
}

func (p *volumePlugin) respondMount(w http.ResponseWriter, r *http.Request) {
	var req nameReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, mountResp{Err: err.Error()})
		return
	}
	mp, err := p.lookup(req.Name)
	if err != nil {
		writeJSON(w, mountResp{Err: err.Error()})
		return
	}
	writeJSON(w, mountResp{Mountpoint: mp})
}

func (p *volumePlugin) ensure(name string) (string, error) {
	mp := mountpoint(name)
	if err := os.MkdirAll(mp, 0o755); err != nil {
		return "", err
	}
	return mp, nil
}

func (p *volumePlugin) lookup(name string) (string, error) {
	p.mu.Lock()
	mp, ok := p.vols[name]
	p.mu.Unlock()
	if ok {
		return mp, nil
	}
	return p.ensure(name)
}

func mountpoint(name string) string {
	return filepath.Join(dataRoot, filepath.Base(name))
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/vnd.docker.plugins.v1.1+json")
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Print(err)
	}
}
