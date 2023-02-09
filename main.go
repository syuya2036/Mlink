package main

import (
	"fmt"
	"strings"
	"os"
	"log"
	"net/http"
	"io"
	"regexp"
	"encoding/json"
	"github.com/line/line-bot-sdk-go/linebot"
)

func main() {
	const ChannelSecret = os.Getenv("CHANNEL_SECRET")
	const AccessToken = os.Getenv("ACCESS_TOKEN")
	bot, err := linebot.New(
		ChannelSecret, // Channel Secret 
		AccessToken, // アクセストークン（ロングターム）
	)
	if err != nil {
		log.Fatal(err)
	}

	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/callback", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					replyMessage := create_message(message.Text)

					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
						log.Print(err)
					}
				}
			}
		}
	})
	
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}


// 音楽サイトの各リンクとタイトル・アーティスト
type Minfo struct {
	Url string
	Title string
	ArtistName string
	Amazon string
	Apple string
	Youtube string
	Spotify string
	Line string
}


func NewMinfo(url string) *Minfo {
	self := new(Minfo)
	self.Url = url

	return self
}


// jsonを解析して各種情報を取得
func (self *Minfo) readMap(songUrlMap map[string]interface {}) {
	links_by_platform := songUrlMap["linksByPlatform"].(map[string]interface {})

	for key, val := range links_by_platform {
		if key == "amazonMusic" {
			self.Amazon = fmt.Sprintf("%v", val.(map[string]interface {})["url"])
		} else if key == "appleMusic" {
			self.Apple = fmt.Sprintf("%v", val.(map[string]interface {})["url"])
		} else if key == "youtubeMusic" {
			self.Youtube = fmt.Sprintf("%v", val.(map[string]interface {})["url"])
		} else if key == "spotify" {
			self.Spotify = fmt.Sprintf("%v", val.(map[string]interface {})["url"])
		}
	}

	entities_by_unique_id := songUrlMap["entitiesByUniqueId"].(map[string]interface {})

	for _, val := range entities_by_unique_id {
		title := fmt.Sprintf("%v", val.(map[string]interface {})["title"])
		artistname := fmt.Sprintf("%v", val.(map[string]interface {})["artistName"])

		self.Title = strings.Replace(title, " ", "", -1)
		self.ArtistName = strings.Replace(artistname, " ", "", -1)

		self.Line = "https://music.line.me/webapp/search?query=" + self.Title + "%20" + self.ArtistName

		break
	}
}

func (self *Minfo) GetMusicUrls() error {
	// APIを叩く
	resp, err := http.Get(self.Url)

	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		err := fmt.Errorf("Status code %d", resp.StatusCode)
		log.Print(err)
		return err
	}

	body, _ := io.ReadAll(resp.Body)

	// jsonフォーマットをmapに変換
	var songUrlMap map[string]interface{}
	json.Unmarshal(body, &songUrlMap)

	self.readMap(songUrlMap)

	return nil
}


func is_url(req_str string) bool {
	re := regexp.MustCompile("^https.*")
    return re.MatchString(req_str)
}

func create_message(requested_message string) string {
	if is_url(requested_message) {
		request_url := "https://api.song.link/v1-alpha.1/links?url=" + requested_message + "&userCountry=JP"
	
		minfo := NewMinfo(request_url)
		err := minfo.GetMusicUrls()

		if err != nil{
			replyText := "リンクは存在しないよ！"
			return replyText
		}
		replyText := "Artist: " + minfo.ArtistName  + "\nTitle: " + minfo.Title + 
					"\nAmazon:\n" + minfo.Amazon + "\nApple:\n" + minfo.Apple + 
					"\nSpotify:\n"+ minfo.Spotify + "\nYoutube:\n" + minfo.Youtube + "\nLine:\n" + minfo.Line


		return replyText
	} else {
		replyText := "リンクを送ってね！"
		return replyText
}