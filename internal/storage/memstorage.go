package storage

import (
	"github.com/azaliaz/go-book/internal/domain/models"
	"github.com/azaliaz/go-book/internal/logger"
	storerrors "github.com/azaliaz/go-book/internal/storage/errors"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type MemStorage struct {
	userStor map[string]models.User
	bookStor map[string]models.Book
}

func New() *MemStorage {
	return &MemStorage{
		userStor: make(map[string]models.User),
		bookStor: make(map[string]models.Book),
	}
}
func (ms *MemStorage) SaveUser(user models.User) (string, error) {
	log := logger.Get()
	uuid := uuid.New().String()
	if _, err := ms.findUser(user.Email); err == nil {
		return "", storerrors.ErrUserExists
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Pass), bcrypt.DefaultCost)
	user.Pass = string(hash)
	if err != nil {
		log.Error().Err(err).Msg("save user failed")

		return "", err
	}
	log.Debug().Str("hash", string(hash)).Send()
	user.UID = uuid
	ms.userStor[uuid] = user
	log.Debug().Any("userStorage", ms.userStor).Send()
	return uuid, nil
}
func (ms *MemStorage) ValidUser(user models.User) (string, error) {
	log := logger.Get()
	log.Debug().Any("userStorage", ms.userStor).Send()
	memuser, err := ms.findUser(user.Email)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(memuser.Pass), []byte(user.Pass)); err != nil {
		return "", storerrors.ErrInvalidPassword
	}
	return memuser.UID, nil
}

func (ms *MemStorage) GetUser(uid string) (models.User, error) {
	log := logger.Get()
	user, ok := ms.userStor[uid]
	if !ok {
		log.Error().Str("uid", uid).Msg("user not found")
		return models.User{}, storerrors.ErrorUserNotFound
	}
	return user, nil
}

func (ms *MemStorage) SaveBook(book models.Book) error {
	memBook, err := ms.findBook(book)
	if err == nil {
		memBook.Count++
		ms.bookStor[memBook.BID] = memBook
		return nil
	}
	book.Count = 1
	bid := uuid.New().String()
	ms.bookStor[bid] = book
	return nil
}

func (ms *MemStorage) findUser(login string) (models.User, error) {
	for _, user := range ms.userStor {
		if user.Email == login {
			return user, nil
		}
	}
	return models.User{}, storerrors.ErrUserDoesNotExists
}
func (ms *MemStorage) findBook(value models.Book) (models.Book, error) {
	for _, book := range ms.bookStor {
		if book.Lable == value.Lable && book.Author == value.Author {
			return book, nil
		}
	}
	return models.Book{}, storerrors.ErrBookDoesNotExists
}
func (ms *MemStorage) GetBooks() ([]models.Book, error) {
	var books []models.Book
	for _, book := range ms.bookStor {
		books = append(books, book)
	}
	if len(books) < 1 {
		return nil, storerrors.ErrEmptyBookList
	}
	return books, nil
}
func (ms *MemStorage) GetBook(buid string) (models.Book, error) {
	log := logger.Get()
	book, ok := ms.bookStor[buid]
	if !ok {
		log.Error().Str("buid", buid).Msg("book not found")
		return models.Book{}, storerrors.ErrBookDoesNotExists
	}
	return book, nil
}
