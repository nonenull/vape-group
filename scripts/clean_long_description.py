#!/usr/bin/env python3
import html
import re
import sys


def strip_images_and_promotions(value: str) -> str:
    value = re.sub(r"\[caption[^\]]*\].*?\[/caption\]", "", value, flags=re.I | re.S)
    value = re.sub(r"<img\b[^>]*>", "", value, flags=re.I)
    value = re.sub(r"<figure\b[^>]*>.*?</figure>", "", value, flags=re.I | re.S)
    value = re.sub(
        r'<a\b[^>]*href="https://line\.me/[^"]*"[^>]*>.*?</a>',
        "",
        value,
        flags=re.I | re.S,
    )
    value = re.sub(
        r'<h[1-6][^>]*>\s*(?:<span[^>]*>)?\s*[（(]點擊此處購買.*?</h[1-6]>',
        "",
        value,
        flags=re.I | re.S,
    )
    value = re.sub(
        r'<a\b[^>]*>\s*(?:<span[^>]*>)?\s*[（(]點擊此處購買.*?</a>',
        "",
        value,
        flags=re.I | re.S,
    )
    return value


def normalize_blocks(value: str) -> str:
    value = value.replace("\r\n", "\n").replace("\r", "\n")
    value = html.unescape(value).replace("&nbsp;", " ")
    value = strip_images_and_promotions(value)

    value = re.sub(r"\sstyle=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\sid=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\sclass=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\sdata-[a-z0-9_-]+=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\salign=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\swidth=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\sheight=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\starget=\"[^\"]*\"", "", value, flags=re.I)
    value = re.sub(r"\srel=\"[^\"]*\"", "", value, flags=re.I)

    value = re.sub(r"<h1\b[^>]*>", "<h2>", value, flags=re.I)
    value = re.sub(r"</h1>", "</h2>", value, flags=re.I)
    value = re.sub(r"<h5\b[^>]*>", "<p><strong>", value, flags=re.I)
    value = re.sub(r"</h5>", "</strong></p>", value, flags=re.I)

    value = re.sub(r"<br\s*/?>", "<br>", value, flags=re.I)
    value = re.sub(r"<p>\s*</p>", "", value, flags=re.I)
    value = re.sub(r"<p>\s*[\u00a0 ]+\s*</p>", "", value, flags=re.I)
    value = re.sub(r"<div>\s*</div>", "", value, flags=re.I)

    value = re.sub(r"<li>\s*<p>", "<li>", value, flags=re.I)
    value = re.sub(r"</p>\s*</li>", "</li>", value, flags=re.I)
    value = re.sub(r"<li>\s*<br>", "<li>", value, flags=re.I)

    value = re.sub(r"<h[23]>\s*產品優勢\s*</h[23]>", "<h3>產品特點</h3>", value, flags=re.I)
    value = re.sub(r"<h[23]>\s*產品特點[:：]?\s*</h[23]>", "<h3>產品特點</h3>", value, flags=re.I)
    value = re.sub(r"<h[23]>\s*產品規格\s*</h[23]>", "<h3>產品規格</h3>", value, flags=re.I)

    value = re.sub(
        r"(運輸時間：訂單下訂確定無誤之後24小時內寄出，2-3天送達指定門市（節假日除外）)",
        r"<p><strong>\1</strong></p>",
        value,
    )

    value = re.sub(r">\s+<", "><", value)
    value = re.sub(r"\n{3,}", "\n\n", value)
    value = re.sub(r"[ \t]{2,}", " ", value)
    return value.strip()


def clean_long_description(value: str) -> str:
    return normalize_blocks(value)


def main() -> None:
    raw = sys.stdin.read()
    sys.stdout.write(clean_long_description(raw))


if __name__ == "__main__":
    main()
