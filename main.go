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
	// 番組ページにあるplayerのクエリパラメタがある要素
	targetElemPlayerParam = "html > body#pagetop > div#container > div#main > div.inner > div.progblock > div.block"
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

func getPlayerParamsFromProgramPage(url *url.URL) ([]string, error) {
	// Documentオブジェクトを取得
	doc, err := getDocument(url)
	if err != nil {
		return nil, err
	}

	// Playerのパラメタリストを取得
	playerParams := []string{}
	doc.Find(targetElemPlayerParam).Each(func(i int, s *goquery.Selection) {
		elem, _ := s.Find("li > a").Attr("href")
		/*
			if !exists {
				return nil, nil
			}
		*/
		playerParam := strings.Split(elem, "'")[1]
		playerParams = append(playerParams, playerParam)
	})

	return playerParams, nil
}

func getHLSURLFromPlayerPage(url *url.URL) (string, string, error) {
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
	rawURL := args[0]

	// URL生成
	targetURL, parseErr := url.Parse(rawURL)
	if parseErr != nil {
		log.Fatalf("Failed to Parse URL %v\n", rawURL)
	}

	// URLが"番組"と"プレイヤー"どちらかの場合で処理を分岐
	rawURLWithoutParam := strings.Split(rawURL, "?")[0]
	var playerURLs []*url.URL
	if rawURLWithoutParam == rawProgramURLWithoutParam {
		playerParams, e := getPlayerParamsFromProgramPage(targetURL)
		if e != nil {
			log.Fatalf("Failed when analysing %v %v\n", targetURL.String(), e)
		}
		for _, playerParam := range playerParams {
			rawPlayerURL := rawPlayerURLWithoutParam + "?" + playerParam
			playerURL, playerURLParseErr := url.Parse(rawPlayerURL)
			if playerURLParseErr != nil {
				log.Fatalf("Failed to Parse Player URL %v\n", rawPlayerURL)
			}
			playerURLs = append(playerURLs, playerURL)
		}

	} else if rawURLWithoutParam == rawPlayerURLWithoutParam {
		playerURLs = []*url.URL{targetURL}
	} else {
		log.Fatalf("Unexpected URL")
	}

	// M3U8のURLを取得
	for _, playerURL := range playerURLs {
		hlsURL, title, errGetHLSURL := getHLSURLFromPlayerPage(playerURL)
		if errGetHLSURL != nil {
			log.Fatalf("Failed to get HLS url (%v)", errGetHLSURL)
		}
		output := "output/" + title + ".aac"
		fmt.Printf("Downloading '%v' from '%v'\n", title, hlsURL)

		// ffmpegでM3U8をダウンロード
		errDownloadM3U8 := exec.Command("ffmpeg", "-i", hlsURL, "-write_xing", "0", output).Run()
		if errDownloadM3U8 != nil {
			log.Panicf("%v\n", errDownloadM3U8)
		}
	}
}
