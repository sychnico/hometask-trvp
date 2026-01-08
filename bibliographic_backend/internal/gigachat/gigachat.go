package gigachat

import (
	"bibliographic_litriture_gigachat/utils"
	"encoding/json"
	"fmt"

	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/evgensoft/gigachat"
)

const chunkSize = 8

// Структура формы
type FormRequest struct {
	UserRequest   string `json:"user_request"`
	PromptType    string `json:"prompt_type,omitempty"`
	ExampleRecord string `json:"example_record,omitempty"`
}

type FormResponse struct {
	Answer string `json:"answer"`
}

type Service struct {
	Client *gigachat.Client
}

func New(cl *gigachat.Client) *Service {
	return &Service{Client: cl}
}

func (s *Service) HandleForm(w http.ResponseWriter, r *http.Request) {
	var req FormRequest

	logger := log.Ctx(r.Context())

	// Парсим входной JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		logger.Error().Msg(fmt.Sprint(err))
		return
	}

	incomingData := req.UserRequest
	if err := utils.FormatInputIsValid(incomingData); err != nil {
		logger.Error().Msg(fmt.Sprintf("utils.FormatInputIsValid: %v", err))
		http.Error(w, "Недостаточно данных", http.StatusBadRequest)
		return
	}

	response, err := s.SendRequest(req)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос", http.StatusInternalServerError)
		logger.Error().Msg(fmt.Sprintf("gptServerClient.SendRequest: %v", err))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Service) HandleFormMultyRow(w http.ResponseWriter, r *http.Request) {
	var req FormRequest

	logger := log.Ctx(r.Context())

	// Парсим входной JSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		logger.Error().Msg(fmt.Sprint(err))
		return
	}

	incomingData := req.UserRequest

	if err := utils.FormatInputIsValid(incomingData); err != nil {
		logger.Error().Msg(fmt.Sprintf("utils.FormatInputIsValid: %v", err))
		http.Error(w, "Недостаточно данных", http.StatusBadRequest)
		return
	}

	unformedLinks, err := splitUserInputText(incomingData)
	if err != nil {
		logger.Error().Msg(fmt.Sprintf("utils.FormatInputIsValid: %v", err))
		http.Error(w, "Слишком длинный запрос", http.StatusBadRequest)
		return
	}

	// unformedLinks := []string{
	// `IEEE/ISO/IEC 26515-2018 "International Standard – Systems and software engineering – Developing information for users in an agile environment". – URL: https://standards.ieee.org/ieee/1363/6936/ (дата обращения: 25.09.2025).`,
	// `Федеральное агентство по техническому регулированию и метрологии. ГОСТ Р ИСО/МЭК 12207–2010 «Процессы жизненного цикла программных средств». – Москва: Стандартинформ, 2011. – 105 с.`,
	// `Бэрри У. Бём, TRW Defense Systems Group. Спиральная модель разработки и сопровождения программного обеспечения. – IEEE Computer Society Publications, 1986. – 26 с.`,
	// }

	typeStrings := make([]string, 0)

	chunkLinks := ChunkStrings(unformedLinks, chunkSize)

	if req.PromptType == "" {
		for _, chunk := range chunkLinks {
			temp, err := s.IdentifyTypes(chunk)
			if err != nil {
				http.Error(w, "Не удалось выполнить запрос", http.StatusInternalServerError)
				logger.Error().Msg(fmt.Sprintf("gptServerClient.IdentifyTypes: %v", err))
				return
			}
			typeStrings = append(typeStrings, temp...)
		}

	} else {
		for i := 0; i < len(unformedLinks); i++ {
			typeStrings = append(typeStrings, req.PromptType)
		}
	}

	responseStrings, err := s.SendMultipleRequest(unformedLinks, typeStrings)
	if err != nil {
		http.Error(w, "Не удалось выполнить запрос", http.StatusInternalServerError)
		logger.Error().Msg(fmt.Sprintf("gptServerClient.SendMultipleRequest: %v", err))
		return
	}

	response := FormResponse{
		Answer: fmt.Sprintf("Библиографические ссылки:\n%s", responseStrings),
	}

	// fmt.Println("-----------Format results--------------")
	// fmt.Println(response.Answer)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func splitUserInputText(userInputText string) ([]string, error) {
	if err := utils.FormatInputIsValid(userInputText); err != nil {
		return nil, fmt.Errorf("%w:%w", utils.ErrInputTooLong, err)
	}

	formatLinks := strings.Split(userInputText, "\n")

	var resultLinks []string // результирующий срез
	for _, s := range formatLinks {
		if strings.TrimSpace(s) != "" { // проверка на наличие непустого содержимого
			resultLinks = append(resultLinks, s)
		}
	}

	// if err := utils.FormatLinksIsValid(formatLinks); err != nil {
	// 	return nil, fmt.Errorf("%w:%w", utils.ErrInputTooLong, err)
	// }

	return resultLinks, nil
}

