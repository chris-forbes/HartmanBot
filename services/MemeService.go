package services

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"samm-bot/model"
	"time"
)

func GetRandomMeme() string {
	response, err := http.Get("https://memes.cdn.cflabs.co.uk")
	if err != nil {
		fmt.Println("Failed to get content from ")
	}

	defer response.Body.Close()
	if response.StatusCode == 200 {
		body, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Http response error, ", err)
		}
		var data model.CdnMeme
		err = json.Unmarshal(body, &data)
		if err != nil {
			fmt.Println("Failed to unmarshal json data, ", err)
			return ""
		}
		response.Body.Close()

		source := rand.NewSource(time.Now().Unix())
		random := rand.New(source)
	GenImage:
		randomIndex := random.Intn(len(data))
		if data[randomIndex].Type != "file" {
			goto GenImage
		} else {
			return "https://memes.cdn.cflabs.co.uk/" + url.PathEscape(data[randomIndex].Name)
		}
	}
	return ""
}
