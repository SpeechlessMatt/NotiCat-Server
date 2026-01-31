package global

// Author: Czy_4201b <speechlessmatt@qq.com>
// Created: 2026-01-21

var (
	/*
	务必记得修改Jwt密钥，否则服务器有安全风险
	*/
	JwtSecret = []byte("NotiCat")
	/*
	SMTPSERVER: SMTP 服务器地址，可以填 URL 地址也可以填 "163" "qq" 这种，
	具体可以查看 mail 里面的 cpp 代码,建议填 URL ,最稳妥
	*/
	SMTPSERVER = ""
	/*
	ACCOUNT: 你的邮箱账号
	*/
	ACCOUNT    = ""
	/*
	AUTHCODE：你的邮箱 STMP 授权码
	*/
	AUTHCODE   = ""
)
