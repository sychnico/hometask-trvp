package search

import (
	"bytes"

	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/rs/zerolog/log"

	"golang.org/x/net/html"
)

const MaxIDsToProcess = 7

func Search(title string, r *http.Request) (map[string]string, error) {

	logger := log.Ctx(r.Context())

	queryParams := url.Values{}
	queryParams.Set("where_fulltext", "on")
	queryParams.Set("where_name", "on")
	queryParams.Set("where_abstract", "on")
	queryParams.Set("where_keywords", "on")
	queryParams.Set("where_affiliation", "")
	queryParams.Set("where_references", "")
	queryParams.Set("type_article", "on")
	queryParams.Set("type_disser", "on")
	queryParams.Set("type_book", "on")
	queryParams.Set("type_report", "on")
	queryParams.Set("type_conf", "on")
	queryParams.Set("type_patent", "on")
	queryParams.Set("type_preprint", "on")
	queryParams.Set("type_grant", "on")
	queryParams.Set("type_dataset", "on")
	queryParams.Set("search_freetext", "")
	queryParams.Set("search_morph", "on")
	queryParams.Set("search_fulltext", "")
	queryParams.Set("search_open", "")
	queryParams.Set("search_results", "")
	queryParams.Set("titles_all", "")
	queryParams.Set("authors_all", "")
	queryParams.Set("rubrics_all", "")
	queryParams.Set("queryboxid", "")
	queryParams.Set("itemboxid", "")
	queryParams.Set("begin_year", "")
	queryParams.Set("end_year", "")
	queryParams.Set("issues", "all")
	queryParams.Set("orderby", "rank")
	queryParams.Set("order", "rev")
	queryParams.Set("changed", "1")
	queryParams.Set("ftext", title)

	// Формируем URL с параметрами
	requrl := "https://www.elibrary.ru/query_results.asp?"
	fullURL := requrl + queryParams.Encode()

	jar, _ := cookiejar.New(nil)
	client := &http.Client{Jar: jar}

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		fmt.Println("Ошибка формирования запроса:", err)
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept-Language", "ru-RU,ru;q=0.9,en-US;q=0.8,en;q=0.7")
	req.Header.Set("DNT", "1")                 // Do Not Track
	req.Header.Set("Connection", "keep-alive") // поддерживает постоянное соединение
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Referer", "https://site.ru/")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Ошибка при выполнении запроса error: %v \n", err))
		return nil, err
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Ошибка чтения ответа error: %v\n", err))
		return nil, err
	}

	htmlContent := string(bodyBytes)

	ids, err := extractIDs(htmlContent)
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("Ошибка извлечения id error: %v\n", err))
		return nil, err
	}

	references := make(map[string]string)
	//fmt.Println(ids)

	spans, _ := ExtractSpanTexts(htmlContent)

	for i, id := range ids[:min(len(ids), MaxIDsToProcess)] {

		result := spans[i*2]
		link := "https://www.elibrary.ru/item.asp?id=" + id
		references[link] = result

	}

	// for idx, text := range spans {
	// 	fmt.Printf("[%d]: %s\n", idx, text)
	// }

	client.Jar.SetCookies(&url.URL{Scheme: "https", Host: "https://elibrary.ru"}, nil) // Сбрасываем cookie
	client.CloseIdleConnections()

	return references, nil
}

func ExtractSpanTexts(htmlContent string) ([]string, error) {
	// Создаем reader из строки HTML
	reader := strings.NewReader(htmlContent)

	// Парсим HTML
	doc, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("ошибка при разборе HTML: %w", err)
	}

	// Список для хранения текстов из тегов <span>
	var spans []string

	// Рекурсивная функция для обхода узлов
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "span" {
			spans = append(spans, innerText(n))
		}

		// Рекурсивно идём по дочерним узлам
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}

	// Начинаем обход дерева с корневого узла
	traverse(doc)

	return spans, nil
}

// Внутренняя функция для извлечения текста из узла
func innerText(n *html.Node) string {
	var buf strings.Builder

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			buf.WriteString(n.Data)
			buf.WriteString(" ")
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}

	walk(n)

	str := strings.TrimSpace(buf.String())
	return strings.ReplaceAll(str, "\n", "")
}

func extractIDs(htmlContent string) ([]string, error) {
	r := bytes.NewBufferString(htmlContent)
	doc, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	var ids []string

	var f func(*html.Node)
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" && strings.HasPrefix(attr.Val, "/item.asp?id=") {
					id := strings.TrimPrefix(attr.Val, "/item.asp?id=")
					ids = append(ids, id)
				}
			}
		}

		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)

	return ids, nil
}
