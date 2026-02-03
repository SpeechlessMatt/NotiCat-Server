from .base import BaseClient
import random
import json
import base64
import time
import re
from urllib.parse import quote
import hashlib

def encode_getActiveFeaturesStr(winWidth=298, winHeight=691, scrollTop=0, scrollLeft=0):
    ds = []
    wh_random = int(random.random() * 114)
    wh = [
        2 * winWidth + 2 * winHeight + 3 * wh_random,
        4 * winWidth - winHeight + wh_random,
        wh_random,
    ]
    of_random = int(random.random() * 514)
    of = [
        3 * scrollTop + 2 * scrollLeft + of_random,
        4 * scrollTop - 4 * scrollLeft + 2 * of_random,
        of_random,
    ]
    activateFeaturesStr = {"ds": ds, "wh": wh, "of": of}
    return json.dumps(activateFeaturesStr, separators=(",", ":"), ensure_ascii=False)


def special_base64(input_string):
    utf8_bytes = input_string.encode("utf-8")
    b64_str = base64.b64encode(utf8_bytes).decode("utf-8")
    return b64_str[:-2]


def webglStr(webgl="WebGL 1.0 (OpenGL ES 2.0 Chromium)"):
    return special_base64(webgl)

def webglVendorAndRenderer(
    webglvar="ANGLE (AMD, AMD Radeon(TM) Graphics (0x0000164C) Direct3D11 vs_5_0 ps_5_0, D3D11)Google Inc. (AMD)",
):
    return special_base64(webglvar)

def gR(e):
    return "".join([chr(ord(c) - 1) for c in e])

def getWbi(url: str):
    return url.split("/")[-1].split(".")[0]

def encodeCombinedWbi(input_string):
    t = [46, 47, 18, 2, 53, 8, 23, 32, 15, 50, 10, 31, 58, 3, 45, 35, 27, 43, 5, 49, 33, 9, 42, 19, 29, 28, 14, 39, 12, 38, 41, 13, 37, 48, 7, 16, 24, 55, 40, 61, 26, 17, 0, 1, 60, 51, 30, 4, 22, 25, 54, 21, 56, 59, 6, 63, 57, 62, 11, 36, 20, 34, 44, 52]

    result = []
    for n in t:
        if n < len(input_string): 
            result.append(input_string[n])

    return "".join(result)[:32]

def wRidGeneratePayload(payload: dict, salt: str):
    wts = int(time.time())

    # add wts
    payload["wts"] = wts

    d = []
    sorted_key = sorted(payload.keys())
    for key in sorted_key:
        value = payload[key]

        if isinstance(value, str):
            value = re.sub(r"[!'()*]", "", value)

        clean_key = quote(str(key), safe='')
        clean_value = quote(str(value), safe='')

        d.append(f"{clean_key}={clean_value}")

    d_str = "&".join(d)

    final_str = d_str + salt

    m = hashlib.md5()
    m.update(final_str.encode('utf-8'))
    w_rid = m.hexdigest()

    payload.pop("wts", None)
    payload["w_rid"] = w_rid
    payload["wts"] = wts

def jsonToResult(data: dict):
    items = data["data"]["items"]
    result = []

    for item in items:
        module_author = item["modules"]["module_author"]
        module_dynamic = item["modules"]["module_dynamic"]
        major = module_dynamic.get("major", None)
        url = "//www.bilibili.com"

        if not major:
            text = module_dynamic["desc"]["text"]
        else:
            opus = major.get("opus", None)
            if not opus:
                text = major["archive"]["title"]
            else:
                text = opus["summary"]["text"]
                url = opus["jump_url"]

        clean_text = re.sub(r'[\x00-\x09\x0b-\x0c\x0e-\x1f\x7f-\x9f\u200b-\u200f\u202a-\u202e]', '', text)
        clean_url = f"https:{url}"
        pub_time = module_author["pub_time"]

        result.append({"title": clean_text, "url": clean_url, "date": pub_time})

    return result

class BilibiliClient(BaseClient):
    client_id = "bili"

    def __init__(self, username, password, extra) -> None:
        super().__init__(username=username, password=password, extra=extra)

    def login(self):
        return

    def isLogin(self) -> bool:
        return False

    def fetch(self):
        url = self.extra.get('url')
        self.logger.debug(self.extra)
        if not url:
            self.logger.error("require 'url' key in extra")
            return

        headers = {
            "Referer": url.rstrip('/'),
            "Origin": "https://space.bilibili.com",
        }

        # space and uid
        pattern = r"^https://space\.bilibili\.com/(\d+)/dynamic/?$"
        match = re.match(pattern, url)
        if match:
            uid = int(match.group(1))
        else:
            self.logger.warning(f"error url: cannot support {url}")
            return

        self.logger.debug("get into main page...")
        self.session.get("https://www.bilibili.com/", headers=headers)
        self.logger.debug("sleep 2 seconds...")
        time.sleep(2)

        self.logger.debug(f"enter {url}")
        self.session.get(url, headers=headers)
        self.logger.debug("sleep 2 seconds...")
        time.sleep(2)

        # spmid
        spmid = "333.1387"
        self.logger.debug(f"spmid: {spmid}")

        req_json = json.dumps({"platform": "web", "device": "pc", "spmid": spmid}, separators=(",", ":"), ensure_ascii=False)

        # wbi
        wbi_api = self.session.get("https://api.bilibili.com/x/web-interface/nav", headers=headers)
        wbi_dict = json.loads(wbi_api.text)
        img_url = wbi_dict['data']['wbi_img']['img_url']
        sub_url = wbi_dict['data']['wbi_img']['sub_url']
        img_key = getWbi(img_url)
        sub_key = getWbi(sub_url)
        self.logger.debug("img_key:" + img_key)
        self.logger.debug("sub_key:" + sub_key)

        # salt
        salt = encodeCombinedWbi(img_key + sub_key)
        self.logger.debug(f"salt: {salt}")

        payload = {
            "offset": "",
            "host_mid": uid,
            "timezone_offset": time.timezone // 60,
            "platform": "web",
            "features": "itemOpusStyle,listOnlyfans,opusBigCover,onlyfansVote,forwardListHidden,decorationCard,commentsNewVersion,onlyfansAssetsV2,ugcDelete,onlyfansQaCard,avatarAutoTheme,sunflowerStyle,cardsEnhance,eva3CardOpus,eva3CardVideo,eva3CardComment,eva3CardUser",
            "web_location": spmid,
            "dm_img_list": "[]",
            "dm_img_str": webglStr(),
            "dm_cover_img_str": webglVendorAndRenderer(),
            "dm_img_inter": encode_getActiveFeaturesStr(),
            "x-bili-device-req-json": req_json,
        }
        
        # w_rid
        wRidGeneratePayload(payload, salt)

        d = []
        for key in payload.keys():
            value = payload[key]

            if isinstance(value, str):
                value = re.sub(r"[!'()*]", "", value)

            clean_key = quote(str(key), safe='')
            clean_value = quote(str(value), safe=':[],')

            d.append(f"{clean_key}={clean_value}")

        d_str = "&".join(d)
        self.logger.debug(f"payload: {d_str}")

        data = self.session.get(f"https://api.bilibili.com/x/polymer/web-dynamic/v1/feed/space?{d_str}", headers=headers).json()
        return jsonToResult(data)
