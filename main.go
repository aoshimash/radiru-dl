package main

import (
	"flag"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
)

const (
	// 番組ページURL(クエリパラメタ抜き)
	rawProgramURLWithoutParam = "https://www.nhk.or.jp/radio/ondemand/detail.html"
	// PlayerページURL(クエリパラメタ抜き)
	rawPlayerURLWithoutParam = "https://www.nhk.or.jp/radio/player/ondemand.html"
	// PlayerページにあるHLS-URLの要素
	targetElemHLSURL = "html > body#playerwin > div#container_player.od > div#ODcontents > div.nol_audio_player"
	// PlayerページにあるHLS-URLの属性
	targetAttrHLSURL = "data-hlsurl"
	// PlayerページにあるTitleの要素
	targetElemTitle = "html > body#playerwin > div#container_player.od > div#ODcontents > div#bangumi > div#title > h3"
)

func getDocument(url *url.URL) (*goquery.Document, error) {

	driver := agouti.ChromeDriver(
		agouti.ChromeOptions("args", []string{
			"--headless",
			"--window-size=300,1200",
			"--blink-settings=imagesEnabled=false",
			"--disable-gpu",
			"no-sandbox",
			"disable-dev-shm-usage",
		}),
	)
	err := driver.Start()
	if err != nil {
		return nil, err
	}
	defer driver.Stop()

	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		return nil, err
	}

	err = page.Navigate(url.String())
	if err != nil {
		return nil, err
	}

	content, err := page.HTML()
	if err != nil {
		return nil, err
	}

	reader := strings.NewReader(content)

	return goquery.NewDocumentFromReader(reader)
}

func getHlsURLFromPlayer(urlParam string) (string, string, error) {

	url, parseErr := url.Parse(rawPlayerURLWithoutParam + "?" + urlParam)
	if parseErr != nil {
		return "", "", parseErr
	}

	// Documentオブジェクトを取得
	doc, err := getDocument(url)
	if err != nil {
		return "", "", err
	}

	// hlsURLを検索
	hlsURL, exists := doc.Find(targetElemHLSURL).Attr(targetAttrHLSURL)
	if !exists {
		return "", "", fmt.Errorf("coulden't find hlsURL")
	}

	// Title検索
	title := doc.Find(targetElemTitle).Text()

	return hlsURL, title, nil
}

func main() {
	// コマンドライン引数の処理
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("Unexpected arguments %v\n", args)
	}
	url := args[0]

	// M3U8のURLを取得
	hlsURL, title, errGetHLSURL := getHlsURLFromPlayer(url)
	if errGetHLSURL != nil {
		log.Fatalf("Failed to get HLS url: %v", errGetHLSURL)
	}

	output := "output/" + title + ".aac"
	fmt.Printf("Downloading '%v' from '%v'\n)", title, hlsURL)

	// ffmpegでM3U8をダウンロード
	errDownloadM3U8 := exec.Command("ffmpeg", "-i", hlsURL, "-write_xing", "0", output).Run()
	if errDownloadM3U8 != nil {
		log.Panicf("%v\n", errDownloadM3U8)
	}
}
