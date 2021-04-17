# radiru-dl

らじるの聞き逃し番組録音スクリプト

- 聞き逃し番組表: https://www.nhk.or.jp/radio/ondemand/

## Usage

イメージプル

```
$ docker pull aoshimash/radiru-dl:latest
```

音声ファイルダウンロード

```
$ docker run -v $(pwd)/output:/root/output -it aoshimash/radiru-dl <プレイヤーURL> <出力ファイル名>
```

e.g.

```
$ docker run -v $(pwd)/output:/root/output -it aoshimash/radiru-dl "https://www.nhk.or.jp/radio/player/ondemand.html?p=0915_01_3208917" "まいにち中国語_第10課.aac"
```

プレイヤーURLは `https://www.nhk.or.jp/radio/player/ondemand.html` にクエリ文字列を追加したもの。
