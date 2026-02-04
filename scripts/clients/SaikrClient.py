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

# Author: Czy_4201b <speechlessmatt@qq.com>
# Created: 2026-02-04

import re
import time
from lxml import etree, html
from lxml_html_clean import Cleaner
from .base import BaseClient

class SaikrClient(BaseClient):
    client_id = "saikr"

    def __init__(self, username, password, extra) -> None:
        super().__init__(username=username, password=password, extra=extra)

    def fetch(self):
        headers = {"Origin": "https://www.saikr.com", "Referer": "https://www.saikr.com/"}
        resp = self.session.get("https://apiv4buffer.saikr.com/api/pc/contest/lists?page=1&limit=10&univs_id=&class_id=&level=0&sort=0", headers=headers)

        data = resp.json()
        assert data["code"] == 200, "server error!"

        result = []
        notice_list = data["data"]["list"]
        for contest in notice_list:
            # title
            title = contest["contest_name"]
            clean_title = re.sub(r'[\x00-\x09\x0b-\x0c\x0e-\x1f\x7f-\x9f\u200b-\u200f\u202a-\u202e]', '', title)

            # date
            date = contest["regist_start_time"]
            timeArray = time.localtime(date)
            full_date = time.strftime("%Y-%m-%d", timeArray)

            # url
            origin_url = contest["contest_url"]
            full_url = f"https://www.saikr.com/{origin_url}"

            result.append({"title": clean_title, "url": full_url, "date": full_date})

        return result

    def fetch_detail(self, url):
        # vse? : https://new.saikr.com/vse/KYBreading25
        # rule: if there is no event4-1-detail-text-box text-body clearfix (class)
        # or event4-1-detail-box event4-1-doc-box (class)
        # https://apiv4buffer.saikr.com/api/pc/contest/info?contest_url=KYBreading25&isp=
        # https://apiv4buffer.saikr.com/api/pc/contest/info?contest_url=Robot%2Farborseek&isp=

        resp = self.session.get(url)
        assert resp.status_code == 200, "server error !"

        resp_html = html.fromstring(resp.text)

        # mode 1: if there is event4-1-detail-text-box -> attachments and body is here
        # body content
        content = resp_html.xpath('//*[@id="eventDetailBox"]/*/div[@class="event4-1-detail-text-box text-body clearfix"]')
        if len(content) != 0:
            container = content[0]

            # the same with BUPTClient
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

            # no need to replace base url because saikr has full url ~
            # self.logger.debug(content_html)

            # attachments
            attachments_content = resp_html.xpath('//*[@id="eventDetailBox"]/div[@class="event4-1-detail-box event4-1-doc-box"]')

            attachments = []
            for div in attachments_content:
                try:
                    a_tags = div.xpath(".//a")
                    if len(a_tags) == 0:
                        continue

                    a_tag = a_tags[0]

                    # title 
                    title = div.xpath(".//div/@title")[0]
                    clean_title = title.strip() if title else "未知标题"

                    # url
                    url = a_tag.xpath("./@href")[0]

                    attachments.append({"titile": clean_title, "url": url})
                except Exception as e:
                    self.logger.warning(f"some error occur when analyse notification.{e}")
                    continue
            return {"html": content_html, "attachments": attachments}

        else:
            jump_to_url = resp.url
            self.logger.debug(f"jump_to_url:{jump_to_url}")

            if not jump_to_url.startswith("https://new.saikr.com/vse/"):
                self.logger.warning("url解析错误")
                return

            query_name = jump_to_url.replace("https://new.saikr.com/vse/", "")
            encoded_name = query_name.replace("/", "%2F")
            query_url = f"https://apiv4buffer.saikr.com/api/pc/contest/info?contest_url={encoded_name}&isp="

            self.logger.debug(f"query_url: {query_url}")

            headers = {
                "Origin": "https://new.saikr.com",
                "Referer": "https://new.saikr.com/"
            }

            query = self.session.get(query_url, headers=headers)
            json = query.json()
            data = json["data"]

            self.logger.debug(data)

            content = data["content"]
            attachments = data["attachment"]

            # content
            container = html.fromstring(content)
            # the same with BUPTClient
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

            # attachments
            results = []
            for attachment in attachments.keys():
                title = attachment
                url = attachments[attachment]
                results.append({"title": title, "url": url})

            return {"html": content_html, "attachments": results}
