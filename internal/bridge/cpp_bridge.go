// Copyright 2026 Czy_4201b
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
