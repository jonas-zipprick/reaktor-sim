package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const prevShiftLinkName = "vorschicht"

// linkToPrevShift links outDir/vorschicht to prevOutDir for browsing exported traces.
// Unix: relative symlink. Windows: directory junction (no symlink privilege).
// If both fail, writes vorschicht.path with a relative path instead.
func linkToPrevShift(outDir, prevOutDir string) error {
	linkPath := filepath.Join(outDir, prevShiftLinkName)
	_ = os.Remove(linkPath)
	_ = os.Remove(linkPath + ".path")

	prevAbs, err := filepath.Abs(prevOutDir)
	if err != nil {
		return err
	}
	info, err := os.Stat(prevAbs)
	if err != nil {
		return fmt.Errorf("prev shift dir: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("prev shift path is not a directory: %s", prevAbs)
	}

	if runtime.GOOS == "windows" {
		if err := linkToPrevShiftWindows(linkPath, prevAbs); err == nil {
			return nil
		} else if err := writePrevShiftPathFile(linkPath, outDir, prevAbs); err == nil {
			return nil
		} else {
			return err
		}
	}

	rel, err := filepath.Rel(outDir, prevOutDir)
	if err != nil {
		return err
	}
	if err := os.Symlink(rel, linkPath); err != nil {
		return writePrevShiftPathFile(linkPath, outDir, prevAbs)
	}
	return nil
}

func linkToPrevShiftWindows(linkPath, prevAbs string) error {
	// Junction points work without SeCreateSymbolicLinkPrivilege (Developer Mode / admin).
	cmd := exec.Command("cmd", "/c", "mklink", "/J", linkPath, prevAbs)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mklink /J: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

func writePrevShiftPathFile(linkPath, outDir, prevAbs string) error {
	rel, err := filepath.Rel(outDir, prevAbs)
	if err != nil {
		return err
	}
	return os.WriteFile(linkPath+".path", []byte(rel+"\n"), 0o644)
}

// resolvePrevShiftDir returns the previous shift export directory for linkPath.
func resolvePrevShiftDir(linkPath string) (string, error) {
	if runtime.GOOS == "windows" {
		if target, err := windowsJunctionTarget(linkPath); err == nil {
			return target, nil
		}
	}
	if resolved, err := filepath.EvalSymlinks(linkPath); err == nil {
		absResolved, err1 := filepath.Abs(resolved)
		absLink, err2 := filepath.Abs(linkPath)
		if err1 == nil && err2 == nil && absResolved != absLink {
			return absResolved, nil
		}
		if err1 != nil {
			return resolved, nil
		}
	}
	data, err := os.ReadFile(linkPath + ".path")
	if err != nil {
		return "", err
	}
	base, err := filepath.Abs(filepath.Dir(linkPath))
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(base, strings.TrimSpace(string(data)))), nil
}

func windowsJunctionTarget(linkPath string) (string, error) {
	abs, err := filepath.Abs(linkPath)
	if err != nil {
		return "", err
	}
	out, err := exec.Command("fsutil", "reparsepoint", "query", abs).CombinedOutput()
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(line)
		if idx := strings.Index(line, `\??\`); idx >= 0 {
			return filepath.Clean(line[idx+len(`\??\`):]), nil
		}
		lower := strings.ToLower(line)
		for _, prefix := range []string{"print name:", "druckname:"} {
			if strings.HasPrefix(lower, prefix) {
				name := strings.TrimSpace(line[len(prefix):])
				return filepath.Clean(name), nil
			}
		}
	}
	return "", fmt.Errorf("junction target not found in fsutil output")
}
