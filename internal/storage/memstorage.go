package userStorage

import (
	"errors"
	"helloapp/internal/domain/models"
	"helloapp/internal/logger"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type MemuserStorage struct {
	userStor map[string]models.User
	bookStor map[string]models.Book
}

func New() *MemuserStorage {
	return &MemuserStorage{
		userStor: make(map[string]models.User),
		bookStor: make(map[string]models.Book),
	}
}
func (ms *MemuserStorage) SaveUser(user models.User) (string, error) {
	log := logger.Get()
	uuid := uuid.New().String()
	if _, err := ms.findUser(user.Email); err == nil {
		return "", errors.New("user already exists")
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
func (ms *MemuserStorage) ValidUser(user models.User) (string, error) {
	log := logger.Get()
	log.Debug().Any("userStorage", ms.userStor).Send()
	memuser, err := ms.findUser(user.Email)
	if err != nil {
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(memuser.Pass), []byte(user.Pass)); err != nil {
		return "", errors.New("invalid password")
	}
	return memuser.UID, nil
}
func (ms *MemuserStorage) SaveBook(book models.Book) error {
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

func (ms *MemuserStorage) findUser(login string) (models.User, error) {
	for _, user := range ms.userStor {
		if user.Email == login {
			return user, nil
		}
	}
	return models.User{}, errors.New("user does not exists")
}
func (ms *MemuserStorage) findBook(value models.Book) (models.Book, error) {
	for _, book := range ms.bookStor {
		if book.Lable == value.Lable && book.Author == value.Author {
			return book, nil
		}
	}
	return models.Book{}, errors.New("user does not exists")
}
