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

from lxml_html_clean import Cleaner
import re
from lxml import etree, html
from urllib.parse import urlparse, parse_qs
from .base import BaseClient

class BUPTClient(BaseClient):
    client_id = "bupt"
    
    def __init__(self, username, password, extra) -> None:
        super().__init__(username=username, password=password, extra=extra)

    def login(self):
        self.logger.info("Start access my.bupt.edu.cn...")
        resp = self.session.get("http://my.bupt.edu.cn")
        assert resp.status_code == 200, "server error!"

        login_html = etree.HTML(resp.text)

        # get excution
        execution = login_html.xpath(
            '//*[@id="loginForm"]/*/input[@name="execution"]/@value'
        )[0]

        # get owner
        query = urlparse(resp.url).query
        service_url = parse_qs(query)["service"][0]
        inner_query = urlparse(service_url).query
        inner_param = parse_qs(inner_query)
        owner = inner_param["owner"][0]

        self.logger.debug(f"execution len: {len(execution)}, owner: {owner}")

        # build payload
        payload = {
            "username": self.username,
            "password": self.password,
            "submit": "登录",
            "type": "username_password",
            "execution": execution,
            "_eventId": "submit",
        }

        # request login ticket
        self.session.post(
            url=f"https://auth.bupt.edu.cn/authserver/login?service=http%3A%2F%2Fmy.bupt.edu.cn%2Fsystem%2Fresource%2Fcode%2Fauth%2Fclogin.jsp%3Fowner%3D{owner}",
            data=payload,
            allow_redirects=True,
        )

        # entrance
        entrance = self.session.get("http://my.bupt.edu.cn/index.jsp?null")
        entrance_html = etree.HTML(entrance.text)
        script_content = entrance_html.xpath("//div/script/text()")[0]
        homepage_path = re.search(r'"(.*?)"', script_content).group(1)
        homepage_url = f"http://my.bupt.edu.cn/{homepage_path}"
        self.logger.info("Get homepage url.")

        # homepage
        homepage = self.session.get(homepage_url)
        assert homepage.status_code == 200, (
            f"Fail to login in homepage, status_code: {homepage.status_code}"
        )
        self.logger.info("Success get login ticket.")

        # save cookies
        self._save_cookies()

    def isLogin(self) -> bool:
        try:
            resp = self.session.get(
                "http://my.bupt.edu.cn/",
                allow_redirects=False
            )
            return resp.status_code == 200
        except Exception:
            return False

    def fetch(self):
        self._ensure_auth()

        # 校内通知 wbtreeid: 115
        noti = self.session.get(
            "http://my.bupt.edu.cn/list.jsp?urltype=tree.TreeTempUrl&wbtreeid=1154"
        )
        assert noti.status_code == 200, f"error status_code: {noti.status_code}"

        noti_html = etree.HTML(noti.text)
        noti_content = noti_html.xpath('//div/ul[@class="newslist list-unstyled"]/li')

        results = []
        for li in noti_content:
            try:
                a_tag = li.xpath(".//a")[0]

                # title
                title = a_tag.xpath("./@title")
                if not title:
                    self.logger.debug("The notification has no title")
                    title = a_tag.xpath("./text()")
                clean_title = title[0].strip() if title else "未知标题"

                # url
                raw_href = a_tag.xpath("./@href")[0]
                full_url = f"http://my.bupt.edu.cn/{raw_href}"

                # publish date
                date_text = li.xpath('.//span[@class="time"]/text()')
                clean_date = date_text[0].strip() if date_text else ""

                results.append(
                    {"title": clean_title, "url": full_url, "date": clean_date}
                )
            except Exception as e:
                self.logger.warning(f"some error occur when analyse notification.{e}")
                continue

        return results

    def fetch_detail(self, url):
        detail_html = self.get_html(url)
        tree = html.fromstring(detail_html)

        # body text
        container = tree.xpath('//div[@class="v_news_content"]')
        if len(container) == 0:
            return {"html": "<p>内容解析失败</p>", "attachments": []}

        container = container[0]

        cleaner = Cleaner(
            scripts=True,  # remove <script>
            javascript=True,  # remove onclick
            comments=True,  # remove HTML comments
            style=True,  # remove <style>
            links=True,  # remove <link>
            meta=True,  # remove <meta>
            page_structure=False,  # save div
            safe_attrs_only=True,  # save attrs like src
            safe_attrs=set(["src", "href", "title", "width", "height"]),
        )

        cleaned_node = cleaner.clean_html(container)
        content_html = etree.tostring(cleaned_node, encoding="unicode", method="html")

        base_url = "http://my.bupt.edu.cn"
        content_html = content_html.replace('src="/', f'src="{base_url}/')
        content_html = content_html.replace('href="/', f'href="{base_url}/')
        content_html = content_html.replace("\\r\\n", "").replace("\\n", "")
        content_html = content_html.replace("\r", "").replace("\n", "")
        content_html = re.sub(r'>\s+<', '><', content_html)
        content_html = content_html.strip()

        # attachments
        attachments = tree.xpath('//div[@class="battch"]/ul/li')

        results = []
        for li in attachments:
            try:
                a_tag = li.xpath(".//a")[0]

                title = a_tag.xpath("./text()")
                clean_title = title[0].strip() if title else "未知标题"

                # url
                raw_href = a_tag.xpath("./@href")[0]
                full_url = f"http://my.bupt.edu.cn{raw_href}"

                results.append({"title": clean_title, "url": full_url})
            except Exception as e:
                self.logger.warning(f"some error occur when analyse notification.{e}")
                continue

        return {"html": content_html, "attachments": results}

