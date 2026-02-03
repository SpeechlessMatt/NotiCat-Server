import re
import time
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
        resp = self.session.get(f"https://apiv4buffer.saikr.com/api/pc/contest/info?contest_url={url}&isp=")
        json = resp.json()
        data = json["data"]
        attachments = data["attachment"]

        self.logger.debug(attachments)
        return

