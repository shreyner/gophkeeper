package httphandlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jaevor/go-nanoid"
	"github.com/minio/minio-go/v7"
	"github.com/shreyner/gophkeeper/internal/server/pgk/contenttype"
	"go.uber.org/zap"

	"github.com/shreyner/gophkeeper/internal/server/middlewares"
	"github.com/shreyner/gophkeeper/internal/server/pgk/stoken"
)

func NewRouter(
	log *zap.Logger,
	stokenService *stoken.Service,
	s3minioClient *minio.Client,
) *chi.Mux {
	randID, _ := nanoid.Standard(36)

	r := chi.NewRouter()

	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	r.Use(chiMiddleware.Timeout(60 * time.Second))

	r.Use(middlewares.Authenticator(log, stokenService))

	r.
		With(chiMiddleware.AllowContentType(contenttype.ContentTypeBinary)).
		Put("/upload", func(wr http.ResponseWriter, r *http.Request) {
			_, ok := middlewares.GetTokenDataCtx(r.Context())

			if !ok {
				http.Error(wr, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}

			defer r.Body.Close()

			fileName := randID()

			uploadFileName, err := s3minioClient.PutObject(
				r.Context(),
				"vault",
				fileName,
				r.Body,
				-1,
				minio.PutObjectOptions{
					DisableMultipart: false,
				},
			)

			if err != nil {
				log.Error("can't upload object to s3", zap.Error(err))
				http.Error(wr, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}

			_, _ = fmt.Fprintln(wr, uploadFileName.Location)
			wr.WriteHeader(http.StatusOK)
		})

	r.Get("/", func(w http.ResponseWriter, wr *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintln(w, "Hello world")
	})

	return r
}
