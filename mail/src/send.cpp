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

#include <curl/curl.h>
#include <getopt.h>
#include <unistd.h>

#include <cstring>
#include <exception>
#include <filesystem>
#include <iostream>
#include <sstream>
#include <string>
#include <unordered_map>
#include <vector>

#include "./base64_encode.h"

struct Attachment {
    std::string fileName;
    std::string content;
};

struct Config {
    std::string smtp_server = "";
    std::string user = "";
    std::string auth_code = "";
    std::string from = "";
    std::string to = "";
    std::string subject = "";
    std::vector<Attachment> attachments;
    std::string body = "";
};

void print_help(const char *prog_name) {
    std::cout << "Usage: " << prog_name
              << " -s <name/url> -u <account> -a <code> -f <email_from> -t <email_to> "
                 "[OPTIONS] <BODY>\n";
}

void print_detail_help(const char *prog_name) {
    std::cout << "NotiCat email CLI\n\n"
              << "Usage: " << prog_name
              << " -s <name/url> -u <account> -a <code> -f <email_from> -t <email_to> "
                 "[OPTIONS] <BODY>\n\n"
              << "Required:\n"
              << "   -s, --smtp-server <server_name/url>    Smtp Server name (e.g.,163) or url\n"
              << "   -u, --user-account <name>              Account on the smtp server\n"
              << "   -a, --auth-code <code>                 Auth code of the account\n"
              << "   -f, --from <email_from>                Sender of the email\n"
              << "   -t, --to <email_to>                    Recipients of the email\n"
              << "   <BODY>                                 Body of the email(HTML)\n"
              << "\n"
              << "Options:\n"
              << "   -S, --subject                          Subject for the email\n"
              << "   -A, --attachment                       Attachments for the email(e.g.,-A "
                 "1.txt -A 2.txt)\n"
              << "   -h, --help                             Show this help\n";
}

// smtp name to smtp url
static const std::unordered_map<std::string, std::string> smtp_server_url_map = {
    {"163", "smtps://smtp.163.com:465"},
    {"126", "smtps://smtp.126.com:465"},
    {"qq", "smtps://smtp.qq.com:465"},
};

