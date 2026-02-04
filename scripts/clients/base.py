# Copyright 2026 Czy_4201b
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

from abc import ABCMeta
from curl_cffi import requests
import re
import pickle
import os
import logging

CLIENT_REGISTER = {}

# name rule: 
# 1. start with letter or _
# 2. all low case
# 3. allow a-z,0-9,_
KEY_PATTERN = re.compile(r'^[a-z_][a-z0-9_]*$')

class ClientMeta(ABCMeta):
    def __new__(mcs, name, bases, attrs):
        cls = super().__new__(mcs, name, bases, attrs)

        if name != "BaseClient":
            # can define client_id or auto generate: remove "Client" and lower()
            client_id = attrs.get('client_id') or name.replace("Client", "").lower()

            if not isinstance(client_id, str):
                raise TypeError(f"[{name}] client_id must be str!")
            if not KEY_PATTERN.match(client_id):
                raise ValueError(f"[{name}] client_id '{client_id}' format error!")
            if client_id in CLIENT_REGISTER:
                raise KeyError(f"[{name}] client_id '{client_id}' has been used by {CLIENT_REGISTER[client_id].__name__}!")

            cls.client_id = client_id
            CLIENT_REGISTER[client_id] = cls

        return cls

class BaseClient(metaclass=ClientMeta):
    client_id = None

    def __init__(self, username, password, extra={}, cookie_dir="cookies"):
        self.name = self.client_id
        self.username = username
        self.password = password
        self.extra = extra
        self.cookie_path = os.path.join(cookie_dir, f"{self.name}_{username}.pkl")
        self.session = requests.Session(impersonate="chrome131")
        self.session.headers.update(
            {
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
                "Accept-Language": "zh-CN,zh;q=0.9",
            }
        )

        self.logger = logging.getLogger(self.name)

        # 确保 cookie 目录存在
        os.makedirs(cookie_dir, exist_ok=True)

    def _save_cookies(self):
        if self.session.cookies:
            cookie_dict = self.session.cookies.get_dict()
            with open(self.cookie_path, "wb") as f:
                pickle.dump(cookie_dict, f)
            self.logger.debug(f"Saved {len(self.session.cookies)} cookies.")
        else:
            self.logger.warning("No cookies to save!")

    def _load_cookies(self) -> bool:
        if os.path.exists(self.cookie_path):
            with open(self.cookie_path, "rb") as f:
                cookie_dict = pickle.load(f)
            self.session.cookies.update(cookie_dict)
            return True
        return False

    def _ensure_auth(self):
        hasCookies = self._load_cookies()
        isLogin = self.isLogin()
        self.logger.debug(f"hasCookies: {hasCookies}, isLogin: {isLogin}")

        if hasCookies and isLogin:
            self.logger.info("Skip Login.")
        else:
            self.logger.info("Session out of date, ready for login...")
            self.login()

    def isLogin(self) -> bool:
        raise NotImplementedError("check login NotImplementedError")

    def login(self):
        raise NotImplementedError("login NotImplementedError")

    def fetch(self):
        raise NotImplementedError("fetch NotImplementedError")

    def get_html(self, url):
        self._ensure_auth()
        resp = self.session.get(url, timeout=10)
        resp.raise_for_status()
        return resp.text

    def download_file(self, url, save_path, referer=None, max_size=None):
        self._ensure_auth()
        try:
            headers = {}
            if referer:
                headers["Referer"] = referer
                self.logger.debug(f"With Referer download: {referer}")

            r = self.session.get(url, stream=True, timeout=30, headers=headers)
            r.raise_for_status()

            content_length = r.headers.get("Content-Length")
            if content_length and max_size:
                total_size_mb = int(content_length) / (1024 * 1024)
                if total_size_mb > max_size:
                    self.logger.error(
                        f"File too large: {total_size_mb:.2f}MB > {max_size}MB"
                    )
                    return False

            downloaded_size = 0
            with open(save_path, "wb") as f:
                for chunk in r.iter_content(chunk_size=8192):
                    if chunk:
                        f.write(chunk)
                        downloaded_size += len(chunk)

                    if max_size:
                        current_mb = downloaded_size / (1024 * 1024)
                        if current_mb > max_size:
                            self.logger.error(
                                f"Download aborted: exceeded {max_size}MB"
                            )
                            f.close()
                            if os.path.exists(save_path):
                                os.remove(save_path)
                            return False

            self.logger.info(f"Download success: {save_path}")
            return True
        except Exception as e:
            self.logger.error(f"Download error: {e}")
            return False

    def fetch_detail(self, url: str):
        self.logger.debug(f"Input url: {url}")
        raise NotImplementedError("fetch detail NotImplementedError")
