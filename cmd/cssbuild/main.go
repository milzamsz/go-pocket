package main

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

func main() {
	candidates := tailwindCandidates()

	bin, err := resolveBinary(candidates)
	if err != nil {
		bin, err = ensureDownloadedTailwind()
		if err != nil {
			if err := runWithNPX(); err != nil {
				fmt.Fprintln(os.Stderr, "tailwindcss binary not found.")
				fmt.Fprintln(os.Stderr, "Install Tailwind CSS v4.3.0 and rerun `task css:build`.")
				fmt.Fprintln(os.Stderr, "Or set TAILWINDCSS_BIN to an absolute binary path.")
				fmt.Fprintln(os.Stderr, "npx fallback also failed:", err)
				os.Exit(1)
			}
			return
		}
	}

	cmd := exec.Command(bin, "-i", "./assets/css/input.css", "-o", "./assets/css/output.css", "--minify")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "tailwindcss build failed via %s: %v\n", bin, err)
		os.Exit(1)
	}
}

func runWithNPX() error {
	cmd := exec.Command("npx", "-y", "@tailwindcss/cli@4.3.0", "-i", "./assets/css/input.css", "-o", "./assets/css/output.css", "--minify")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func tailwindCandidates() []string {
	var out []string
	if env := os.Getenv("TAILWINDCSS_BIN"); env != "" {
		out = append(out, env)
	}

	out = append(out, "tailwindcss")

	roots := []string{
		".",
		"./bin",
		"./tools",
	}
	names := []string{"tailwindcss"}
	if runtime.GOOS == "windows" {
		names = append(names, "tailwindcss.exe")
	}

	for _, root := range roots {
		for _, name := range names {
			out = append(out, filepath.Join(root, name))
		}
	}

	return out
}

func resolveBinary(candidates []string) (string, error) {
	for _, candidate := range candidates {
		if candidate == "" {
			continue
		}

		if path, err := exec.LookPath(candidate); err == nil {
			return path, nil
		}

		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}
	return "", errors.New("not found")
}

func ensureDownloadedTailwind() (string, error) {
	fileName, urls := releaseCandidates()
	if fileName == "" || len(urls) == 0 {
		return "", errors.New("unsupported platform")
	}

	targetDir := filepath.Join(".", "tools")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("create tools dir: %w", err)
	}
	target := filepath.Join(targetDir, fileName)

	for _, url := range urls {
		if err := downloadFile(url, target); err != nil {
			continue
		}
		if runtime.GOOS != "windows" {
			if err := os.Chmod(target, 0o755); err != nil {
				return "", fmt.Errorf("chmod tailwind binary: %w", err)
			}
		}
		return target, nil
	}

	return "", errors.New("download failed")
}

func releaseCandidates() (string, []string) {
	const version = "v4.3.0"
	var file string
	switch runtime.GOOS {
	case "windows":
		if runtime.GOARCH == "amd64" {
			file = "tailwindcss-windows-x64.exe"
		} else if runtime.GOARCH == "arm64" {
			file = "tailwindcss-windows-arm64.exe"
		}
	case "linux":
		if runtime.GOARCH == "amd64" {
			file = "tailwindcss-linux-x64"
		} else if runtime.GOARCH == "arm64" {
			file = "tailwindcss-linux-arm64"
		}
	case "darwin":
		if runtime.GOARCH == "amd64" {
			file = "tailwindcss-macos-x64"
		} else if runtime.GOARCH == "arm64" {
			file = "tailwindcss-macos-arm64"
		}
	}
	if file == "" {
		return "", nil
	}

	base := "https://github.com/tailwindlabs/tailwindcss/releases/download/" + version + "/" + file
	alt := "https://github.com/tailwindlabs/tailwindcss/releases/latest/download/" + file
	return file, []string{base, alt}
}

func downloadFile(url string, dest string) error {
	client := &http.Client{Timeout: 45 * time.Second}
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "go-pocket-cssbuild/1.0")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download status %d", resp.StatusCode)
	}

	tmp := dest + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}

	if err := os.Rename(tmp, dest); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return nil
}
