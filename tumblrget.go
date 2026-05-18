/*
The MIT License (MIT)

Copyright (c) 2026/05/16 TakuyaChaen

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.

*/

package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func fetchAndSave(urlStr string, folder string) error {
	// Parse URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}

	// Extract file name from url
	filename := path.Base(u.Path)
	if filename == "" || filename == "/" {
		filename = "index.html"
	}

	// create folder if not exists
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		if err := os.MkdirAll(folder, 0755); err != nil {
			return err
		}
	}

	// save path
	savePath := filepath.Join(folder, filename)

	// do nothing if file exists
	if _, err := os.Stat(savePath); err == nil {
		return nil
	}

	// HTTP GET
	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// read file result
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// write file
	if err := ioutil.WriteFile(savePath, body, 0644); err != nil {
		return err
	}

	return nil
}

func FetchURL(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func ExtractSrcTexts(text string) []string {
	var results []string
	startKey := `src=`
	endKey := `regular`

	for {
		start := strings.Index(text, startKey)
		if start == -1 {
			break
		}
		start += len(startKey)

		end := strings.Index(text[start:], endKey)
		if end == -1 {
			break
		}
		end += start

		content := text[start:end]
		results = append(results, content)

		text = text[end+len(endKey):]
	}

	return results
}

func ExtractLastURL(text string) string {
	re := regexp.MustCompile(`https://[^"'\s]+?\.(jpg|png)`)
	urls := re.FindAllString(text, -1)

	if len(urls) == 0 {
		return ""
	}
	return urls[len(urls)-1]
}

func tumblrUrl(account string, page int) string {
	return fmt.Sprintf("https://%s.tumblr.com/api/read?type=photo&num=20&start=%d", account, page)
}

func main() {
	// check argument number
	if len(os.Args) < 4 {
		fmt.Println("argument: account start_num end_num")
		return
	}

	account := os.Args[1]

	// check second argument
	start, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("second argument must be number")
		return
	}

	// check third argument
	end, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("third argument must be number")
		return
	}

	if start > end {
		fmt.Println("start must be little than end")
		return
	}

	for i := start; i <= end; i++ {
		time.Sleep(200 * time.Millisecond)
		url := tumblrUrl(account, i)

		//get image src
		html, err := FetchURL(url)
		if err != nil {
			panic(err)
		}

		srcs := ExtractSrcTexts(html)
		for _, src_text := range srcs {
			img_url := ExtractLastURL(src_text)
			time.Sleep(200 * time.Millisecond)
			// get image file and save
			err := fetchAndSave(img_url, account)
			if err != nil {
				fmt.Println("error:", err)
				return
			}

		}
	}

}
