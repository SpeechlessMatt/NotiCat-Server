from curl_cffi import requests
import pickle
import os
import logging


class BaseClient:
    def __init__(self, name, username, password, extra={}, cookie_dir="cookies"):
        self.name = name  # 比如 "BUPT"
        self.username = username
        self.password = password
        self.extra = extra
        self.cookie_path = os.path.join(cookie_dir, f"{name}_{username}.pkl")
        self.session = requests.Session(impersonate="chrome131")
        self.session.headers.update(
            {
                "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
                "Accept-Language": "zh-CN,zh;q=0.9",
            }
        )

        self.logger = logging.getLogger(name)

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

    def fetch_detail(self, url):
        raise NotImplementedError("fetch detail NotImplementedError")
