// Supabase example: a self-hosted stack built with containerkit.
//
// The SQL, gateway configs, and key scripts under docker/ come from
// https://supabase.com/docs/guides/self-hosting/docker (self-hosted/v0.8.4).
// The project itself is Go: services, health checks, dependencies, published
// ports, and the envoy/kong/caddy/nginx, postgres, pooler, storage, and logs
// choices are containerkit calls in stack.go.
//
// Ctrl+C kills the stack and deletes the temp directory the volume files
// were written to. That directory is outside this repo so Docker is not asked
// to mount the project path.
package main

import (
	"bytes"
	"context"
	"embed"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	. "github.com/aptd3v/containerkit/pkg/kit"
)

//go:embed docker
//go:embed docker/.env.example
//go:embed docker/volumes/db/_supabase.sql
var dockerFiles embed.FS

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	fs := flag.NewFlagSet("supabase", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)
	gateway := fs.String("gateway", "envoy", "api gateway: envoy, kong, caddy, or nginx")
	postgres := fs.String("postgres", "17", "postgres major version: 15 or 17")
	pooler := fs.String("pooler", "supavisor", "connection pooler: supavisor or pgbouncer")
	storage := fs.String("storage", "file", "storage backend: file, s3, or rustfs")
	logsOn := fs.Bool("logs", false, "start logflare and vector")
	if err := fs.Parse(os.Args[1:]); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}

	dir, err := os.MkdirTemp("", "containerkit-supabase-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)
	if err := materialize(dir); err != nil {
		return err
	}
	env, err := prepareEnv(dir)
	if err != nil {
		return err
	}
	project, err := setupStack(stackConfig{
		dir:      dir,
		env:      env,
		gateway:  *gateway,
		postgres: *postgres,
		pooler:   *pooler,
		storage:  *storage,
		logs:     *logsOn,
	})
	if err != nil {
		return err
	}
	if _, err := project.Project(); err != nil {
		return fmt.Errorf("project: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	log := NewLogger(Color(true), OutputStream(os.Stdout))
	cli, err := Connect(Log(log))
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer cli.Close()

	defer func() {
		downCtx := context.Background()
		if err := project.Kill(downCtx, cli.Composer(), StopSignal(SIGKILL), RemoveOrphans(true)); err != nil {
			fmt.Fprintf(os.Stderr, "kill: %v\n", err)
		}
		if err := project.Down(downCtx, cli.Composer(), RemoveOrphans(true), Volumes(true), Log(log)); err != nil {
			fmt.Fprintf(os.Stderr, "down: %v\n", err)
		}
	}()

	// Attached Up streams container stdout and stderr as soon as each
	// container starts, including while health checks are still running.
	if err := project.Up(ctx, cli.Composer(), Log(log), RemoveOrphans(true), Wait(true), Timeout(5*time.Minute)); err != nil && !errors.Is(err, context.Canceled) {
		return fmt.Errorf("up: %w", err)
	}
	return nil
}

func materialize(dir string) error {
	err := fs.WalkDir(dockerFiles, "docker", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel("docker", path)
		if err != nil {
			return err
		}
		dest := filepath.Join(dir, rel)
		if entry.IsDir() {
			return os.MkdirAll(dest, 0o755)
		}
		data, err := dockerFiles.ReadFile(path)
		if err != nil {
			return err
		}
		mode := os.FileMode(0o644)
		if filepath.Ext(path) == ".sh" {
			mode = 0o755
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return os.WriteFile(dest, data, mode)
	})
	if err != nil {
		return fmt.Errorf("embed: %w", err)
	}
	// Directory embeds skip names that start with "." or "_".
	for _, name := range []string{".env.example", "volumes/db/_supabase.sql"} {
		data, err := dockerFiles.ReadFile("docker/" + name)
		if err != nil {
			return fmt.Errorf("embed %s: %w", name, err)
		}
		if err := os.MkdirAll(filepath.Dir(filepath.Join(dir, name)), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0o644); err != nil {
			return err
		}
	}
	for _, empty := range []string{"volumes/db/data", "volumes/snippets", "volumes/storage"} {
		if err := os.MkdirAll(filepath.Join(dir, empty), 0o755); err != nil {
			return err
		}
	}
	return nil
}

func prepareEnv(dir string) (map[string]string, error) {
	src := filepath.Join(dir, ".env.example")
	dst := filepath.Join(dir, ".env")
	data, err := os.ReadFile(src)
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return nil, err
	}
	// add-new-auth-keys.sh uncomments JWT settings in a compose file after it
	// writes the keys. The stack reads those keys from .env, so the file only
	// has to exist for the script.
	stub := []byte("#GOTRUE_JWT_KEYS:\n#API_JWT_JWKS:\n#JWT_JWKS:\n#SUPABASE_JWKS:\n")
	if err := os.WriteFile(filepath.Join(dir, "docker-compose.yml"), stub, 0o644); err != nil {
		return nil, err
	}
	for _, script := range []string{
		"utils/generate-keys.sh",
		"utils/add-new-auth-keys.sh",
	} {
		cmd := exec.Command("sh", script, "--update-env")
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return nil, fmt.Errorf("%s: %w", script, err)
		}
	}
	env, err := parseEnv(dst)
	if err != nil {
		return nil, err
	}
	if env["JWT_SECRET"] == "" || env["ANON_KEY"] == "" {
		return nil, fmt.Errorf("generated .env is missing JWT_SECRET or ANON_KEY")
	}
	return env, nil
}

func parseEnv(path string) (map[string]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	env := map[string]string{}
	for _, line := range bytes.Split(data, []byte("\n")) {
		text := strings.TrimSpace(string(line))
		if text == "" || strings.HasPrefix(text, "#") {
			continue
		}
		key, val, ok := strings.Cut(text, "=")
		if !ok || strings.TrimSpace(key) == "" {
			continue
		}
		env[strings.TrimSpace(key)] = unquote(val)
	}
	return env, nil
}

func unquote(v string) string {
	v = strings.TrimSpace(v)
	if len(v) >= 2 {
		if (v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'') {
			return v[1 : len(v)-1]
		}
	}
	return v
}
