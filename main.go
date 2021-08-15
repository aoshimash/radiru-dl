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

func getPlayerParamsFromProgramPage(url *url.URL) ([]string, error) {
	// Documentオブジェクトを取得
	doc, err := getGoqueryDocument(url)
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

func getRadiruPlayer(url *url.URL) (playerInfo RadiruPlayer, err error) {
	// Documentオブジェクトを取得
	doc, err := getGoqueryDocument(url)
	if err != nil {
		return
	}

	// hlsURLを検索
	hlsURLStr, exists := doc.Find(targetElemHLSURL).Attr(targetAttrHLSURL)
	if !exists {
		err = fmt.Errorf("coulden't find hlsURL")
		return
	}
	hlsURL, err := url.Parse(hlsURLStr)
	if err != nil {
		return
	}

	// Title検索
	title := doc.Find(targetElemTitle).Text()

	playerInfo = RadiruPlayer{
		hlsURL: hlsURL,
		title:  title,
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

	// RadiruPlayerを取得
	var radiruPlayers []RadiruPlayer
	for _, playerURL := range playerURLs {
		radiruPlayer, err := getRadiruPlayer(playerURL)
		if err != nil {
			log.Fatalf("Failed to get HLS url (%v)", err)
		}
		radiruPlayers = append(radiruPlayers, radiruPlayer)
		//output := "output/" + playerInfo.title + ".aac"
		//fmt.Printf("Downloading '%v' from '%v'\n", playerInfo.title, playerInfo.hlsURL)
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
