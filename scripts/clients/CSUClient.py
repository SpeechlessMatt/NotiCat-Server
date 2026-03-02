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

# from lxml_html_clean import Cleaner
import re
import random
import string
from Crypto.Cipher import AES
from Crypto.Util.Padding import pad
import base64
from lxml import etree
from urllib.parse import quote
from .base import BaseClient

def get_random_string(length):
    # simulate randomString
    chars = string.ascii_letters + string.digits
    return ''.join(random.choice(chars) for _ in range(length))

def encrypt_aes(password: str, salt: str):
    # raw_text
    n = get_random_string(64) + password
    # key
    f = salt.strip().encode('utf-8')
    # iv
    c = get_random_string(16).encode('utf-8')

    # CBC
    cipher = AES.new(f, AES.MODE_CBC, c)

    # padding and encrypt
    padded_data = pad(n.encode('utf-8'), AES.block_size, style='pkcs7')
    encrypted_bytes = cipher.encrypt(padded_data)

    return base64.b64encode(encrypted_bytes).decode('utf-8')

# gemini generate
def generate_html_body(title, date, url, dept="本科生院", category="校内通知"):
    """
    不再依赖 dict，直接传参，支持默认值
    """
    primary_color = "#005BAC" 
    bg_color = "#f4f7f9"

    html_template = f"""
    <div style="background-color: {bg_color}; padding: 30px 15px; font-family: 'Helvetica Neue', Helvetica, Arial, sans-serif;">
        <div style="max-width: 600px; margin: 0 auto; background-color: #ffffff; border-radius: 12px; overflow: hidden; box-shadow: 0 10px 30px rgba(0,0,0,0.08);">
            <div style="height: 6px; background-color: {primary_color};"></div>
            
            <div style="padding: 30px;">
                <h2 style="color: #2c3e50; font-size: 22px; line-height: 1.4; margin-bottom: 25px; border-left: 5px solid {primary_color}; padding-left: 15px; font-weight: 700;">
                    {title}
                </h2>

                <div style="background-color: #f8fbff; border: 1px solid #eef4fb; border-radius: 10px; padding: 20px; margin-bottom: 30px;">
                    <table style="width: 100%; border-collapse: collapse; font-size: 15px; color: #5d6d7e;">
                        <tr>
                            <td style="padding: 8px 0; width: 90px; font-weight: 600; color: #34495e;">发布部门：</td>
                            <td style="padding: 8px 0; color: #2c3e50;">{dept}</td>
                        </tr>
                        <tr>
                            <td style="padding: 8px 0; font-weight: 600; color: #34495e;">发布日期：</td>
                            <td style="padding: 8px 0; color: #2c3e50;">{date}</td>
                        </tr>
                        <tr>
                            <td style="padding: 8px 0; font-weight: 600; color: #34495e;">通知分类：</td>
                            <td style="padding: 8px 0;">
                                <span style="background-color: #e1f0ff; color: {primary_color}; padding: 4px 12px; border-radius: 20px; font-size: 13px; font-weight: 600;">
                                    {category}
                                </span>
                            </td>
                        </tr>
                    </table>
                </div>

                <div style="padding: 20px; border-top: 1px solid #f0f0f0;">
                    <p style="color: #7f8c8d; font-size: 14px; margin-bottom: 15px;">
                        <strong style="color: #34495e;">💡 温馨提示：</strong><br>
                        详细内容已封装在邮件附件的 <b style="color: {primary_color};">PDF</b> 文件中，请查阅。
                    </p>
                </div>

                <div style="text-align: center; margin-top: 10px;">
                    <a href="{url}" style="background-color: {primary_color}; color: #ffffff; padding: 14px 35px; text-decoration: none; border-radius: 8px; font-size: 16px; font-weight: bold; display: inline-block; transition: all 0.3s;">
                        点击在浏览器中预览原文
                    </a>
                </div>
            </div>

            <div style="background-color: #fafafa; padding: 20px; text-align: center; font-size: 12px; color: #bdc3c7; border-top: 1px solid #f0f0f0;">
                <p style="margin: 0;">此邮件由 NotiCat-CSUClient 爬取</p>
                <p style="margin: 5px 0 0;">© NotiCat-Server</p>
            </div>
        </div>
    </div>
    """
    return html_template

