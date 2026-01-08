package utils

// import (
// 	"fmt"
// 	"log"

// 	gigachat "github.com/evgensoft/gigachat"
// )

// func main() {
// 	// Создаем клиент (по умолчанию используется ScopePersonal)
// 	client := gigachat.NewClient(os.Getenv("GIGACHAT_CLIENT_ID"), os.Getenv("GIGACHAT_CLIENT_SECRET"))

// 	// Получаем токен
// 	token, err := client.GetToken()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	fmt.Printf("Токен: %s\n", token.AccessToken)
// 	fmt.Printf("Срок действия до: %d\n", token.ExpiresAt)

// 	client.Chat(&gigachat.ChatRequest{
// 		Model:    gigachat.ModelGigaChat,
// 		Messages: []gigachat.Message{},
// 	})
// }
