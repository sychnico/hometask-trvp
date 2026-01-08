package gigachat

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/evgensoft/gigachat"
)

var (
	expected_success_type_request = `Интернет-ресурс
Закон, нормативный акт и т.п.
Книга`
)

func TestSendPromptRequestGigachat(t *testing.T) {

	gigachatClient := gigachat.NewClient(os.Getenv("GIGACHAT_CLIENT_ID"), os.Getenv("GIGACHAT_CLIENT_SECRET"))

	gigaChatService := New(gigachatClient)

	t.Run("Success", func(t *testing.T) {
		unformedLinks := []string{
			`IEEE/ISO/IEC 26515-2018 "International Standard – Systems and software engineering – Developing information for users in an agile environment". – URL: https://standards.ieee.org/ieee/1363/6936/ (дата обращения: 25.09.2025).`,
			`Федеральное агентство по техническому регулированию и метрологии. ГОСТ Р ИСО/МЭК 12207–2010 «Процессы жизненного цикла программных средств». – Москва: Стандартинформ, 2011. – 105 с.`,
			`Бэрри У. Бём, TRW Defense Systems Group. Спиральная модель разработки и сопровождения программного обеспечения. – IEEE Computer Society Publications, 1986. – 26 с.`,
		}

		directive, userMessage := buildTypePrompt(unformedLinks)

		fmt.Printf("directive, userMessage are: '%v'\n '%v'\n", directive, userMessage)

		types, err := gigaChatService.SendPromptRequest(directive, userMessage)

		fmt.Printf("raw types are: '%v'\n", types)

		assert.NoError(t, err, "gptServerClient.SendRequest should not return an error")

		fmt.Printf("collected types are: '%v'\n'%v'\n'%v'\n", types[0], types[1], types[2])

		assert.Equal(t, expected_success_type_request, strings.Join(types, "\n"))
	})

}
