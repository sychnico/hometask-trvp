package main

import (
	"bibliographic_litriture_gigachat/internal/config"
	errs "bibliographic_litriture_gigachat/internal/errors"
	gigachatService "bibliographic_litriture_gigachat/internal/gigachat"
	"bibliographic_litriture_gigachat/internal/middleware"
	search "bibliographic_litriture_gigachat/internal/search"
	utils "bibliographic_litriture_gigachat/utils"
	"fmt"
	"time"

	"encoding/json"
	"net/http"

	"github.com/evgensoft/gigachat"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

type elibraryArticlesJSON struct {
	Title string `json:"title"`
	Link  string `json:"link"`
}

func handleElibrarySearch(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()
	query := params.Get("query")

	logger := log.Ctx(r.Context())

	if err := utils.SearchInputIsValid(query); err != nil {
		logger.Printf("utils.SearchInputIsValid error: %v", err)
		http.Error(w, "Недостаточно данных для поиска", http.StatusBadRequest)
		return
	}

	articles, err := search.Search(query, r)
	if err != nil {
		logger.Printf("Ошибка при выполнении запроса error: %v \n", err)
		http.Error(w, "Не удалось выполнить поиск", http.StatusBadRequest)
		return
	}

	var response []elibraryArticlesJSON

	for link, title := range articles {
		article := elibraryArticlesJSON{
			Title: title,
			Link:  link,
		}
		response = append(response, article)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	cfg, err := config.SetupNewConfig()
	if err != nil {
		log.Fatal().Err(fmt.Errorf("%w: %s", err, errs.ErrLoadConfig)).Msg("Could not setup server config")
	}

	gigachatClient := gigachat.NewClient(
		viper.GetString(config.GIGACHAT_CLIENT_ID),
		viper.GetString(config.GIGACHAT_CLIENT_SECRET),
	)
	// Получаем токен
	token, err := gigachatClient.GetToken()
	if err != nil {
		log.Fatal().Msg(fmt.Errorf("could not connect to GigaChat: %w", err).Error())
	}

	log.Info().Msg(fmt.Sprintf(
		"Connection to GigaChat was successfull.\n GigaChatStats:\n GigaChat Токен: %s\n GigaChat Срок действия до: %s\n",
		token.AccessToken,
		time.Unix(token.ExpiresAt, 0).Format(time.RFC3339),
	))

	gigaChatService := gigachatService.New(gigachatClient)

	r := mux.NewRouter()

	// middleware
	r.Use(middleware.RequestWithLoggerMiddleware)
	r.Use(middleware.PreventPanicMiddleware)
	r.Use(middleware.MiddlewareCors)

	r.HandleFunc("/api/request", gigaChatService.HandleForm).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/search_elibrary", handleElibrarySearch).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/api/requestMultyRow", gigaChatService.HandleFormMultyRow).Methods(http.MethodPost, http.MethodOptions)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Address, cfg.Server.Port),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
		Handler:      r,
	}

	log.Info().Msg(fmt.Sprintf("Server started on adress and port %s:%d", cfg.Server.Address, cfg.Server.Port))
	log.Fatal().Msg(fmt.Sprint(srv.ListenAndServe()))
}
