// Package bridge some utils
package bridge

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-19

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

type Notice struct {
	Title string `json:"title"`
	URL   string `json:"url"`
	Date  string `json:"date"`
}

func (n Notice) ContentHash() string {
	title := strings.TrimSpace(n.Title)
	url := strings.TrimSpace(n.URL)
	content := title + "|" + url

	logicHash := fmt.Sprintf("%x", sha256.Sum256([]byte(content)))
	return logicHash
}

type Action string

const (
	ActionList     Action = "list"
	ActionDetail   Action = "detail"
	ActionDownload Action = "download"
)

func (c Client) IsValid() bool {
	return SupportedClients[c]
}

type FetchOptions struct {
	Client   Client
	Account  string
	Password string
	Extra    map[string]any
}

func FetchFromPython(opts *FetchOptions) ([]Notice, error) {
	extraStr := "{}"
	if opts.Extra != nil {
		bytes, err := json.Marshal(opts.Extra)
		if err != nil {
			return nil, fmt.Errorf("marshal extra options failed: %w", err)
		}
		extraStr = string(bytes)
	}

	cmd := exec.Command("python3", "scripts/catcher.py",
		string(opts.Client), opts.Account, opts.Password,
		"--action", "list",
		"--extra", extraStr,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Python 报错了！错误码: %d", exitErr.ExitCode())
			log.Printf("错误堆栈: %s", string(exitErr.Stderr))
		} else {
			log.Printf("根本没跑起来: %v\n", err)
		}
		return nil, err
	}

	var notices []Notice
	err = json.Unmarshal(output, &notices)
	if err != nil {
		log.Printf("fail to analyse JSON: %v", err)
		log.Printf("Python output: %s", string(output))
		return nil, err
	}

	return notices, nil
}

type DetailOptions struct {
	Client   Client
	Account  string
	Password string
	URL      string
	Extra    map[string]any
}

type Detail struct {
	Body        string       `json:"html"`
	Attachments []Attachment `json:"attachments"`
}

type Attachment struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

func FetchDetailFromPython(opts *DetailOptions) (*Detail, error) {
	extraStr := "{}"
	if opts.Extra != nil {
		bytes, err := json.Marshal(opts.Extra)
		if err != nil {
			return nil, fmt.Errorf("marshal extra options failed: %w", err)
		}
		extraStr = string(bytes)
	}

	cmd := exec.Command("python3", "scripts/catcher.py",
		string(opts.Client), opts.Account, opts.Password,
		"--action", "detail",
		"--url", opts.URL,
		"--extra", extraStr,
	)

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Python 报错了！错误码: %d", exitErr.ExitCode())
			log.Printf("错误堆栈: %s", string(exitErr.Stderr))
		} else {
			log.Printf("根本没跑起来: %v\n", err)
		}
		return nil, err
	}

	var detail Detail
	err = json.Unmarshal(output, &detail)
	if err != nil {
		log.Printf("fail to analyse JSON: %v", err)
		log.Printf("Python output: %s", string(output))
		return nil, err
	}

	return &detail, nil
}

type DownloadOptions struct {
	Client   Client
	Account  string
	Password string
	URL      string
	MaxSize  int
	SavePath string
	Referer  string
	Extra    map[string]any
}

func DownloadFromPython(opts *DownloadOptions) error {
	args := []string{
		"scripts/catcher.py",
		string(opts.Client), opts.Account, opts.Password,
		"--action", "download",
		"--url", opts.URL,
		"--save-path", opts.SavePath,
	}

	// extra
	extraStr := "{}"
	if opts.Extra != nil {
		bytes, err := json.Marshal(opts.Extra)
		if err != nil {
			return fmt.Errorf("marshal extra options failed: %w", err)
		}
		extraStr = string(bytes)
	}
	args = append(args, "--extra", extraStr)

	// max size
	if opts.MaxSize > 0 {
		args = append(args, "--max-size", strconv.Itoa(opts.MaxSize))
	}

	// Referer
	if opts.Referer != "" {
		args = append(args, "--referer", opts.Referer)
	}

	cmd := exec.Command("python3", args...)

	_, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			log.Printf("Python 报错了！错误码: %d", exitErr.ExitCode())
			log.Printf("错误堆栈: %s", string(exitErr.Stderr))
		} else {
			log.Printf("根本没跑起来: %v\n", err)
		}
		return err
	}

	return nil
}
