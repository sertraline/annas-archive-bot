package main

import (
	"fmt"
	"html"
	"net/url"
	"strconv"
	"time"

	"strings"

	goquery "github.com/PuerkitoBio/goquery"
	colly "github.com/gocolly/colly/v2"
	tele "gopkg.in/telebot.v3"
)

type BookStorageItem struct {
	message tele.StoredMessage
	items   interface{}
	page    int
	maxPage int
	sender  int64
	expires time.Time
}

type BookItem struct {
	Meta      string
	Title     string
	Publisher string
	Authors   string
	URL       string
	Image     string
}

var (
	selector        = &tele.ReplyMarkup{}
	bookBtnReset    = selector.Data("🔄", "reset")
	bookBtnPrev     = selector.Data("⬅", "prev")
	bookBtnNext     = selector.Data("➡", "next")
	bookBtnBack     = selector.Data("Back", "back")
	bookBtnDownload = selector.Data("Download", "dl", "0")
	bookStorage     = make(map[int64]map[int]interface{})
)

func getReply(item *BookItem) string {
	reply := ""
	if item.Image != "" {
		reply = reply + fmt.Sprintf("<a href=\"%s\">\u200b</a>\n", item.Image)
	}
	if item.Title != "" {
		reply = reply + fmt.Sprintf("📎 <b>%s</b>\n\n", html.EscapeString(item.Title))
	}
	if item.Meta != "" {
		reply = reply + fmt.Sprintf("• %s\n", html.EscapeString(item.Meta))
	}
	if item.Title != "" {
		reply = reply + fmt.Sprintf("• %s\n", html.EscapeString(item.Title))
	}
	if item.Publisher != "" {
		reply = reply + fmt.Sprintf("• %s\n", html.EscapeString(item.Publisher))
	}
	if item.Authors != "" {
		reply = reply + fmt.Sprintf("• %s\n\n", html.EscapeString(item.Authors))
	}
	return reply
}

func BookPaginator(c tele.Context) error {
	if c.Message().Payload == "" {
		return nil
	}
	items, err := FindBook(c.Message().Payload)
	if err != nil || len(items) == 0 {
		return nil
	}

	c.Set("items", items)
	c.Set("page", 0)
	c.Set("maxPage", len(items))

	bookBtnNext = selector.Data(fmt.Sprintf("➡ %d", 2), "next")
	bookBtnDownload = selector.Data("Download", "dl", "1")
	selector.Inline(
		selector.Row(bookBtnNext),
		selector.Row(bookBtnDownload),
	)

	item := items[0]
	reply := getReply(item)

	m, _ := c.Bot().Send(c.Recipient(), reply, selector, tele.ModeHTML)

	_, ok := bookStorage[m.Chat.ID]
	if !ok {
		bookStorage[m.Chat.ID] = make(map[int]interface{})
	}

	bookStorage[m.Chat.ID][m.ID] = &BookStorageItem{
		message: tele.StoredMessage{ChatID: m.Chat.ID, MessageID: strconv.Itoa(m.ID)},
		items:   items,
		page:    1,
		maxPage: len(items),
		sender:  c.Message().Sender.ID,
		expires: time.Now().Local().Add(time.Hour * time.Duration(1)),
	}

	return c.Respond()
}

func ResetPage(c tele.Context) error {
	mc := c.Callback().Message

	_, ok := bookStorage[mc.Chat.ID]
	if !ok {
		return c.Respond()
	}
	bi, ok := bookStorage[mc.Chat.ID][mc.ID]
	if !ok {
		return c.Respond()
	}
	bookItem := bi.(*BookStorageItem)
	if bookItem.sender != c.Callback().Sender.ID {
		fmt.Println("ID don't match: ", bookItem.sender, c.Callback().Sender.ID)
		return c.Respond(&tele.CallbackResponse{
			Text: "This is not for you, you silly goober",
		})
	}

	items := bookItem.items.([]*BookItem)

	bookBtnNext = selector.Data(fmt.Sprintf("➡ %d", 2), "next")
	bookBtnDownload = selector.Data("Download", "dl", "1")
	selector.Inline(
		selector.Row(bookBtnNext),
		selector.Row(bookBtnDownload),
	)

	item := items[0]
	reply := getReply(item)

	m, err := c.Bot().Edit(bookItem.message, reply, selector, tele.ModeHTML)
	if err != nil {
		return c.Respond()
	}
	bookStorage[m.Chat.ID][m.ID] = &BookStorageItem{
		message: tele.StoredMessage{ChatID: m.Chat.ID, MessageID: strconv.Itoa(m.ID)},
		items:   items,
		page:    1,
		maxPage: len(items),
		sender:  c.Callback().Sender.ID,
		expires: time.Now().Local().Add(time.Hour * time.Duration(1)),
	}

	return c.Respond()
}

