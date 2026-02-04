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

from .base import BaseClient
from lxml import etree, html
from lxml_html_clean import Cleaner
import re

class CMathcClient(BaseClient):
    client_id = "cmathc"

    def __init__(self, username, password, extra) -> None:
        super().__init__(username=username, password=password, extra=extra)

    def fetch(self):
        self.logger.debug("cmathc client start")
        resp = self.session.get("https://www.cmathc.org.cn/news/")
        noti_html = etree.HTML(resp.text)
        noti_content = noti_html.xpath('//div/ul[@class="newslist ny"]/li')

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
                full_url = f"https://www.cmathc.org.cn{raw_href}"

                # publish date
                date_text = li.xpath('.//span/text()')
                clean_date = date_text[0].strip() if date_text else ""

                results.append(
                    {"title": clean_title, "url": full_url, "date": clean_date}
                )

            except Exception as e:
                self.logger.warning(f"some error occur when analyse notification.{e}")
                continue

        return results

    def fetch_detail(self, url):
        body_content = self.session.get(url).text
        tree = html.fromstring(body_content)

        container = tree.xpath('//div[@class="article_txt"]')
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

        base_url = "https://www.cmathc.org.cn"
        content_html = content_html.replace('src="/', f'src="{base_url}/')
        content_html = content_html.replace('href="/', f'href="{base_url}/')
        content_html = content_html.replace("\\r\\n", "").replace("\\n", "")
        content_html = content_html.replace("\r", "").replace("\n", "")
        content_html = re.sub(r'>\s+<', '><', content_html)
        content_html = content_html.strip()

        return {"html": content_html, "attachments": []}
