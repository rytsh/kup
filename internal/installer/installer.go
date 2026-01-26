package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/rytsh/kup/internal/config"
)

// Tool represents an installable tool
type Tool struct {
	Name        string
	Description string
	GetURL      func(cfg *config.Config) string
	GetCommand  func(cfg *config.Config) string
	Explanation string
	PostInstall func(binPath string) error
}

// DownloadProgress represents download progress
type DownloadProgress struct {
	Tool       string
	Downloaded int64
	Total      int64
	Done       bool
	Error      error
}

// GetTools returns all available tools
func GetTools() []Tool {
	return []Tool{
		KubectlTool(),
		K9sTool(),
		KindTool(),
	}
}

// KubectlTool returns the kubectl tool definition
func KubectlTool() Tool {
	return Tool{
		Name:        "kubectl",
		Description: "Kubernetes command-line tool for running commands against clusters",
		GetURL: func(cfg *config.Config) string {
			os := cfg.GetOS()
			arch := cfg.GetArchitecture()
			// Get latest stable version
			return fmt.Sprintf(
				"https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/%s/%s/kubectl",
				os, arch,
			)
		},
		GetCommand: func(cfg *config.Config) string {
			os := cfg.GetOS()
			arch := cfg.GetArchitecture()
			binPath := cfg.BinPath
			return fmt.Sprintf(
				`curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/%s/%s/kubectl" && \
chmod +x kubectl && \
mv kubectl %s/kubectl`,
				os, arch, binPath,
			)
		},
		Explanation: `This command will:
1. Fetch the latest stable Kubernetes version number
2. Download the kubectl binary for your OS and architecture
3. Make it executable (chmod +x)
4. Move it to your bin directory`,
	}
}

// K9sTool returns the k9s tool definition
func K9sTool() Tool {
	return Tool{
		Name:        "k9s",
		Description: "Terminal UI to interact with your Kubernetes clusters",
		GetURL: func(cfg *config.Config) string {
			os := cfg.GetOS()
			arch := cfg.GetArchitecture()
			osName := os
			if os == "darwin" {
				osName = "Darwin"
			} else if os == "linux" {
				osName = "Linux"
			}
			archName := arch
			if arch == "amd64" {
				archName = "amd64"
			} else if arch == "arm64" {
				archName = "arm64"
			}
			return fmt.Sprintf(
				"https://github.com/derailed/k9s/releases/latest/download/k9s_%s_%s.tar.gz",
				osName, archName,
			)
		},
		GetCommand: func(cfg *config.Config) string {
			os := cfg.GetOS()
			arch := cfg.GetArchitecture()
			binPath := cfg.BinPath
			osName := os
			if os == "darwin" {
				osName = "Darwin"
			} else if os == "linux" {
				osName = "Linux"
			}
			archName := arch
			if arch == "amd64" {
				archName = "amd64"
			} else if arch == "arm64" {
				archName = "arm64"
			}
			return fmt.Sprintf(
				`curl -LO "https://github.com/derailed/k9s/releases/latest/download/k9s_%s_%s.tar.gz" && \
tar -xzf k9s_%s_%s.tar.gz k9s && \
chmod +x k9s && \
mv k9s %s/k9s && \
rm k9s_%s_%s.tar.gz`,
				osName, archName, osName, archName, binPath, osName, archName,
			)
		},
		Explanation: `This command will:
1. Download the latest k9s release archive for your OS and architecture
2. Extract the k9s binary from the tar.gz archive
3. Make it executable (chmod +x)
4. Move it to your bin directory
5. Clean up the downloaded archive`,
	}
}

// KindTool returns the kind tool definition
func KindTool() Tool {
	return Tool{
		Name:        "kind",
		Description: "Tool for running local Kubernetes clusters using Docker containers",
		GetURL: func(cfg *config.Config) string {
			os := cfg.GetOS()
			arch := cfg.GetArchitecture()
			return fmt.Sprintf(
				"https://kind.sigs.k8s.io/dl/latest/kind-%s-%s",
				os, arch,
			)
		},
		GetCommand: func(cfg *config.Config) string {
			os := cfg.GetOS()
			arch := cfg.GetArchitecture()
			binPath := cfg.BinPath
			return fmt.Sprintf(
				`curl -Lo kind "https://kind.sigs.k8s.io/dl/latest/kind-%s-%s" && \
chmod +x kind && \
mv kind %s/kind`,
				os, arch, binPath,
			)
		},
		Explanation: `This command will:
1. Download the latest kind binary for your OS and architecture
2. Make it executable (chmod +x)
3. Move it to your bin directory`,
	}
}

// Installer handles tool installation
type Installer struct {
	config *config.Config
	client *http.Client
}

