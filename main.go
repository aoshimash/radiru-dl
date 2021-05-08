package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/sclevine/agouti"
)

func getDocument(url string) (*goquery.Document, error) {

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

	err = page.Navigate(url)
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

func getHlsUrlFromPlayer(url string) (string, string, error) {

	// Documentオブジェクトを取得
	doc, err := getDocument(url)
	if err != nil {
		return "", "", err
	}

	// hlsURLを検索
	target_elem_hlsurl := "html > body#playerwin > div#container_player.od > div#ODcontents > div.nol_audio_player"
	target_attr_hlsurl := "data-hlsurl"
	hlsURL, exists := doc.Find(target_elem_hlsurl).Attr(target_attr_hlsurl)
	if !exists {
		return "", "", fmt.Errorf("coulden't find hlsURL")
	}

	// Title検索
	target_elem_title := "html > body#playerwin > div#container_player.od > div#ODcontents > div#bangumi > div#title > h3"
	title := doc.Find(target_elem_title).Text()

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
	hlsURL, title, err_get_hlsurl := getHlsUrlFromPlayer(url)
	if err_get_hlsurl != nil {
		log.Fatalf("Failed to get HLS url: %v", err_get_hlsurl)
	}

	output := "output/" + title + ".aac"
	fmt.Printf("Downloading '%v' from '%v'\n)", title, hlsURL)

	// ffmpegでM3U8をダウンロード
	err_download_m3u8 := exec.Command("ffmpeg", "-i", hlsURL, "-write_xing", "0", output).Run()
	if err_download_m3u8 != nil {
		log.Panicf("%v\n", err_download_m3u8)
	}

}
