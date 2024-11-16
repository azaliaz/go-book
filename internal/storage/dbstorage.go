package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/azaliaz/go-book/internal/domain/consts"
	"github.com/azaliaz/go-book/internal/domain/models"
	"github.com/azaliaz/go-book/internal/logger"
	storerrors "github.com/azaliaz/go-book/internal/storage/errors"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
 
	"github.com/jackc/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type DBStorage struct {
	conn *pgx.Conn
}

func NewDB(ctx context.Context, addr string) (*DBStorage, error) {
	conn, err := pgx.Connect(ctx, addr)
	if err != nil {
		return nil, err
	}
	return &DBStorage{
		conn: conn,
	}, nil
}
func (dbs *DBStorage) SaveUser(user models.User) (string, error) {
	log := logger.Get()
	uuid := uuid.New().String()
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Pass), bcrypt.DefaultCost)
	if err != nil {
		log.Error().Err(err).Msg("save user failed")
		return "", err
	}
	log.Debug().Str("hash", string(hash)).Send()
	user.Pass = string(hash)
	user.UID = uuid
	ctx, cancel := context.WithTimeout(context.Background(), consts.DBctxTimeout)
	defer cancel()
	_, err = dbs.conn.Exec(ctx, "INSERT INTO users (uid, email, pass, age) VALUES ($1, $2, $3, $4)",
		user.UID, user.Email, user.Pass, user.Age)
	if err != nil {
		var pgErr *pgconn.PgError //привели к типу
		if errors.As(err, &pgErr) {
			if pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) { //подходит ли она под этот тип
				return "", storerrors.ErrUserExists
			}
		}
		return "", err
	}
	return user.UID, nil
}
func (dbs *DBStorage) ValidUser(user models.User) (string, error) { //получение пользователя из бд
	log := logger.Get()
	ctx, cancel := context.WithTimeout(context.Background(), consts.DBctxTimeout)
	defer cancel()
	row := dbs.conn.QueryRow(ctx, "SELECT uid, email, pass FROM users WHERE email=$1", user.Email)
	var usr models.User
	if err := row.Scan(&usr.UID, &usr.Email, &usr.Pass); err != nil {
		log.Error().Err(err).Msg("failed scan db data")
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(usr.Pass), []byte(user.Pass)); err != nil {
		log.Error().Err(err).Msg("failed compare hash and password")
		return "", storerrors.ErrInvalidPassword
	}
	log.Debug().Any("db user", usr).Msg("user form data base")
	return usr.UID, nil

}

func (dbs *DBStorage) GetUser(uid string) (models.User, error) { //получение пользователя по uid
	log := logger.Get()
	ctx, cancel := context.WithTimeout(context.Background(), consts.DBctxTimeout)
	defer cancel()
	row := dbs.conn.QueryRow(ctx, "SELECT uid, email, pass, age FROM users WHERE uid=$1", uid)
	var usr models.User
	if err := row.Scan(&usr.UID, &usr.Email, &usr.Pass, &usr.Age); err != nil {
		log.Error().Err(err).Msg("failed scan db data")
		return models.User{}, err
	}
	log.Debug().Any("db user", usr).Msg("user form data base")
	return usr, nil
}


func (dbs *DBStorage) SaveBook(book models.Book) error {
	log := logger.Get()
	ctx, cancel := context.WithTimeout(context.Background(), consts.DBCtxTimeout)
	defer cancel()
	var bid string
	var count int
	log.Debug().Msgf("search book %s %s", book.Author, book.Lable)
	err := dbs.conn.QueryRow(ctx, `SELECT bid, count FROM books 
		WHERE lable=$1 AND author=$2`, book.Lable, book.Author).Scan(&bid, &count)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			bid := uuid.New().String()
			_, err := dbs.conn.Exec(ctx,
				`INSERT INTO books (bid, lable, author, "desc", age, count) 
				VALUES ($1, $2, $3, $4, $5, $6)`,
				bid, book.Lable, book.Author, book.Desc, book.Age, book.Count)
			if err != nil {
				log.Error().Err(err).Msg("save book failed")
				return nil
			}
			return nil
		}
		log.Error().Err(err).Msg("get book count failed")
		return err
	}
	log.Debug().Int("book count", count).Msg("book count")
	_, err = dbs.conn.Exec(ctx, "UPDATE books SET count=count + 1 WHERE bid=$1", bid)
	if err != nil {
		log.Error().Err(err).Msg("update book count failed")
		return err
	}
	return nil
	return nil
}

func (dbs *DBStorage) GetBooks() ([]models.Book, error) {
	log := logger.Get()
	ctx, cancel := context.WithTimeout(context.Background(), consts.DBctxTimeout)
	defer cancel()
	rows, err := dbs.conn.Query(ctx, "SELECT * FROM books")
	if err != nil {
		log.Error().Err(err).Msg("failed get all books from db")
	}
	var books []models.Book
	for rows.Next() {
		var book models.Book
		if err := rows.Scan(&book.BID, &book.BID, &book.Lable, &book.Author, &book.Desc, &book.Age, &book.Count); err != nil {
			log.Error().Err(err).Msg("failed to scan data from db")
			return nil, err
		}
		books = append(books, book)
	}
	return books, nil
}
func (dbs *DBStorage) GetBook(buid string) (models.Book, error) {
	log := logger.Get()
	ctx, cancel := context.WithTimeout(context.Background(), consts.DBctxTimeout)
	defer cancel()
	row := dbs.conn.QueryRow(ctx, `SELECT bid, lable, author, "desc", age, count FROM books WHERE bid = $1 AND deleted=false`, buid)
	var book models.Book
	if err := row.Scan(&book.BID, &book.BID, &book.Lable, &book.Author, &book.Desc, &book.Age, &book.Count); err != nil {
		log.Error().Err(err).Msg("failed to scan data from db")
		return models.Book{}, err
	}

	return book, nil
}

func Migrations(dbDsn string, migrationsPath string) error {
	log := logger.Get()
	migratePath := fmt.Sprintf("file://%s", migrationsPath)
	m, err := migrate.New(migratePath, dbDsn)
	if err != nil {
		return err
	}
	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info().Msg("no mirations apply")
			return nil
		}
		return err
	}
	log.Info().Msg("all mirations apply")
	return nil
}
