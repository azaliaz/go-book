package server

import (
	"errors"
	"net/http"

	"github.com/azaliaz/go-book/internal/config"
	"github.com/azaliaz/go-book/internal/domain/models"
	"github.com/azaliaz/go-book/internal/logger"
	storerrors "github.com/azaliaz/go-book/internal/storage/errors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type Storage interface {
	SaveUser(models.User) (string, error)
	ValidUser(models.User) (string, error)
	SaveBook(models.Book) error
	GetUser(string) (models.User, error)
	GetBooks() ([]models.Book, error)
	GetBook(string) (models.Book, error)
}

type Server struct {
	serv    *http.Server
	valid   *validator.Validate
	storage Storage
}

func New(cfg config.Config, stor Storage) *Server {
	server := http.Server{
		Addr: cfg.Addr,
	}
	valid := validator.New()
	return &Server{serv: &server, valid: valid, storage: stor}
}

func (s *Server) Run() error {
	log := logger.Get()
	router := gin.Default()
	router.GET("/", func(ctx *gin.Context) { ctx.String(200, "Hello") })

	users := router.Group("/users")
	{
		users.POST("/register", s.register)
		users.POST("/login", s.login)
		users.GET("/:id", s.userInfo)

	}
	books := router.Group("/books")
	{
		books.GET("/", s.allBooks)
		books.GET("/:id", s.bookInfo)

	}
	router.POST("/add-book", s.addBook)
	router.POST("/book-return", s.bookReturn)
	s.serv.Handler = router
	log.Info().Str("host", s.serv.Addr).Msg("server started")
	if err := s.serv.ListenAndServe(); err != nil {
		return err
	}
	return nil
}

func (s *Server) register(ctx *gin.Context) {
	log := logger.Get()
	var user models.User
	if err := ctx.ShouldBindBodyWithJSON(&user); err != nil {
		log.Error().Err(err).Msg("unmarshal body failed")
		ctx.String(http.StatusBadRequest, "incorrectly entered data")
		return
	}
	if err := s.valid.Struct(user); err != nil {
		log.Error().Err(err).Msg("validate user failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	uuid, err := s.storage.SaveUser(user)
	if err != nil {
		if errors.Is(err, storerrors.ErrUserExists) {
			log.Error().Msg(err.Error())
			ctx.String(http.StatusConflict, err.Error())
			return
		}
		log.Error().Err(err).Msg("save user failed")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Debug().Str("uuid", uuid).Send()

	ctx.String(http.StatusCreated, uuid)
}

func (s *Server) login(ctx *gin.Context) {
	log := logger.Get()
	var user models.User
	if err := ctx.ShouldBindBodyWithJSON(&user); err != nil {
		log.Error().Err(err).Msg("unmarshal body failed")
		ctx.String(http.StatusBadRequest, "incorrectly entered data")
		return
	}
	if err := s.valid.Struct(user); err != nil {
		log.Error().Err(err).Msg("validate user failed")
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	uuid, err := s.storage.ValidUser(user)
	if err != nil {
		if errors.Is(err, storerrors.ErrUserDoesNotExists) {
			log.Error().Err(err).Msg("user not found")
			ctx.String(http.StatusNotFound, "invalid login or password")
		}
		if errors.Is(err, storerrors.ErrInvalidPassword) {
			log.Error().Err(err).Msg("invalid pass")
			ctx.String(http.StatusUnauthorized, err.Error())
			return
		}

		log.Error().Err(err).Msg("validate user failed")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.String(200, "user %s are logined", uuid)
}
func (s *Server) userInfo(ctx *gin.Context) {
	//TODO: возможно быть возможно тольки про наличии токена
	//TODO: (когда пользователь вошел в систему)
	id := ctx.Param("id")
	user, err := s.storage.GetUser(id)
	if err != nil {
		if errors.Is(err, storerrors.ErrorUserNotFound) {
			ctx.String(http.StatusNotFound, err.Error())
			return
		}
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusNotFound, user)
}
func (s *Server) allBooks(ctx *gin.Context) {
	books, err := s.storage.GetBooks()
	if err != nil {
		if errors.Is(err, storerrors.ErrEmptyBookList) {
			ctx.String(http.StatusNotFound, err.Error())
			return
		}
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, books)
}

func (s *Server) addBook(ctx *gin.Context) {
	log := logger.Get()
	var book models.Book
	if err := ctx.ShouldBindBodyWithJSON(&book); err != nil {
		log.Error().Err(err).Msg("unmarshal body failed")
		ctx.String(http.StatusBadRequest, "incorrectly entered data")
		return
	}
	if err := s.storage.SaveBook(book); err != nil {
		log.Error().Err(err).Msg("save book failed")
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.String(http.StatusOK, "book %s %s was added", book.Author, book.Lable)

}
func (s *Server) bookInfo(ctx *gin.Context) {
	//TODO: возможно быть возможно тольки про наличии токена
	//TODO: (когда пользователь вошел в систему)
	id := ctx.Param("id")
	book, err := s.storage.GetBook(id)
	if err != nil {
		if errors.Is(err, storerrors.ErrBookDoesNotExists) {
			ctx.String(http.StatusNotFound, err.Error())
			return
		}
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
	ctx.JSON(http.StatusNotFound, book)
}
func (s *Server) bookReturn(ctx *gin.Context) {

}