int main(int argc, char *argv[]) {
    if (argc == 1) {
        print_help(argv[0]);
        return 1;
    }

    Config config;

    static struct option long_options[] = {{"smtp-server", required_argument, 0, 's'},
                                           {"user-account", required_argument, 0, 'u'},
                                           {"auth-code", required_argument, 0, 'a'},
                                           {"from", required_argument, 0, 'f'},
                                           {"to", required_argument, 0, 't'},
                                           {"help", no_argument, 0, 'h'},
                                           {"subject", required_argument, 0, 'S'},
                                           {"attachment", required_argument, 0, 'A'},
                                           {0, 0, 0, 0}};

    int opt;
    int option_index = 0;
    opterr = 0;

    while ((opt = getopt_long(argc, argv, "s:u:a:f:t:hS:A:", long_options, &option_index)) != -1) {
        switch (opt) {
            case 's':
                config.smtp_server = optarg;
                break;
            case 'u':
                config.user = optarg;
                break;
            case 'a':
                config.auth_code = optarg;
                break;
            case 'f':
                config.from = optarg;
                break;
            case 't':
                config.to = optarg;
                break;
            case 'h':
                print_detail_help(argv[0]);
                return 0;
            case 'S':
                config.subject = optarg;
                break;
            case 'A': {
                std::string filename = std::filesystem::path(optarg).filename().string();
                try {
                    std::string content = file_to_base64(optarg);
                    config.attachments.push_back(Attachment{filename, content});
                } catch (const std::exception &e) {
                    std::cerr << "cannot read the file:" << optarg << '\n';
                    return 1;
                }
                break;
            }
            default:
                break;
        }
    }

    // missing BODY content
    if (optind >= argc) {
        print_help(argv[0]);
        return 1;
    }

    std::vector<std::string> missing_args;

    // check the required arguments
    if (config.smtp_server.empty()) missing_args.push_back("--smtp-server (-s)");
    if (config.user.empty()) missing_args.push_back("--user-account (-u)");
    if (config.auth_code.empty()) missing_args.push_back("--auth-code (-a)");
    if (config.from.empty()) missing_args.push_back("--from (-f)");
    if (config.to.empty()) missing_args.push_back("--to (-t)");

    if (!missing_args.empty()) {
        std::cerr << "error: Missing required arguments:\n";
        for (const auto &arg : missing_args) {
            std::cerr << " " << arg << "\n";
        }
        std::cerr << "\nUse --help or -h to check usage.\n";
        return 1;
    }

    // change name to url
    auto smtp_server_url = smtp_server_url_map.find(config.smtp_server);
    if (smtp_server_url != smtp_server_url_map.end()) config.smtp_server = smtp_server_url->second;

    // body content
    config.body = argv[optind];

    CURL *curl;
    CURLcode res = CURLE_OK;

    // mail content
    std::string boundary = "----=_NextPart_5D2E123456789";
    std::stringstream msg;
    msg << "From: " << config.from << "\r\n"
        << "To: " << config.to << "\r\n"
        << "Subject: " << config.subject << "\r\n"
        << "MIME-Version: 1.0" << "\r\n"
        << "Content-Type: multipart/mixed; boundary=\"" << boundary << "\"\r\n"
        << "\r\n";

    // body content
    msg << "--" << boundary << "\r\n"
        << "Content-Type: text/html; charset=utf-8" << "\r\n"
        << "Content-Transfer-Encoding: 8bit" << "\r\n"
        << "\r\n"
        << config.body << "\r\n";

    for (const auto &att : config.attachments) {
        msg << "--" << boundary << "\r\n"
            << "Content-Type: application/octet-stream; name=\"" << att.fileName << "\"\r\n"
            << "Content-Transfer-Encoding: base64" << "\r\n"
            << "Content-Disposition: attachment; filename=\"" << att.fileName << "\"\r\n"
            << "\r\n"
            << att.content << "\r\n";
    }

    // email end
    msg << "--" << boundary << "--\r\n";

    std::string payload = msg.str();

    curl = curl_easy_init();
    if (curl) {
        curl_easy_setopt(curl, CURLOPT_URL, config.smtp_server.c_str());
        curl_easy_setopt(curl, CURLOPT_USERNAME, config.user.c_str());
        curl_easy_setopt(curl, CURLOPT_PASSWORD, config.auth_code.c_str());
        curl_easy_setopt(curl, CURLOPT_MAIL_FROM, config.from.c_str());

        struct curl_slist *recipients = NULL;
        recipients = curl_slist_append(recipients, config.to.c_str());
        curl_easy_setopt(curl, CURLOPT_MAIL_RCPT, recipients);

        // 设置读取数据的回调（这里直接把字符串传过去）
        curl_easy_setopt(curl, CURLOPT_READDATA, &payload);
        curl_easy_setopt(curl, CURLOPT_UPLOAD, 1L);

        auto read_callback = [](char *ptr, size_t size, size_t nmemb, void *userp) -> size_t {
            std::string *data = (std::string *)userp;
            size_t len = data->length();
            if (len > 0) {
                size_t copy_size = (len < size * nmemb) ? len : size * nmemb;
                memcpy(ptr, data->c_str(), copy_size);
                data->erase(0, copy_size);
                return copy_size;
            }
            return 0;
        };
        curl_easy_setopt(curl, CURLOPT_READFUNCTION, +read_callback);

        // retry
        int retries = 3;
        while (retries > 0) {
            res = curl_easy_perform(curl);
            if (res == CURLE_OK) {
                std::cout << "Success send！\n";
                break;
            } else {
                std::cerr << "Fail: " << curl_easy_strerror(res) << '\n'
                          << "Remaining attemps: " << --retries << '\n';
                if (retries > 0) sleep(10);
            }
        }

        curl_slist_free_all(recipients);
        curl_easy_cleanup(curl);
    }

    return (int)res;
}
