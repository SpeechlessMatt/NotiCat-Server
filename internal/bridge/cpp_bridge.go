package bridge

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-21

import (
	"log"
	"os/exec"
)

type SendOptions struct {
	SMTPServer  string
	Account     string
	AuthCode    string
	From        string
	To          string
	Subject     string
	Body        string
	Attachments []string
}

func SendMail(opts *SendOptions) error {
	args := []string{
		"--smtp-server", opts.SMTPServer,
		"--user-account", opts.Account,
		"--auth-code", opts.AuthCode,
		"--from", opts.From,
		"--to", opts.To,
		"--subject", opts.Subject,
		opts.Body,
	}

	for _, path := range opts.Attachments {
		args = append(args, "--attachment", path)
	}

	cmd := exec.Command("mail/bin/send", args...)

	_, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode := exitError.ExitCode()
			exitStr := exitError.Stderr
			log.Printf("C++程序报错 (Exit Code %d): %s", exitCode, string(exitStr))
			return err
		}
		log.Printf("启动失败: %v", err)
		return err
	}

	return nil
}
