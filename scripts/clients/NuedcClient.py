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
import time
import re

class NuedcClient(BaseClient):
    def __init__(self, username, password, extra) -> None:
        super().__init__(username=username, password=password, extra=extra)

    def fetch(self):
        resp = self.session.get("https://www.nuedc-training.com.cn/index/news/index")
        noti_html = etree.HTML(resp.text)
        noti_content = noti_html.xpath('//*[@id="newWrap"]/ul/li')

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
                href = a_tag.xpath("./@href")[0]

                # publish date
                date = li.xpath('./@data-time')
                if len(date) != 0:
                    timestamp = int(date[0])
                    time_struct = time.localtime(timestamp)
                    full_date = time.strftime("%Y-%m-%d", time_struct)
                else:
                    full_date = "未知日期"

                results.append(
                    {"title": clean_title, "url": href, "date": full_date}
                )

            except Exception as e:
                self.logger.warning(f"some error occur when analyse notification.{e}")
                continue

        return results

    def fetch_detail(self, url):
        body_content = self.session.get(url).text
        tree = html.fromstring(body_content)

        container = tree.xpath('//div[@class="newsMain-content"]')
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

        content_html = content_html.replace("\\r\\n", "").replace("\\n", "")
        content_html = content_html.replace("\r", "").replace("\n", "")
        content_html = re.sub(r'>\s+<', '><', content_html)
        content_html = content_html.strip()

        # attachments
        attachments = tree.xpath('//div[@class="bbs-data"]/ul[contains(@class, "bbs-data-list")]/li')
        
        results = []
        for li in attachments:
            a_tag = li.xpath('.//a')[0]

            title = a_tag.xpath("./text()")
            clean_title = title[0].strip() if title else "未知标题"

            # url
            href = a_tag.xpath("./@href")[0]

            results.append({"title": clean_title, "url": href})

        return {"html": content_html, "attachments": results}
    
    def download_file(self, url, save_path, referer=None, max_size=None, is_ensure_auth=False):
        return super().download_file(url, save_path, referer, max_size, is_ensure_auth)
