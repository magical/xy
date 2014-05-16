set -eu

prog="/tmp/riptext"
dump="$1"
dir="textrip"

rip() {
    mkdir -p "$dir/$2"
    "$prog" "$dump/$1" "$dir/$2"
}

go build -o "$prog" "`dirname "$0"`"/text.go
rip a/0/7/2 text/ja-kana
rip a/0/7/3 text/ja-kanji
rip a/0/7/4 text/en
rip a/0/7/5 text/fr
rip a/0/7/6 text/it
rip a/0/7/7 text/de
rip a/0/7/8 text/es
rip a/0/7/9 text/ko

rip a/0/8/0 script/ja-kana
rip a/0/8/1 script/ja-kanji
rip a/0/8/2 script/en
rip a/0/8/3 script/fr
rip a/0/8/4 script/it
rip a/0/8/5 script/de
rip a/0/8/6 script/es
rip a/0/8/7 script/ko
