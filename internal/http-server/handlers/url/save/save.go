package save

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"url-shortener/internal/storage"
	"url-shortener/lib/logger/api/response"
	"url-shortener/lib/logger/sl"
	"url-shortener/lib/random"
)

const aliasLength = 8

type Request struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type Response struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

type URLSaver interface {
	SaveURL(urlToSave string, alias string) (int64, error)
}

func New(log *slog.Logger, urlSaver URLSaver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "handlers.url.save.New"

		log = log.With(
			slog.String("op", op),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		var req Request

		if err := render.DecodeJSON(r.Body, &req); err != nil {
			log.Error("failed to decode request body", sl.Err(err))
			render.JSON(w, r, response.Error("failed to decode request"))
			return
		}

		if err := validator.New().Struct(req); err != nil {
			ve := err.(validator.ValidationErrors)

			log.Error("invalid request", sl.Err(err))
			render.JSON(w, r, response.ValidationError(ve))
			return
		}

		alias := req.Alias
		if alias == "" {
			alias = random.NewAlias(aliasLength)
		}

		id, err := urlSaver.SaveURL(req.URL, alias)
		if err != nil {
			if errors.Is(err, storage.ErrURLExists) {
				log.Info("url already exists", slog.String("url", req.URL))
				render.JSON(w, r, response.Error("url already exists"))
				return
			}

			log.Error("failed to add url", sl.Err(err))
			render.JSON(w, r, response.Error("failed to add url"))
			return
		}

		_ = id // если пока не используешь — можно удалить, но оставляю как на видео

		log.Info("url added", slog.String("alias", alias))

		render.JSON(w, r, Response{
			Response: response.OK(),
			Alias:    alias,
		})
	}
}
