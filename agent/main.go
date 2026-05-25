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

// websiteBlockCategories lists domains per category for hosts-file blocking.
var websiteBlockCategories = map[string][]string{
	"social_media": {
		"facebook.com", "www.facebook.com",
		"x.com", "twitter.com", "www.twitter.com",
		"instagram.com", "www.instagram.com",
		"tiktok.com", "www.tiktok.com",
		"snapchat.com", "www.snapchat.com",
		"linkedin.com", "www.linkedin.com",
		"reddit.com", "www.reddit.com",
		"pinterest.com", "www.pinterest.com",
		"tumblr.com", "www.tumblr.com",
		"discord.com", "www.discord.com",
		"telegram.org", "web.telegram.org",
	},
	"video_streaming": {
		"youtube.com", "www.youtube.com", "youtu.be",
		"netflix.com", "www.netflix.com",
		"primevideo.com", "www.primevideo.com",
		"hotstar.com", "www.hotstar.com",
		"jiocinema.com", "www.jiocinema.com",
		"hulu.com", "www.hulu.com",
		"disneyplus.com", "www.disneyplus.com",
		"twitch.tv", "www.twitch.tv",
		"vimeo.com", "www.vimeo.com",
		"dailymotion.com", "www.dailymotion.com",
		"zee5.com", "www.zee5.com",
		"sonyliv.com", "www.sonyliv.com",
		"mxplayer.in", "www.mxplayer.in",
	},
	"shopping": {
		"amazon.com", "www.amazon.com",
		"amazon.in", "www.amazon.in",
		"flipkart.com", "www.flipkart.com",
		"myntra.com", "www.myntra.com",
		"ebay.com", "www.ebay.com",
		"alibaba.com", "www.alibaba.com",
		"aliexpress.com", "www.aliexpress.com",
		"snapdeal.com", "www.snapdeal.com",
		"meesho.com", "www.meesho.com",
		"ajio.com", "www.ajio.com",
		"nykaa.com", "www.nykaa.com",
	},
	"entertainment": {
		"spotify.com", "open.spotify.com",
		"soundcloud.com", "www.soundcloud.com",
		"crunchyroll.com", "www.crunchyroll.com",
		"gaana.com", "www.gaana.com",
		"jiosaavn.com", "www.jiosaavn.com",
		"wynk.in", "www.wynk.in",
	},
}

func hostsFilePath() string {
	if runtime.GOOS == "windows" {
		return `C:\Windows\System32\drivers\etc\hosts`
	}
	return "/etc/hosts"
}

const hostsBlockTag = "# firewall-agent-block"

// blockWebsiteCategory appends 0.0.0.0 entries for every domain in the category
// into the system hosts file. Each line is tagged so they can be removed later.
func blockWebsiteCategory(category string) error {
	domains, ok := websiteBlockCategories[category]
	if !ok {
		return fmt.Errorf("unknown category %q", category)
	}

	path := hostsFilePath()
	existing, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read hosts file: %w", err)
	}

	var sb strings.Builder
	for _, domain := range domains {
		marker := fmt.Sprintf("%s:%s", hostsBlockTag, domain)
		if strings.Contains(string(existing), marker) {
			continue // already blocked
		}
		sb.WriteString(fmt.Sprintf("0.0.0.0 %s %s\n", domain, marker))
	}

	if sb.Len() == 0 {
		return nil // nothing new to add
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open hosts file: %w", err)
	}
	defer f.Close()
	_, err = f.WriteString(sb.String())
	return err
}

// unblockWebsiteCategory removes hosts-file entries that were added by blockWebsiteCategory.
func unblockWebsiteCategory(category string) error {
	if _, ok := websiteBlockCategories[category]; !ok {
		return fmt.Errorf("unknown category %q", category)
	}

	path := hostsFilePath()
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read hosts file: %w", err)
	}

	var kept []string
	for _, line := range strings.Split(string(data), "\n") {
		if strings.Contains(line, hostsBlockTag) {
			taggedCategory := false
			for _, domain := range websiteBlockCategories[category] {
				if strings.Contains(line, hostsBlockTag+":"+domain) {
					taggedCategory = true
					break
				}
			}
			if taggedCategory {
				continue
			}
		}
		kept = append(kept, line)
	}

	return ioutil.WriteFile(path, []byte(strings.Join(kept, "\n")), 0644)
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
			case "website.block_category":
				var payload struct {
					Category string `json:"category"`
				}
				if err := json.Unmarshal([]byte(cmd.Payload), &payload); err != nil {
					result = "invalid payload: " + err.Error()
				} else if err := blockWebsiteCategory(payload.Category); err != nil {
					result = err.Error()
				} else {
					success = true
					result = fmt.Sprintf("category %q blocked via hosts file", payload.Category)
				}
			case "website.unblock_category":
				var payload struct {
					Category string `json:"category"`
				}
				if err := json.Unmarshal([]byte(cmd.Payload), &payload); err != nil {
					result = "invalid payload: " + err.Error()
				} else if err := unblockWebsiteCategory(payload.Category); err != nil {
					result = err.Error()
				} else {
					success = true
					result = fmt.Sprintf("category %q unblocked from hosts file", payload.Category)
				}
			default:
				result = "Unknown command type"
			}
			reportResult(cfg, cmd.ID, success, result)
		}
		time.Sleep(30 * time.Second)
	}
}
