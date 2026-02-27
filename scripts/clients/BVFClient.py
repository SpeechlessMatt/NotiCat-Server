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

class BVFClient(BaseClient):
    def __init__(self, username, password, extra) -> None:
        super().__init__(username=username, password=password, extra=extra)

    def fetch(self):
        self.logger.debug("Beijing Volunteer Federation client start...")

        extra_url = self.extra.get('url')
        self.logger.debug(self.extra)
        if not extra_url:
            self.logger.error("use default url: https://www.bv2008.cn/app/opp/list.php")
            extra_url = "https://www.bv2008.cn/app/opp/list.php"

        resp = self.session.get(extra_url)
        assert resp.status_code == 200, "server error!"

        html_body = etree.HTML(resp.text)
        
        # clearfix list
        clearfix_list = html_body.xpath('//div[@class="m10"]/ul[contains(@class, "list1") and contains(@class, "clearfix")]/li[@class="clearfix"]')

        results = []
        for li in clearfix_list:
            child_div = li.xpath('.//div[1]/div')[0]
            a = li.xpath('.//div[1]/a')[0]
            
            # title
            title = a.xpath('./@title')
            if not title:
                self.logger.debug("The notification has no title")
                clean_title = "未知标题"
            else:
                clean_title = title[0].strip()

            # url
            href = a.xpath('./@href')[0]
            url = f"https://www.bv2008.cn{href}"

            # date
            date = child_div.xpath('./text()')
            clean_date = date[0].strip() if date else ""

            results.append(
                {"title": clean_title, "url": url, "date": clean_date}
            )

        return results

    def fetch_detail(self, url: str):
        resp = self.session.get(url)
        assert resp.status_code == 200, "server error!"

        body_text = resp.text
        tree = html.fromstring(body_text)
        
        # body text
        container = tree.xpath('//div[@id="main_body"]')
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
        base_url = "https://www.bv2008.cn"
        cleaned_node.make_links_absolute(base_url)

        content_html = etree.tostring(cleaned_node, encoding="unicode", method="html")
        content_html = re.sub(r'>\s+<', '><', content_html)

        return {"html": content_html, "attachments": []}

