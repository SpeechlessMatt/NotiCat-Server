import json
import sys
import argparse
from clients.BUPTClient import BUPTClient
from clients.BilibiliClient import BilibiliClient
from clients.CMathcClient import CMathcClient
import logging

logging.basicConfig(
    level=logging.DEBUG, # 设置级别：DEBUG, INFO, WARNING, ERROR, CRITICAL
    format='%(asctime)s [%(levelname)s] %(message)s', # 时间 [级别] 消息
    datefmt='%Y-%m-%d %H:%M:%S',
    stream=sys.stderr,
)

def main():
    parser = argparse.ArgumentParser(description="NotiCat Python Catcher CLI")

    # basic arguments: client username password
    parser.add_argument("client", help="Client type (e.g., bupt, bilibili)")
    parser.add_argument("username", help="Student ID")
    parser.add_argument("password", help="Password")

    # mode: --action (default: list)
    parser.add_argument("--action", choices=["list", "detail", "download"], default="list", help="Catch Mode")
    
    # download
    parser.add_argument("--url", help="URL for detail or download")
    parser.add_argument("--save-path", help="Path to save downloaded file")
    parser.add_argument("--max-size", type=int, help="Max size of files to download (MB)")
    parser.add_argument("--referer", help="Referer URL for downloading")
    parser.add_argument("--extra", help="Extra JSON to provide additional API fields without modifying the script")

    args = parser.parse_args()

    # defined clients
    clients = {
        "bupt": BUPTClient,
        "bili": BilibiliClient,
        "cmathc": CMathcClient
    }

    # required args
    if args.client not in clients:
        logging.error(f"Unsupported client: {args.client}")
        sys.exit(1)

    # extra
    logging.debug(f"extra: {args.extra}")
    extra = {}
    if args.extra:
        try:
            extra_data = json.loads(args.extra)
            extra.update(extra_data)
        except json.JSONDecodeError:
            logging.error("error: --extra is not a valid JSON string")
            sys.exit(1)

    # client
    client = clients[args.client](args.username, args.password, extra)
    
    result = None
    try:
        if args.action == "list":
            result = client.fetch()
        
        elif args.action == "detail":
            if not args.url:
                logging.error("Detail action requires --url")
                sys.exit(1)
            result = client.fetch_detail(args.url)
            
        elif args.action == "download":
            if not args.url or not args.save_path:
                logging.error("Download action requires --url and --save-path")
                sys.exit(1)

            download_kwargs = {
                "referer": args.referer,
                "max_size": args.max_size
            }

            success = client.download_file(args.url, args.save_path, **download_kwargs)
            result = {"success": success, "path": args.save_path}

        # 4. 统一输出 JSON 到 stdout
        if result is not None:
            sys.stdout.write(json.dumps(result, ensure_ascii=False))
            sys.stdout.flush()

    except Exception as e:
        logging.exception(f"Action {args.action} failed: {e}")
        sys.exit(1)

if __name__ == "__main__":
    main()