func BackPage(c tele.Context) error {
	mc := c.Callback().Message

	_, ok := bookStorage[mc.Chat.ID]
	if !ok {
		return c.Respond()
	}
	bi, ok := bookStorage[mc.Chat.ID][mc.ID]
	if !ok {
		return c.Respond()
	}
	bookItem := bi.(*BookStorageItem)
	if bookItem.sender != c.Callback().Sender.ID {
		fmt.Println("ID don't match: ", bookItem.sender, c.Callback().Sender.ID)
		return c.Respond(&tele.CallbackResponse{
			Text: "This is not for you, you silly goober",
		})
	}

	items := bookItem.items.([]*BookItem)
	page := bookItem.page
	maxPage := bookItem.maxPage

	bookBtnPrev = selector.Data(fmt.Sprintf("⬅ %d", page-1), "prev")
	bookBtnNext = selector.Data(fmt.Sprintf("➡ %d", page+1), "next")
	bookBtnDownload = selector.Data("Download", "dl", strconv.Itoa(page))
	if page > 1 && page < maxPage {
		selector.Inline(
			selector.Row(bookBtnReset, bookBtnPrev, bookBtnNext),
			selector.Row(bookBtnDownload),
		)
	} else if page >= maxPage {
		selector.Inline(
			selector.Row(bookBtnReset, bookBtnPrev),
			selector.Row(bookBtnDownload),
		)
	} else {
		selector.Inline(
			selector.Row(bookBtnNext),
			selector.Row(bookBtnDownload),
		)
	}

	item := items[page-1]
	reply := getReply(item)

	m, err := c.Bot().Edit(bookItem.message, reply, selector, tele.ModeHTML)
	if err != nil {
		return c.Respond()
	}
	bookStorage[m.Chat.ID][m.ID] = &BookStorageItem{
		message: tele.StoredMessage{ChatID: m.Chat.ID, MessageID: strconv.Itoa(m.ID)},
		items:   items,
		page:    page,
		maxPage: len(items),
		sender:  c.Callback().Sender.ID,
		expires: time.Now().Local().Add(time.Hour * time.Duration(1)),
	}

	return c.Respond()
}

func GetNextPage(c tele.Context) error {
	mc := c.Callback().Message

	_, ok := bookStorage[mc.Chat.ID]
	if !ok {
		return c.Respond()
	}
	bi, ok := bookStorage[mc.Chat.ID][mc.ID]
	if !ok {
		return c.Respond()
	}
	bookItem := bi.(*BookStorageItem)
	if bookItem.sender != c.Callback().Sender.ID {
		fmt.Println("ID don't match: ", bookItem.sender, c.Callback().Sender.ID)
		return c.Respond(&tele.CallbackResponse{
			Text: "This is not for you, you silly goober",
		})
	}

	items := bookItem.items.([]*BookItem)
	page := bookItem.page
	maxPage := bookItem.maxPage

	page = page + 1
	if page >= maxPage {
		page = maxPage
	}
	bookBtnPrev = selector.Data(fmt.Sprintf("⬅ %d", page-1), "prev")
	bookBtnNext = selector.Data(fmt.Sprintf("➡ %d", page+1), "next")
	bookBtnDownload = selector.Data("Download", "dl", strconv.Itoa(page))
	if page > 1 && page < maxPage {
		selector.Inline(
			selector.Row(bookBtnReset, bookBtnPrev, bookBtnNext),
			selector.Row(bookBtnDownload),
		)
	} else if page >= maxPage {
		selector.Inline(
			selector.Row(bookBtnReset, bookBtnPrev),
			selector.Row(bookBtnDownload),
		)
	} else {
		selector.Inline(
			selector.Row(bookBtnNext),
			selector.Row(bookBtnDownload),
		)
	}

	item := items[page-1]
	reply := getReply(item)

	m, err := c.Bot().Edit(bookItem.message, reply, selector, tele.ModeHTML)
	if err != nil {
		return c.Respond()
	}
	bookStorage[m.Chat.ID][m.ID] = &BookStorageItem{
		message: tele.StoredMessage{ChatID: m.Chat.ID, MessageID: strconv.Itoa(m.ID)},
		items:   items,
		page:    page,
		maxPage: len(items),
		sender:  c.Callback().Sender.ID,
		expires: time.Now().Local().Add(time.Hour * time.Duration(1)),
	}

	return c.Respond()
}

