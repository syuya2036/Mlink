package anylink
import (
    "fmt"
    "net/http"
	"encoding/json"
	"io"
)

// func main(){
// 	url := "https://api.song.link/v1-alpha.1/links?url=https://music.youtube.com/watch?v=37W7Y2RRyiM&feature=share&userCountry=JP"
// 	musicInfo := newMinfo(url)
// 	err := musicInfo.GetMusicUrls()

// 	if err != nil {
// 		fmt.Println("ajfjkld")
// 	}

// 	fmt.Println(musicInfo.Amazon)
// }

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


func newMinfo(url string) *Minfo {
	self := new(Minfo)
	self.Url = url

	return self
}


// jsonを解析して各種情報を取得
func (self *Minfo) readMap(songUrlMap map[string]interface {}) {
	linksByPlatforms := songUrlMap["linksByPlatform"].(map[string]interface {})
	infoByAmazon := songUrlMap["entitiesByUniqueId"].(map[string]interface {})["AMAZON_SONG::B07QWVM6V6"].(map[string]interface {})
	// linksByPlatform -> amazonMusic -> url
	self.Amazon = fmt.Sprintf("%v",linksByPlatforms["amazonMusic"].(map[string]interface {})["url"])
	self.Apple = fmt.Sprintf("%v",linksByPlatforms["appleMusic"].(map[string]interface {})["url"])
	self.Youtube = fmt.Sprintf("%v",linksByPlatforms["youtubeMusic"].(map[string]interface {})["url"])
	self.Spotify = fmt.Sprintf("%v",linksByPlatforms["spotify"].(map[string]interface {})["url"])
	self.Title = fmt.Sprintf("%v",infoByAmazon["title"])
	self.ArtistName = fmt.Sprintf("%v",infoByAmazon["artistName"])
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
		return err
	}

	body, _ := io.ReadAll(resp.Body)

	// jsonフォーマットをmapに変換
	var songUrlMap map[string]interface{}
	json.Unmarshal(body, &songUrlMap)

	self.readMap(songUrlMap)

	return nil
}