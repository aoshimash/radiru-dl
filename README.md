# radiru-dl

らじるの聞き逃し番組録音スクリプト

- 聞き逃し番組表: https://www.nhk.or.jp/radio/ondemand/

## Usage

イメージプル

```bash
$ docker pull aoshimash/radiru-dl:latest
```


番組ページのURLまたはプレイヤーページのURLを指定して音声ファイルをダウンロードすることができます。

番組ページとは`https://www.nhk.or.jp/radio/ondemand/detail.html?p=****_**` のことです。番組ページを指定した場合はページ内の全放送が録音されます。
プレイヤーページとは、``https://www.nhk.or.jp/radio/player/ondemand.html?p=****_**_*****` のことです。プレイヤーページを指定した場合はそのプレイヤーで放送される１番組のみが録音されます。
番組ごとプレイヤーごとに最後のクエリストリングが異なります。


```bash
$ docker run -v $(pwd)/output:/root/output -it aoshimash/radiru-dl <番組ページURL/プレイヤーURL>
```

e.g.

```bash
$ docker run -v $(pwd)/output:/root/output -it aoshimash/radiru-dl "https://www.nhk.or.jp/radio/ondemand/detail.html?p=0045_01"
```

```bash
$ docker run -v $(pwd)/output:/root/output -it aoshimash/radiru-dl "https://www.nhk.or.jp/radio/player/ondemand.html?p=0045_01_44612"
```