func GetPrevPage(c tele.Context) error {
	mc := c.Callback().Message

	_, ok := bookStorage[mc.Chat.ID]
	if !ok {
		return c.Respond()
	}
	bi, ok := bookStorage[mc.Chat.ID][mc.ID]
	if !ok {
		return c.Respond()
	}
	bookItem := bi.(*BookStorageItem)
	if bookItem.sender != c.Callback().Sender.ID {
		fmt.Println("ID don't match: ", bookItem.sender, c.Callback().Sender.ID)
		return c.Respond(&tele.CallbackResponse{
			Text: "This is not for you, you silly goober",
		})
	}

	items := bookItem.items.([]*BookItem)
	page := bookItem.page
	maxPage := bookItem.maxPage

	page = page - 1
	if page <= 0 {
		page = 0
	}
	bookBtnPrev = selector.Data(fmt.Sprintf("⬅ %d", page-1), "prev")
	bookBtnNext = selector.Data(fmt.Sprintf("➡ %d", page+1), "next")
	bookBtnDownload := selector.Data("Download", "dl", strconv.Itoa(page))
	if page > 1 && page < maxPage {
		selector.Inline(
			selector.Row(bookBtnReset, bookBtnPrev, bookBtnNext),
			selector.Row(bookBtnDownload),
		)
	} else if page >= maxPage {
		selector.Inline(
			selector.Row(bookBtnReset, bookBtnPrev),
			selector.Row(bookBtnDownload),
		)
	} else {
		selector.Inline(
			selector.Row(bookBtnNext),
			selector.Row(bookBtnDownload),
		)
	}

	item := items[page-1]
	reply := getReply(item)

	m, err := c.Bot().Edit(bookItem.message, reply, selector, tele.ModeHTML)
	if err != nil {
		return c.Respond()
	}
	bookStorage[m.Chat.ID][m.ID] = &BookStorageItem{
		message: tele.StoredMessage{ChatID: m.Chat.ID, MessageID: strconv.Itoa(m.ID)},
		items:   items,
		page:    page,
		maxPage: len(items),
		sender:  c.Callback().Sender.ID,
		expires: time.Now().Local().Add(time.Hour * time.Duration(1)),
	}

	return c.Respond()
}

