package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

type Command struct {
	ID          int64  `json:"id"`
	CommandType string `json:"command_type"`
	Payload     string `json:"payload_json"`
}

type Config struct {
	ServerURL  string `json:"server_url"`
	AgentToken string `json:"agent_token"`
}

func loadConfig() (Config, error) {
	data, err := ioutil.ReadFile("agent_config.json")
	if err != nil {
		return Config{}, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func pollCommands(cfg Config) ([]Command, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("POST", cfg.ServerURL+"/api/v1/public/agent/commands/next", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Agent-Token", cfg.AgentToken)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("server returned %d", resp.StatusCode)
	}
	var result struct {
		Item *Command `json:"item"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.Item == nil {
		return nil, nil
	}
	return []Command{*result.Item}, nil
}

func reportResult(cfg Config, cmdID int64, success bool, resultText string) error {
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/api/v1/public/agent/commands/%d/result", cfg.ServerURL, cmdID)
	body := fmt.Sprintf(`{"agent_token":"%s","success":%v,"result_text":"%s"}`,
		cfg.AgentToken, success, resultText)
	req, err := http.NewRequest("POST", url, ioutil.NopCloser(strings.NewReader(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func blockUSB() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-Command", "$path='Registry::HKEY_LOCAL_MACHINE\\SYSTEM\\CurrentControlSet\\Services\\USBSTOR'; if (-not (Test-Path $path)) { New-Item -Path $path -Force | Out-Null }; New-ItemProperty -Path $path -Name Start -PropertyType DWord -Value 4 -Force | Out-Null")
		return cmd.Run()
	case "linux":
		cmd := exec.Command("modprobe", "-r", "usb_storage")
		return cmd.Run()
	case "darwin":
		return fmt.Errorf("USB block not implemented for macOS")
	default:
		return fmt.Errorf("unsupported OS")
	}
}

func unblockUSB() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("powershell", "-Command", "$path='Registry::HKEY_LOCAL_MACHINE\\SYSTEM\\CurrentControlSet\\Services\\USBSTOR'; if (-not (Test-Path $path)) { New-Item -Path $path -Force | Out-Null }; New-ItemProperty -Path $path -Name Start -PropertyType DWord -Value 3 -Force | Out-Null")
		return cmd.Run()
	case "linux":
		cmd := exec.Command("modprobe", "usb_storage")
		return cmd.Run()
	case "darwin":
		return fmt.Errorf("USB unblock not implemented for macOS")
	default:
		return fmt.Errorf("unsupported OS")
	}
}

func main() {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Println("Config error:", err)
		os.Exit(1)
	}
	for {
		cmds, err := pollCommands(cfg)
		if err != nil {
			fmt.Println("Poll error:", err)
			time.Sleep(30 * time.Second)
			continue
		}
		for _, cmd := range cmds {
			var result string
			success := false
			switch cmd.CommandType {
			case "usb.block":
				err := blockUSB()
				if err == nil {
					success = true
					result = "USB blocked"
				} else {
					result = err.Error()
				}
			case "usb.unblock":
				err := unblockUSB()
				if err == nil {
					success = true
					result = "USB unblocked"
				} else {
					result = err.Error()
				}
			default:
				result = "Unknown command type"
			}
			reportResult(cfg, cmd.ID, success, result)
		}
		time.Sleep(30 * time.Second)
	}
}