class CSUClient(BaseClient):
    client_id = "csu"
    
    def login(self):
        headers = {
            "Host": "oa.csu.edu.cn"
        }
        resp = self.session.get("https://oa.csu.edu.cn/con/ggtz", headers=headers)
        assert resp.status_code == 200, "server error!"

        self.logger.debug(f'jump to :{resp.url}')

        from urllib.parse import urlparse, parse_qs
        query = urlparse(resp.url).query
        params = parse_qs(query)

        service_value = params.get('service')[0]
        login_url = f"https://ca.csu.edu.cn/authserver/login?service={service_value}"

        html_body = etree.HTML(resp.text)

        # salt
        pwdEncryptSalt = html_body.xpath('//*[@id="pwdEncryptSalt"]/@value')
        if len(pwdEncryptSalt) == 0:
            return

        # execution
        execution_input = html_body.xpath('//*[@id="qrLoginForm"]/input[@name="execution"]/@value')
        if len(execution_input) == 0:
            return

        self.logger.debug(f"salt: {pwdEncryptSalt[0]}")
        salt = pwdEncryptSalt[0]
        
        username = self.username
        password = encrypt_aes(self.password, salt)
        captcha = ""
        execution = execution_input[0]

        payload = {
            "username": username,
            "password": password,
            "captcha": captcha,
            "_eventId": "submit",
            "cllt": "userNameLogin",
            "dllt": "generalLogin",
            "lt": "",
            "execution": execution
        }

        self.logger.info("post login request now!")
        login_resp = self.session.post(login_url, data=payload, allow_redirects=True)
        assert login_resp.status_code == 200, "login error !"

        self.logger.debug(f"当前所有 Cookies: {self.session.cookies.get_dict()}")

        # save cookies
        self._save_cookies()

    def isLogin(self) -> bool:
        try:
            # travel to ca first~
            self.session.get("https://ca.csu.edu.cn/personalInfo/personCenter/index.html")

            headers = {
                "Host": "oa.csu.edu.cn"
            }
            resp = self.session.get("https://oa.csu.edu.cn/con/ggtz", headers=headers)

            # self.logger.debug(resp.text)
            return resp.url == "https://oa.csu.edu.cn/con/ggtz"
        except Exception:
            return False

    def fetch(self):
        self._ensure_auth()
        
        headers = {
            "Host": "oa.csu.edu.cn",
            "Origin": "https://oa.csu.edu.cn"
        }

        # 本科生院
        payload = {
            "params": '{"tableName":"ZNDX_ZHBG_GGTZ","tjnr":"IGFuZCBRQ0JNTUMgbGlrZSAnJeacrOenkeeUn+mZoiUnIA==","pxzd":""}',
            "pageSize": 1,
            "pageNo": 30
        }

        resp = self.session.post(
            "https://oa.csu.edu.cn/con/xnbg/contentList",
            data=payload,
            headers=headers
        ).json()

        data = resp["data"]
        results = []
        for notice in data:
            date = notice["DJSJ"]

            # title
            title = notice["WJBT"]
            clean_title = title.strip() if title else "未知标题"

            url = f"https://oa.csu.edu.cn/con/PDFPage?YWMC={notice["YWMC"]}&JLNM={notice["JLNM"]}"

            results.append(
                {"title": clean_title, "url": url, "date": date}
            )

        return results

    def fetch_detail(self, url: str):
        self._ensure_auth()

        headers = {
            "Host": "oa.csu.edu.cn",
            "Referer": "https://oa.csu.edu.cn/con/ggtz"
        }

        resp = self.session.get(url, headers=headers)

        # self.logger.debug(resp.text)

        tree = etree.HTML(resp.text)
        title = ''.join(tree.xpath('//*[@id="printDiv"]/div/div/div[1]/text()')).replace("\xa0", "").replace("\n", "").replace(">", "").strip()
        # self.logger.debug(title)
        date = tree.xpath('//*[@id="printDiv"]/div/div/div[1]/div/text()')[0].replace("发布日期：", "").strip()
        
        pattern = r"showAccessoryList\('([A-Z0-9]+)'\)"
        match = re.search(pattern, resp.text)
        if match:
            key_jlnm = str(match.group(1))
        else:
            self.logger.warning("error jlnm")
            return

        pattern = r'WJFL:"(\d+)"'
        match = re.search(pattern, resp.text)
        if match:
            key_wjfl = int(match.group(1))
        else:
            self.logger.warning("error wjfl")
            return

        pdf = tree.xpath('//*[@id="download"]/@href')[0]

        # attachments -- first pdf
        results = []
        pdf_title = title + ".pdf"
        pdf_url = f"https://oa.csu.edu.cn{pdf}"
        results.append({"title": pdf_title, "url": pdf_url})

        # attachment list
        headers = {
            "Host": "oa.csu.edu.cn",
            "Origin": "https://oa.csu.edu.cn",
            "Referer": url
        }

        payload = {
            "JLNM": key_jlnm,
            "WJFL": key_wjfl
        }

        # self.logger.debug(payload)

        attachments_resp = self.session.post("https://oa.csu.edu.cn/con/xnbg/getFjList", headers=headers, data=payload).json()
        for attachment in attachments_resp:
            filename = attachment["FJMC"]
            file_url = f"https://oa.csu.edu.cn/con/xnbg/downLoadFj?NBBM={attachment["NBBM"]}&fullfilename={quote(filename)}"
            results.append({"title": attachment["FJMC"], "url": file_url})

        # self.logger.debug(attachments_resp.status_code)
        # self.logger.debug(attachments_resp.text)

        html_body = generate_html_body(title=title, date=date, url=url)
        self.logger.debug(html_body)

        return {"html": html_body, "attachments": results}