func DownloadItem(c tele.Context) error {
	cd := c.Callback().Data
	mc := c.Callback().Message
	if cd == "" {
		c.Respond()
	}
	conv, err := strconv.Atoi(cd)
	if err != nil {
		c.Respond()
	}

	_, ok := bookStorage[mc.Chat.ID]
	if !ok {
		return c.Respond()
	}
	bi, ok := bookStorage[mc.Chat.ID][mc.ID]
	if !ok {
		return c.Respond()
	}
	bookItem := bi.(*BookStorageItem)
	if bookItem.sender != c.Callback().Sender.ID {
		fmt.Println("ID don't match: ", bookItem.sender, c.Callback().Sender.ID)
		return c.Respond(&tele.CallbackResponse{
			Text: "This is not for you, you silly goober",
		})
	}

	page := bookItem.page
	items := bookItem.items.([]*BookItem)
	item := items[conv-1]

	coll := colly.NewCollector(
		colly.Async(true),
	)

	urls := make([]string, 0)
	coll.OnHTML("a", func(e *colly.HTMLElement) {
		if strings.Contains(e.Attr("class"), "js-download-link") {
			if e.Attr("href") != "" {
				urls = append(urls, e.Attr("href"))
			}
		}
	})

	coll.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	fullURL := "https://annas-archive.org/" + item.URL
	coll.Visit(fullURL)
	coll.Wait()

	rows := make([]tele.Row, 0)
	rows = append(rows, selector.Row(bookBtnBack))
	fmt.Println("URLS list: ", urls)
	for i, u := range urls {
		// skip URLs that require authentication
		if strings.HasPrefix(u, "/fast_download") {
			continue
		}
		// these URLs require captcha verification
		if strings.HasPrefix(u, "/slow_download") {
			u = "https://annas-archive.org" + u
		}
		if len(rows) > 4 {
			break
		}

		rows = append(rows, selector.Row(selector.URL(fmt.Sprintf("Mirror #%d", i), u)))
	}

	selector.Inline(
		rows...,
	)

	reply := ""
	if item.Title != "" {
		reply = reply + fmt.Sprintf("📎 <b>%s</b>\n\n", html.EscapeString(item.Title))
	}
	if item.Meta != "" {
		reply = reply + fmt.Sprintf("• %s\n", html.EscapeString(item.Meta))
	}

	m, err := c.Bot().Edit(bookItem.message, reply, selector, tele.ModeHTML)
	if err != nil {
		fmt.Println(err)
		return c.Respond()
	}
	bookStorage[m.Chat.ID][m.ID] = &BookStorageItem{
		message: tele.StoredMessage{ChatID: m.Chat.ID, MessageID: strconv.Itoa(m.ID)},
		items:   items,
		page:    page,
		maxPage: len(items),
		sender:  c.Callback().Sender.ID,
		expires: time.Now().Local().Add(time.Hour * time.Duration(1)),
	}

	return c.Respond()
}

func FindBook(query string) ([]*BookItem, error) {
	c := colly.NewCollector(
		colly.Async(true),
	)

	bookList := make([]*string, 0)

	c.OnHTML("div", func(e *colly.HTMLElement) {
		if strings.Contains(e.Attr("class"), "h-[125]") {
			v, err := e.DOM.Html()
			if err != nil {
				fmt.Println(err)
				return
			}
			v = strings.TrimSpace(v)
			if strings.HasPrefix(v, "<!-") {
				v = strings.Trim(v, "<!-")
				v = strings.Trim(v, "->")
				v = html.UnescapeString(v)
			}

			bookList = append(bookList, &v)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL.String())
	})

	fullURL := "https://annas-archive.org/search?q=" + url.QueryEscape(query)
	c.Visit(fullURL)
	c.Wait()

	bookListParsed := make([]*BookItem, 0)

	for i := 0; i < len(bookList); i++ {
		var doc, _ = goquery.NewDocumentFromReader(strings.NewReader(*bookList[i]))
		doc.Find("a").Each(func(i int, s *goquery.Selection) {
			v := strings.Split(strings.TrimSpace(s.Text()), "\n")

			img := s.Find("img").AttrOr("src", "")
			if len(v) == 4 {
				bookListParsed = append(bookListParsed, &BookItem{
					Meta:      strings.TrimSpace(v[0]),
					Title:     strings.TrimSpace(v[1]),
					Publisher: strings.TrimSpace(v[2]),
					Authors:   strings.TrimSpace(v[3]),
					URL:       s.AttrOr("href", ""),
					Image:     img,
				})
			} else if len(v) == 3 {
				bookListParsed = append(bookListParsed, &BookItem{
					Meta:      strings.TrimSpace(v[0]),
					Title:     strings.TrimSpace(v[1]),
					Publisher: strings.TrimSpace(v[2]),
					URL:       s.AttrOr("href", ""),
					Image:     img,
				})
			} else if len(v) == 2 {
				bookListParsed = append(bookListParsed, &BookItem{
					Meta:  strings.TrimSpace(v[0]),
					Title: strings.TrimSpace(v[1]),
					URL:   s.AttrOr("href", ""),
					Image: img,
				})
			} else {
				bookListParsed = append(bookListParsed, &BookItem{
					Meta:  strings.TrimSpace(v[0]),
					URL:   s.AttrOr("href", ""),
					Image: img,
				})
			}

		})
	}

	return bookListParsed, nil

}
