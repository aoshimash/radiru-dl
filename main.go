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

func getEmbededHLSURL(url string) string {
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
		log.Printf("Failed to start driver: %v", err)
	}
	defer driver.Stop()

	page, err := driver.NewPage(agouti.Browser("chrome"))
	if err != nil {
		log.Printf("Failed to open page: %v", err)
	}

	err = page.Navigate(url)
	if err != nil {
		log.Printf("Failed to navigate: %v", err)
	}

	content, err := page.HTML()
	if err != nil {
		log.Printf("Failed to get html: %v", err)
	}

	reader := strings.NewReader(content)

	doc, _ := goquery.NewDocumentFromReader(reader)

	var hlsURL string
	var exists bool
	target_elements := "html > body#playerwin > div#container_player.od > div#ODcontents > div.nol_audio_player"
	doc.Find(target_elements).Each(func(i int, s *goquery.Selection) {
		hlsURL, exists = s.Attr("data-hlsurl")
		if !exists {
			log.Printf("Coulden't find Attribute 'data-hlsurl' from '%v'", url)
		}
	})

	if hlsURL == "" {
		log.Printf("Coulden't find Element '%v' from '%v'", target_elements, url)
	}

	return hlsURL
}

func main() {
	// コマンドライン引数の処理
	flag.Parse()
	args := flag.Args()
	if len(args) != 2 {
		log.Fatalf("Unexpected arguments %v\n", args)
	}
	url := args[0]
	title := args[1]
	output := "output/" + title

	// M3U8のURLを取得
	hlsURL := getEmbededHLSURL(url)
	fmt.Printf("HLS-URL: %v\n", hlsURL)

	// ffmpegでM3U8をダウンロード
	err := exec.Command("ffmpeg", "-i", hlsURL, "-write_xing", "0", output).Run()
	if err != nil {
		log.Panicf("%v\n", err)
	}

}
