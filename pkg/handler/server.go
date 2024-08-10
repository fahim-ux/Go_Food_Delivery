package handler

import (
	"Go_Food_Delivery/pkg/database"
	"Go_Food_Delivery/pkg/storage"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"log/slog"
	"os"
)

type Server struct {
	gin     *gin.Engine
	db      database.Database
	storage storage.ImageStorage
}

func (server *Server) Storage() storage.ImageStorage {
	return server.storage
}

func (server *Server) Engine() database.Database {
	return server.db
}

func (server *Server) Gin() *gin.Engine {
	return server.gin
}

func NewServer(db database.Database) *Server {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	ginEngine := gin.New()

	// Setting Logger & MultipartMemory
	ginEngine.Use(sloggin.New(logger))
	ginEngine.Use(gin.Recovery())
	ginEngine.MaxMultipartMemory = 8 << 20 // 8 MB

	localStoragePath := os.Getenv("LOCAL_STORAGE_PATH")
	if len(localStoragePath) > 0 {
		// Set static path
		ginEngine.Static(os.Getenv("STORAGE_DIRECTORY"), localStoragePath)
	}

	return &Server{
		gin:     ginEngine,
		db:      db,
		storage: storage.CreateImageStorage(os.Getenv("STORAGE_TYPE")),
	}
}

func (server *Server) Run() error {
	return server.gin.Run(":8080")
}
