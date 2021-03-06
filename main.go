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
	PROGRAM_URL = "https://www.nhk.or.jp/radio/ondemand/detail.html"
	// 番組ページにあるplayerのクエリパラメタがある要素
	SELECTOR_TO_FIND_QUERY_PARAMETER = "html > body#pagetop > div#container > div#main > div.inner > div.progblock > div.block"
	// PlayerページURL(クエリパラメタ抜き)
	PLAYER_URL = "https://www.nhk.or.jp/radio/player/ondemand.html"
	// PlayerページにあるHLS-URLの要素
	SELECTOR_TO_FIND_HLS_URL = "html > body#playerwin > div#container_player.od > div#ODcontents > div.nol_audio_player"
	// PlayerページにあるHLS-URLの属性
	ATTR_NAME_TO_FIND_HLS_URL = "data-hlsurl"
	// PlayerページにあるTitleの要素
	SELECTOR_TO_FIND_TITLE = "html > body#playerwin > div#container_player.od > div#ODcontents > div#bangumi > div#title > h3"
)

type RadiruPlayer struct {
	hlsURL *url.URL
	title  string
}

func getGoqueryDocument(url *url.URL) (doc *goquery.Document, err error) {

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
	err = driver.Start()
	if err != nil {
		return
	}
	defer driver.Stop()

	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		return
	}

	err = page.Navigate(url.String())
	if err != nil {
		return
	}

	content, err := page.HTML()
	if err != nil {
		return
	}

	reader := strings.NewReader(content)
	doc, err = goquery.NewDocumentFromReader(reader)

	return
}

func getPlayerParamsFromProgramPage(url *url.URL) (playerParams []string, err error) {
	// Documentオブジェクトを取得
	doc, err := getGoqueryDocument(url)
	if err != nil {
		return
	}

	// Playerのパラメタリストを取得
	doc.Find(SELECTOR_TO_FIND_QUERY_PARAMETER).Each(func(i int, s *goquery.Selection) {
		elem, err := s.Find("li > a").Attr("href")
		if !err {
			return
		}
		playerParam := strings.Split(elem, "'")[1]
		playerParams = append(playerParams, playerParam)
	})

	return
}

func getRadiruPlayer(url *url.URL) (radiruPlayer RadiruPlayer, err error) {
	// Documentオブジェクトを取得
	doc, err := getGoqueryDocument(url)
	if err != nil {
		return
	}

	// hlsURLを検索
	hlsURLStr, exists := doc.Find(SELECTOR_TO_FIND_HLS_URL).Attr(ATTR_NAME_TO_FIND_HLS_URL)
	if !exists {
		err = fmt.Errorf("coulden't find hlsURL")
		return
	}
	hlsURL, err := url.Parse(hlsURLStr)
	if err != nil {
		return
	}

	// Title検索
	title := doc.Find(SELECTOR_TO_FIND_TITLE).Text()

	radiruPlayer = RadiruPlayer{
		hlsURL: hlsURL,
		title:  title,
	}

	return
}

func getRadiruPlayers(targetURL *url.URL) (radiruPlayers []RadiruPlayer, err error) {
	// URLが"番組"と"プレイヤー"どちらかの場合で処理を分岐
	urlStrWithoutParam := strings.Split(targetURL.String(), "?")[0]
	var playerURLs []*url.URL
	if urlStrWithoutParam == PROGRAM_URL {
		var playerParams []string
		playerParams, err = getPlayerParamsFromProgramPage(targetURL)
		if err != nil {
			return
		}

		for _, playerParam := range playerParams {
			rawPlayerURL := PLAYER_URL + "?" + playerParam
			var playerURL *url.URL
			playerURL, err = url.Parse(rawPlayerURL)
			if err != nil {
				return
			}
			playerURLs = append(playerURLs, playerURL)
		}
	} else if urlStrWithoutParam == PLAYER_URL {
		playerURLs = append(playerURLs, targetURL)
	} else {
		err = fmt.Errorf("invalid URL")
		return
	}

	// RadiruPlayerを取得
	for _, playerURL := range playerURLs {
		var radiruPlayer RadiruPlayer
		radiruPlayer, err = getRadiruPlayer(playerURL)
		if err != nil {
			return
		}
		radiruPlayers = append(radiruPlayers, radiruPlayer)
	}
	return
}

func main() {
	// コマンドライン引数の処理
	flag.Parse()
	args := flag.Args()
	if len(args) != 1 {
		log.Fatalf("Unexpected arguments %v\n", args)
	}
	urlStr := args[0]

	// URL生成
	targetURL, parseErr := url.Parse(urlStr)
	if parseErr != nil {
		log.Fatalf("Failed to Parse URL %v\n", urlStr)
	}
	radiruPlayers, err := getRadiruPlayers(targetURL)
	if err != nil {
		log.Fatalf("Failed to get RadiruPlayers %v\n", err)
	}

	fmt.Println("Radiru Player")
	for _, radiruPlayer := range radiruPlayers {
		fmt.Printf("title: %s, hlsURL: %s\n", radiruPlayer.title, radiruPlayer.hlsURL)
	}

	// ffmpegでM3U8をダウンロード
	for _, radiruPlayer := range radiruPlayers {
		output := "output/" + radiruPlayer.title + ".aac"
		fmt.Printf("Downloading '%v'\n", output)
		err := exec.Command("ffmpeg", "-i", radiruPlayer.hlsURL.String(), "-write_xing", "0", output).Run()
		if err != nil {
			log.Panicf("%v\n", err)
		}
	}

}