func ChunkStrings(arr []string, n int) [][]string {
	// Проверка граничных случаев
	if n <= 0 {
		return [][]string{}
	}

	if len(arr) == 0 {
		return [][]string{}
	}

	if n >= len(arr) {
		return [][]string{append([]string{}, arr...)}
	}

	var result [][]string
	for i := 0; i < len(arr); i += n {
		end := i + n
		if end > len(arr) {
			end = len(arr)
		}

		// Создаем копию чанка, чтобы избежать изменений исходного массива
		chunk := make([]string, end-i)
		copy(chunk, arr[i:end])
		result = append(result, chunk)
	}

	return result
}

func (s *Service) SendRequest(req FormRequest) (FormResponse, error) {
	directive, userMessage := buildPrompt(req)

	// Запрос в GigaChat
	chatReq := &gigachat.ChatRequest{
		Model: gigachat.ModelGigaChat,
		Messages: []gigachat.Message{
			{
				Role:    gigachat.RoleSystem,
				Content: directive,
			},
			{
				Role:    gigachat.RoleUser,
				Content: userMessage,
			},
		},
	}

	// Отправляем запрос
	resp, err := s.Client.Chat(chatReq)
	if err != nil {
		return FormResponse{}, fmt.Errorf("s.Client.Chat: %w", err)
	}

	// Формируем ответ
	response := FormResponse{
		Answer: "",
	}

	for _, choice := range resp.Choices {
		response.Answer += fmt.Sprintf("Библиографическая запись:\n%s\n", choice.Message.Content)
	}

	// format answer
	currentDate := time.Now().Format("02.01.2006") // текущая дата в формате DD.MM.YYYY

	// заменяем все вхождения "DD.MM.YYYY" на текущую дату
	response.Answer = strings.ReplaceAll(response.Answer, "DD.MM.YYYY", currentDate)

	// no spaces between word and :
	response.Answer = strings.ReplaceAll(response.Answer, " \u003a", "\u003a")

	// correct En Dash as in GOST
	response.Answer = strings.ReplaceAll(response.Answer, "\u2014", "\u2013")

	// correct URL as in GOST
	if !strings.Contains(response.Answer, "[Электронный ресурс] \u2013 URL:") {
		response.Answer = strings.ReplaceAll(response.Answer, "URL:", "[Электронный ресурс] \u2013 URL:")
	}

	response.Answer = formatDateReference(response.Answer)
	// response.Answer = "Исследование механизма балансировки нагрузки многосерверной сетевой системы на основе теории Марковских процессов / Т. Н. Моисеев, О. Я. Кравец // Информационные технологии моделирования и управления. – 2005."

	return response, nil
}

// Форматируем дату обращения, добавляя круглые скобки,
// если указанный формат ещё не установлен
func formatDateReference(input string) string {
	// 1. Регулярное выражение для поиска строки вида "(дата обращения: выбран пользователем)"
	reWithParens := regexp.MustCompile(`(?i)\(дата обращения: [^)]*\)`)

	// 2. Регулярное выражение для поиска строки вида "дата обращения: выбран пользователем" (без скобок)
	reWithoutParens := regexp.MustCompile(`(?i)дата обращения: [^\n\r]*`)

	// Текущая дата в формате DD.MM.YYYY
	currentDate := time.Now().Format("02.01.2006")
	replacement := fmt.Sprintf("(дата обращения: %s)", currentDate)

	// Case 1: Заменяем, если есть строка со скобками
	if reWithParens.MatchString(input) {
		return reWithParens.ReplaceAllString(input, replacement)
	}

	// Case 2: Добавляем скобки, если есть строка без них
	if reWithoutParens.MatchString(input) {
		return reWithoutParens.ReplaceAllString(input, replacement)
	}

	// Возвращаем исходную строку, если ни один из шаблонов не найден
	return input
}

func buildPrompt(req FormRequest) (string, string) {
	// Собираем текст вопроса для GPT из формы
	userMessage := fmt.Sprintf("%s'%s'", GPTUserRequestAnnotationString, req.UserRequest)
	promptType := GPTDefaultRecordTypeString

	if req.PromptType != "" && req.PromptType != "Другой" {
		promptType = fmt.Sprintf("\nPrompt type: %s", req.PromptType)
	}

	// Если есть пример записи, добавляем к запросу
	if req.ExampleRecord != "" {
		userMessage = fmt.Sprintf("\nExample: %s. %s", req.ExampleRecord, userMessage)
	} else if example, ok := exampleMap[req.PromptType]; ok {
		userMessage = fmt.Sprintf("\nExample: %s. %s", example, userMessage)
	}

	directive := fmt.Sprintf("%s\n%s'%s'", GPTDirectiveString, GPTLibraryRecordTypeString, promptType)

	return directive, userMessage
}

