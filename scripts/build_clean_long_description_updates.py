#!/usr/bin/env python3
import argparse
import base64
import html
import re
import sys
from typing import Iterable


BLOCK_TAGS = ("h2", "h3", "p", "ul", "ol", "table", "blockquote")


def strip_images_and_promotions(value: str) -> str:
    value = re.sub(r"\[caption[^\]]*\].*?\[/caption\]", "", value, flags=re.I | re.S)
    value = re.sub(r"\[video[^\]]*\].*?\[/video\]", "", value, flags=re.I | re.S)
    value = re.sub(r"<img\b[^>]*>", "", value, flags=re.I)
    value = re.sub(r"<figure\b[^>]*>.*?</figure>", "", value, flags=re.I | re.S)
    value = re.sub(r"#image_title", "", value, flags=re.I)
    value = re.sub(
        r'<a\b[^>]*href="https://line\.me/[^"]*"[^>]*>.*?</a>',
        "",
        value,
        flags=re.I | re.S,
    )
    value = re.sub(
        r'<h[1-6][^>]*>\s*(?:<span[^>]*>)?\s*[（(][^<]*(?:限時|點擊此處購買|點擊購買|購買五盒|買五瓶送一瓶)[^<]*.*?</h[1-6]>',
        "",
        value,
        flags=re.I | re.S,
    )
    return value


def normalize_html(value: str) -> str:
    value = value.replace("\r\n", "\n").replace("\r", "\n")
    value = html.unescape(value).replace("&nbsp;", " ").replace("\u00a0", " ")
    value = strip_images_and_promotions(value)

    value = re.sub(r"\s(?:style|id|class|align|width|height|target|rel)=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\sdata-[a-z0-9_-]+=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"<span\b[^>]*>", "", value, flags=re.I)
    value = re.sub(r"</span>", "", value, flags=re.I)

    value = re.sub(r"<h1\b[^>]*>", "<h2>", value, flags=re.I)
    value = re.sub(r"</h1>", "</h2>", value, flags=re.I)
    value = re.sub(r"<h5\b[^>]*>", "<p><strong>", value, flags=re.I)
    value = re.sub(r"</h5>", "</strong></p>", value, flags=re.I)
    value = re.sub(r"<h6\b[^>]*>", "<p><strong>", value, flags=re.I)
    value = re.sub(r"</h6>", "</strong></p>", value, flags=re.I)
    value = re.sub(r"^(<h[23]>.*?</h[23]>)\s*<strong>(.*?)</strong>", r"\1<p><strong>\2</strong></p>", value, count=1, flags=re.I | re.S)

    value = re.sub(r"<br\s*/?>", "<br>", value, flags=re.I)
    value = re.sub(r"<p>\s*</p>", "", value, flags=re.I)
    value = re.sub(r"<p>\s*<br>\s*</p>", "", value, flags=re.I)
    value = re.sub(r"<p>\s*[  ]+\s*</p>", "", value, flags=re.I)
    value = re.sub(r"^\s*、\s*$", "", value, flags=re.M)
    value = re.sub(r"^\s*[、，]\s*\n", "", value, flags=re.M)
    value = re.sub(r"<div>\s*</div>", "", value, flags=re.I)
    value = re.sub(r"<h[1-6]>\s*</h[1-6]>", "", value, flags=re.I)

    value = re.sub(r"<li>\s*<p>", "<li>", value, flags=re.I)
    value = re.sub(r"</p>\s*</li>", "</li>", value, flags=re.I)
    value = re.sub(r"<li>\s*<br>", "<li>", value, flags=re.I)
    value = re.sub(r"<p>\s*<strong>\s*產品特點[:：]?\s*</strong>\s*</p>", "<h3>產品特點</h3>", value, flags=re.I)
    value = re.sub(r"<h[23]>\s*產品優勢\s*</h[23]>", "<h3>產品特點</h3>", value, flags=re.I)
    value = re.sub(r"<h[23]>\s*主機特點[:：]?\s*</h[23]>", "<h3>產品特點</h3>", value, flags=re.I)
    value = re.sub(r"<h[23]>\s*產品特點[:：]?\s*</h[23]>", "<h3>產品特點</h3>", value, flags=re.I)
    value = re.sub(r"<h[23]>\s*產品規格\s*</h[23]>", "<h3>產品規格</h3>", value, flags=re.I)
    value = re.sub(r"<h4>\s*<strong>\s*(.*?)\s*</strong>\s*</h4>", r"<h3>\1</h3>", value, flags=re.I)
    value = re.sub(r"<div>\s*(<table.*?</table>)\s*</div>", r"\1", value, flags=re.I | re.S)
    value = re.sub(r"<table>\s*<tbody>", "<table><tbody>", value, flags=re.I)
    value = re.sub(r"</tbody>\s*</table>", "</tbody></table>", value, flags=re.I)

    value = re.sub(
        r"(運輸時間：訂單下訂確定無誤之後24小時內寄出，2-3天送達指定門市（節假日除外）)",
        r"<p><strong>\1</strong></p>",
        value,
    )

    if re.match(r"^\s*<strong>.*?</strong>\s*", value, flags=re.I | re.S):
        value = re.sub(r"^\s*<strong>(.*?)</strong>", r"<p><strong>\1</strong></p>", value, count=1, flags=re.I | re.S)

    value = re.sub(r">\s+<", "><", value)
    for tag in BLOCK_TAGS:
        value = re.sub(fr"(</{tag}>)", r"\1\n", value, flags=re.I)
    value = re.sub(r"\n{3,}", "\n\n", value)
    value = re.sub(r"[ \t]{2,}", " ", value)
    return value.strip()


def iter_rows(lines: Iterable[str]):
    for raw_line in lines:
        line = raw_line.rstrip("\n")
        if not line.strip():
            continue
        parts = line.split("\t", 1)
        if len(parts) != 2:
            continue
        product_id, hex_text = parts
        if not hex_text:
            continue
        yield int(product_id), bytes.fromhex(hex_text).decode("utf-8", errors="ignore")


def build_update_sql(product_id: int, cleaned: str) -> str:
    encoded = base64.b64encode(cleaned.encode("utf-8")).decode("ascii")
    return (
        "UPDATE products "
        "SET specifications = JSON_SET(COALESCE(specifications, JSON_OBJECT()), "
        f"'$.longDescription', CONVERT(FROM_BASE64('{encoded}') USING utf8mb4)) "
        f"WHERE id = {product_id};"
    )


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--preview", action="store_true")
    parser.add_argument("--limit", type=int, default=3)
    args = parser.parse_args()

    rows = list(iter_rows(sys.stdin))
    if args.preview:
        for product_id, original in rows[: args.limit]:
            cleaned = normalize_html(original)
            print("=" * 80)
            print(f"ID: {product_id}")
            print("- CLEANED -")
            print(cleaned[:1800])
            print()
        return

    print("START TRANSACTION;")
    for product_id, original in rows:
        cleaned = normalize_html(original)
        print(build_update_sql(product_id, cleaned))
    print("COMMIT;")


if __name__ == "__main__":
    main()