// NewInstaller creates a new installer
func NewInstaller(cfg *config.Config) *Installer {
	transport := http.DefaultTransport.(*http.Transport).Clone()

	if cfg.ProxyURL != "" {
		// Note: proxy configuration would go here
		// For simplicity, we'll skip proxy setup in this example
	}

	return &Installer{
		config: cfg,
		client: &http.Client{
			Transport: transport,
			Timeout:   time.Duration(cfg.Timeout) * time.Second,
		},
	}
}

// Install downloads and installs a tool
func (i *Installer) Install(ctx context.Context, tool Tool, progressCh chan<- DownloadProgress) error {
	// Ensure bin directory exists
	if err := os.MkdirAll(i.config.BinPath, 0755); err != nil {
		return fmt.Errorf("failed to create bin directory: %w", err)
	}

	// For now, we'll use a simplified direct download approach
	// In production, you'd want to handle the shell commands or implement proper download logic

	url := i.getDirectDownloadURL(tool)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := i.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status: %s", resp.Status)
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", tool.Name+"-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	// Download with progress
	total := resp.ContentLength
	downloaded := int64(0)

	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			_, writeErr := tmpFile.Write(buf[:n])
			if writeErr != nil {
				return fmt.Errorf("failed to write: %w", writeErr)
			}
			downloaded += int64(n)

			if progressCh != nil {
				select {
				case progressCh <- DownloadProgress{
					Tool:       tool.Name,
					Downloaded: downloaded,
					Total:      total,
				}:
				default:
				}
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("failed to read: %w", err)
		}
	}

	tmpFile.Close()

	// Move to destination
	destPath := filepath.Join(i.config.BinPath, tool.Name)

	// Handle tar.gz files (like k9s)
	if tool.Name == "k9s" {
		if err := i.extractTarGz(tmpFile.Name(), destPath); err != nil {
			return err
		}
	} else {
		// Direct binary
		if err := os.Rename(tmpFile.Name(), destPath); err != nil {
			// If rename fails (cross-device), try copy
			if err := i.copyFile(tmpFile.Name(), destPath); err != nil {
				return fmt.Errorf("failed to move binary: %w", err)
			}
		}
	}

	// Make executable
	if err := os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("failed to make executable: %w", err)
	}

	// Run post-install if defined
	if tool.PostInstall != nil {
		if err := tool.PostInstall(destPath); err != nil {
			return fmt.Errorf("post-install failed: %w", err)
		}
	}

	if progressCh != nil {
		progressCh <- DownloadProgress{
			Tool:       tool.Name,
			Downloaded: total,
			Total:      total,
			Done:       true,
		}
	}

	return nil
}

func (i *Installer) getDirectDownloadURL(tool Tool) string {
	os := i.config.GetOS()
	arch := i.config.GetArchitecture()

	switch tool.Name {
	case "kubectl":
		// We need to get the latest version first, but for simplicity use a recent stable
		return fmt.Sprintf(
			"https://dl.k8s.io/release/v1.29.0/bin/%s/%s/kubectl",
			os, arch,
		)
	case "k9s":
		osName := os
		if os == "darwin" {
			osName = "Darwin"
		} else if os == "linux" {
			osName = "Linux"
		}
		archName := arch
		return fmt.Sprintf(
			"https://github.com/derailed/k9s/releases/latest/download/k9s_%s_%s.tar.gz",
			osName, archName,
		)
	case "kind":
		return fmt.Sprintf(
			"https://kind.sigs.k8s.io/dl/v0.20.0/kind-%s-%s",
			os, arch,
		)
	}

	return tool.GetURL(i.config)
}

func (i *Installer) copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func (i *Installer) extractTarGz(src, destBinary string) error {
	// For simplicity, we'll use the system tar command
	// In production, you might want to use archive/tar
	cmd := fmt.Sprintf("tar -xzf %s -C %s k9s && mv %s/k9s %s",
		src,
		filepath.Dir(destBinary),
		filepath.Dir(destBinary),
		destBinary,
	)
	_ = cmd // We'll handle this with exec in a real implementation

	// For now, return an error indicating we need shell execution
	return fmt.Errorf("tar extraction requires shell execution - use the command shown")
}

// IsInstalled checks if a tool is already installed
func (i *Installer) IsInstalled(tool Tool) bool {
	binPath := filepath.Join(i.config.BinPath, tool.Name)
	_, err := os.Stat(binPath)
	return err == nil
}

// GetInstalledVersion attempts to get the version of an installed tool
func (i *Installer) GetInstalledVersion(tool Tool) string {
	binPath := filepath.Join(i.config.BinPath, tool.Name)
	if _, err := os.Stat(binPath); err != nil {
		return "not installed"
	}
	return "installed"
}

// GetSystemInfo returns system information
func GetSystemInfo() (string, string) {
	return runtime.GOOS, runtime.GOARCH
}