//task13--------------------------------------------------------------------------------------------

func (s *Service) IdentifyTypes(unformedLinks []string) ([]string, error) {
	directive, userMessage := buildTypePrompt(unformedLinks)
	types, err := s.SendPromptRequest(directive, userMessage)
	if err != nil {
		return nil, err
	}
	return types, nil
}

func (s *Service) SendPromptRequest(directive string, userMessage string) ([]string, error) {

	// Запрос в GigaChat
	chatReq := &gigachat.ChatRequest{
		Model: gigachat.ModelGigaChat,
		Messages: []gigachat.Message{
			// {
			// 	Role:    gigachat.RoleSystem,
			// 	Content: directive,
			// },
			{
				Role:    gigachat.RoleUser,
				Content: directive + userMessage,
			},
		},
	}

	// Отправляем запрос
	resp, err := s.Client.Chat(chatReq)
	if err != nil {
		return nil, fmt.Errorf("s.Client.Chat: %w", err)
	}

	// Формируем ответ
	response := FormResponse{
		Answer: "",
	}

	for _, choice := range resp.Choices {
		response.Answer += strings.TrimSpace(choice.Message.Content)
	}

	// Разделение строки по ";"
	parts := strings.Split(response.Answer, ";")

	// Очистка каждого элемента от пробелов
	result := make([]string, len(parts))
	for i, part := range parts {
		result[i] = strings.TrimSpace(part)
	}

	fmt.Println("response.Answer and result are:")
	fmt.Println(response.Answer)
	fmt.Println(result)

	return result, nil
}

func buildTypePrompt(unformedLinks []string) (string, string) {
	userMessage := fmt.Sprintf("%s %d. %s'", GPTUserTypeRequestAnnotationCountTotal, len(unformedLinks), GPTUserTypeRequestAnnotationString)
	rawLinks := make([]string, len(unformedLinks))
	for _, unformedStr := range unformedLinks {
		rawLinks = append(rawLinks, strings.TrimSpace(unformedStr))
	}
	userMessage += strings.Join(rawLinks, ";\n") + "'"

	directive := GPTTypeDirectiveString

	return directive, userMessage
}

//task14 вариант с кучей промптов, исправить потом

func (s *Service) SendMultipleRequest(unformedLinks []string, types []string) (string, error) {
	result := ""
	fmt.Println("unformed links and types length")
	fmt.Println(len(unformedLinks), len(types))
	for i, link := range unformedLinks {
		req := FormRequest{
			UserRequest: link,
			PromptType:  types[i],
		}
		resp, err := s.SendRequest(req)
		if err != nil {
			return "nil", fmt.Errorf("s.Client.Chat: %w", err)
		}
		result += (strconv.Itoa(i+1) + ". " + strings.TrimPrefix(resp.Answer, "Библиографическая запись:\n"))
	}
	return result, nil
}

// // GetToken получает токен доступа
// func (s *Service) ResetChat() error {
// 	dataScope := fmt.Sprintf("scope=%s", gigachat.ScopePersonal)

// 	// Создаем запрос
// 	httpReq, err := http.NewRequest("POST", fmt.Sprintf("%s/reset", baseURLToken), bytes.NewBufferString(dataScope))
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка создания запроса: %w", err)
// 	}

// 	// Устанавливаем заголовки
// 	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
// 	httpReq.Header.Set("Accept", "application/json")
// 	httpReq.Header.Set("Authorization", c.basicAuth)
// 	httpReq.Header.Set("RqUID", generateUUID())

// 	// Отправляем запрос
// 	resp, err := c.httpClient.Do(httpReq)
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка отправки запроса: %w", err)
// 	}
// 	defer resp.Body.Close()

// 	// Читаем тело ответа
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return nil, fmt.Errorf("ошибка чтения ответа: %w", err)
// 	}

// 	// Проверяем статус ответа
// 	if resp.StatusCode != http.StatusOK {
// 		var errResp ErrorResponse
// 		if err := json.Unmarshal(body, &errResp); err != nil {
// 			return nil, fmt.Errorf("ошибка разбора ответа об ошибке: %w", err)
// 		}
// 		return nil, &APIError{
// 			Code:    errResp.Code,
// 			Message: errResp.Message,
// 		}
// 	}

// 	// Разбираем успешный ответ
// 	var tokenResp TokenResponse
// 	if err := json.Unmarshal(body, &tokenResp); err != nil {
// 		return nil, fmt.Errorf("ошибка разбора ответа: %w", err)
// 	}

// 	// Сохраняем токен и время его истечения
// 	c.token = &tokenResp
// 	c.tokenExpiry = time.Unix(tokenResp.ExpiresAt/1000-60, 0)

// 	return &tokenResp, nil
// }
