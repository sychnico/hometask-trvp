package bot

import (
	"bibliographic_litriture_gigachat/internal/gigachat"
	"bibliographic_litriture_gigachat/internal/search"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	updateTimeout = 60
	messagePrefix = "📚"
)

var (
	ErrTokenIsEmpty   = errors.New("TELEGRAM_TOKEN environment variable is empty")
	ErrInvalidCommand = errors.New("invalid command")
)

type TelegramBot struct {
	bot             *tgbotapi.BotAPI
	gigachatService *gigachat.Service
}

func New(gigachatService *gigachat.Service) (*TelegramBot, error) {
	token := os.Getenv("TELEGRAM_TOKEN")
	if token == "" {
		return nil, ErrTokenIsEmpty
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, fmt.Errorf("tgbotapi.NewBotAPI: %w", err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	return &TelegramBot{
		bot:             bot,
		gigachatService: gigachatService,
	}, nil
}

func (b *TelegramBot) Start() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = updateTimeout

	updates := b.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			go b.processMessage(update.Message)
		}
	}

	return nil
}

func (b *TelegramBot) processMessage(msg *tgbotapi.Message) {
	log.Printf("[%s] %s", msg.From.UserName, msg.Text)

	response, err := b.handleMessage(msg)
	if err != nil {
		log.Printf("Error handling message: %v", err)
		response = b.formatError(err)
	}

	if err := b.sendResponse(msg.Chat.ID, msg.MessageID, response); err != nil {
		log.Printf("Error sending response: %v", err)
	}
}

func (b *TelegramBot) handleMessage(msg *tgbotapi.Message) (string, error) {
	switch {
	case msg.IsCommand():
		return b.handleCommand(msg)
	default:
		return b.handleTextMessage(msg)
	}
}

func (b *TelegramBot) handleCommand(msg *tgbotapi.Message) (string, error) {
	switch msg.Command() {
	case "start":
		return b.handleStartCommand(), nil
	case "help":
		return b.handleHelpCommand(), nil
	case "search":
		return b.handleSearchCommand(msg.CommandArguments())
	case "generate_ref":
		return b.handleGenerateRefCommand(msg.CommandArguments())
	default:
		return "", ErrInvalidCommand
	}
}

func (b *TelegramBot) handleStartCommand() string {
	return fmt.Sprintf(`%s Добро пожаловать в бот библиографических ссылок!

Доступные команды:
/search <запрос> - поиск литературы
/generate_ref <запрос> <тип> <пример> - генерация ссылки
/help - помощь

Просто отправьте текст для поиска литературы!`, messagePrefix)
}

func (b *TelegramBot) handleHelpCommand() string {
	return fmt.Sprintf(`%s Помощь по командам:

/search <запрос>
Поиск литературы по ключевым словам
Пример: /search искусственный интеллект

/generate_ref "<запрос>" "<тип>" "<пример>"
Генерация библиографической ссылки
Пример: /generate_ref "Иванов А.И. Машинное обучение. М.: Наука, 2020." "книга" "Исследование механизма балансировки нагрузки многосерверной сетевой системы на основе теории Марковских процессов / Т. Н. Моисеев, О. Я. Кравец // Информационные технологии моделирования и управления. – 2005."

Просто отправьте текст для автоматического поиска!`, messagePrefix)
}

func (b *TelegramBot) handleSearchCommand(query string) (string, error) {
	if strings.TrimSpace(query) == "" {
		return "Пожалуйста, укажите запрос для поиска\nПример: /search искусственный интеллект", nil
	}

	references, err := search.Search(query, nil)
	if err != nil {
		return "", fmt.Errorf("search.Search: %w", err)
	}

	return b.formatSearchResults(references), nil
}

func (b *TelegramBot) handleGenerateRefCommand(args string) (string, error) {
	parts := strings.Split(args, "\"")
	var cleanParts []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			cleanParts = append(cleanParts, trimmed)
		}
	}

	if len(cleanParts) < 3 {
		return "Недостаточно аргументов. Используйте: /generate_ref \"<запрос>\" \"<тип>\" \"<пример>\"", nil
	}

	req := gigachat.FormRequest{
		UserRequest:   cleanParts[0],
		PromptType:    cleanParts[1],
		ExampleRecord: cleanParts[2],
	}

	resp, err := b.gigachatService.SendRequest(req)
	if err != nil {
		return "", fmt.Errorf("gigachatService.SendRequest: %w", err)
	}

	return b.formatGigaChatResponse(resp.Answer), nil
}

func (b *TelegramBot) handleTextMessage(msg *tgbotapi.Message) (string, error) {
	text := strings.TrimSpace(msg.Text)
	if text == "" {
		return "Пожалуйста, введите запрос для поиска литературы", nil
	}

	references, err := search.Search(text, nil)
	if err != nil {
		return "", fmt.Errorf("search.Search: %w", err)
	}

	return b.formatSearchResults(references), nil
}

func (b *TelegramBot) formatSearchResults(data map[string]string) string {
	if len(data) == 0 {
		return "🔍 По вашему запросу ничего не найдено"
	}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("%s Найдены следующие источники:\n\n", messagePrefix))

	for key, value := range data {
		result.WriteString(fmt.Sprintf("📖 *%s*\n%s\n\n", key, value))
	}

	return result.String()
}

func (b *TelegramBot) formatGigaChatResponse(response string) string {
	return fmt.Sprintf("%s Сгенерированная библиографическая ссылка:\n\n%s", messagePrefix, response)
}

func (b *TelegramBot) formatError(err error) string {
	switch {
	case errors.Is(err, ErrInvalidCommand):
		return "Неизвестная команда. Используйте /help для списка команд"
	default:
		return "Произошла ошибка при обработке запроса. Попробуйте позже"
	}
}

func (b *TelegramBot) sendResponse(chatID int64, replyToID int, text string) error {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyToID

	_, err := b.bot.Send(msg)
	return err
}
