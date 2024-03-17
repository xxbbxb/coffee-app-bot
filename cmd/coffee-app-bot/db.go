package main

import (
	"coffee-app-bot/pkg/router"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

type UserStatus int

const (
	UserStatusNew UserStatus = iota
	UserStatusDisabled
	UserStatusActive
)

type ContactStatus int

const (
	ContactStatusNone ContactStatus = iota
	ContactStatusCoffee
	ContactStatusRejected
	ContactStatusMatched
)

type User struct {
	Id           int64      `db:"id"`
	ShownName    string     `db:"shownName"`
	Login        string     `db:"login"`
	Status       UserStatus `db:"status"`
	PhotoFileId  string     `db:"photoFileId"`
	Bio          string     `db:"bio"`
	CreatedAt    time.Time  `db:"createdAt"`
	LastActiveAt time.Time  `db:"lastActiveAt"`
}

type dbContact struct {
	UserId    int64         `db:"userId"`
	ContactId int64         `db:"contactId"`
	Status    ContactStatus `db:"status"`
	CreatedAt time.Time     `db:"createdAt"`
}

type Settings struct {
	Visibility  int
	OfflineCity string
	OnLineReady bool
}

type dbSetting struct {
	UserId       int64        `db:"userId"`
	SettingName  string       `db:"settingName"`
	SettingValue SettingValue `db:"settingValue"`
}

type SettingValue string

func (v SettingValue) MustBool() bool {
	if v == "yes" {
		return true
	}
	if v == "true" {
		return true
	}
	return false
}

func (v SettingValue) MustInt() int {
	if conv, err := strconv.Atoi(string(v)); err == nil {
		return conv
	}
	return 0
}

func (v SettingValue) MustString() string {
	return string(v)
}

type Database struct {
	conn *sqlx.DB
	log  *logrus.Logger
}

func NewDB(l *logrus.Logger, dsn string) (*Database, error) {

	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, err
	}
	return &Database{
		conn: db,
		log:  l,
	}, nil

}

func (db *Database) GetUser(id int64) (User, error) {
	u := User{}
	if err := db.conn.QueryRowx("SELECT * FROM users WHERE id=?", id).StructScan(&u); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, nil
		}
		return u, err
	}
	return u, nil
}

func (db *Database) AddOrUpdateUser(u User) error {
	tx := db.conn.MustBegin()
	defer tx.Rollback()
	_, err := tx.NamedExec(`
	    INSERT INTO users (id, login, shownName, bio, status, photoFileId)
		VALUES (:id, :login, :shownName, :bio, :status, :photoFileId)
		ON DUPLICATE KEY UPDATE
			login=:login,
			shownName=:shownName,
			bio=:bio,
			status=:status,
			photoFileId=:photoFileId`, u)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *Database) DeleteUser(u User) error {
	tx := db.conn.MustBegin()
	defer tx.Rollback()
	_, err := tx.NamedExec(`
	    DELETE FROM users WHERE id=:id`, u)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *Database) SetSettingValue(userId int64, name string, val string) error {
	s := dbSetting{
		UserId:       userId,
		SettingName:  name,
		SettingValue: SettingValue(val),
	}
	tx := db.conn.MustBegin()
	defer tx.Rollback()
	_, err := tx.NamedExec(`
	    INSERT INTO users_settings (userId, settingName, settingValue)
		VALUES (:userId, :settingName, :settingValue)
		ON DUPLICATE KEY UPDATE
			settingValue=:settingValue`, s)

	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *Database) DeleteSetting(userId int64, name string) error {
	tx := db.conn.MustBegin()
	defer tx.Rollback()
	_, err := tx.Exec(`DELETE FROM users_settings WHERE userId=? AND settingName=?`, userId, name)

	if err != nil {
		return err
	}
	return tx.Commit()
}

func (db *Database) GetSettings(userId int64) (*Settings, error) {
	res := Settings{
		Visibility:  100,
		OnLineReady: true,
	}
	rows, err := db.conn.Queryx("SELECT * FROM users_settings WHERE userId=? AND settingName in ('Visibility', 'OfflineCity', 'OnLineReady')", userId)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var s dbSetting
		err = rows.StructScan(&s)
		if err != nil {
			return nil, err
		}
		switch s.SettingName {
		case "Visibility":
			res.Visibility = s.SettingValue.MustInt()
		case "OfflineCity":
			res.OfflineCity = s.SettingValue.MustString()
		case "OnLineReady":
			res.OnLineReady = s.SettingValue.MustBool()
		}
	}

	return &res, nil
}

func (db *Database) GetMatchingIdsWithVisibility(userId int64) ([]int64, []int, error) {
	rows, err := db.conn.Queryx(`
	SELECT
		id,
		cast(ifnull((select settingValue from users_settings where userId = id and settingName = 'Visibility'),100) as unsigned) as visibility
	FROM users
	WHERE
		id <> ? AND status > 1 AND id not in (select contactId from users_contacts where userId = ?);`, userId, userId)
	if err != nil {
		return nil, nil, err
	}
	var Ids []int64
	var Visibility []int
	for rows.Next() {
		var id int64
		var vis int
		err := rows.Scan(&id, &vis)
		if err != nil {
			return nil, nil, err
		}
		Ids = append(Ids, id)
		Visibility = append(Visibility, vis)
	}
	return Ids, Visibility, nil
}

func (db *Database) GetContactsStatus(userId int64, contactId int64) (ContactStatus, error) {
	c := dbContact{}
	if err := db.conn.QueryRowx("SELECT * FROM users_contacts WHERE userId=? AND contactId=?", userId, contactId).StructScan(&c); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ContactStatusNone, nil
		}
		return ContactStatusNone, err
	}
	return c.Status, nil
}

func (db *Database) SetContactsStatus(userId int64, contactId int64, st ...ContactStatus) error {
	insert := `INSERT INTO users_contacts (userId, contactId, status) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE status=?`
	ids := []int64{userId, contactId}
	tx := db.conn.MustBegin()
	defer tx.Rollback()
	switch len(st) {
	case 1:
		_, err := tx.Exec(insert, userId, contactId, st[0], st[0])
		if err != nil {
			return err
		}
	case 2:
		for i := 0; i < len(st); i++ {
			_, err := tx.Exec(insert,
				ids[i], ids[1-i], st[i], st[1-i])
			if err != nil {
				return err
			}
		}

	default:
		return fmt.Errorf("unexpected statuses count: %v in SetContactsStatus", st)
	}

	return tx.Commit()
}

func CoffeeDBMiddleware(log *logrus.Logger, dsn string) func(next router.Handler) router.Handler {
	db, err := NewDB(log, dsn)
	if err != nil {
		log.WithError(err).Fatal("Unable to create connection to database")
	}
	return func(next router.Handler) router.Handler {
		return router.HandlerFunc(func(u *router.Update) {
			user, err := db.GetUser(u.ChatID())
			if err != nil {
				log.WithError(err).Errorf("Unable to read %d from DB", u.ChatID())
				return
			}
			if user.Id == 0 {
				user.Id = u.ChatID()
				user.Login = u.FromChat().UserName
			}
			ctx := u.Context()
			ctx = context.WithValue(ctx, router.ContextKey("tg-db-user"), user)
			ctx = context.WithValue(ctx, router.ContextKey("tg-db-user-id"), user.Id)
			ctx = context.WithValue(ctx, router.ContextKey("tg-db-conn"), db)

			next.Serve(u.WithContext(ctx))
		})
	}
}

func GetDB(ctx context.Context) *Database {
	if db, ok := ctx.Value(router.ContextKey("tg-db-conn")).(*Database); ok {
		return db
	}
	panic("no DB in context")
}

func AddOrUpdateUser(ctx context.Context, user User) error {
	if user.Status == UserStatusNew && user.ShownName != "" && user.Bio != "" {
		user.Status = UserStatusActive
	}
	return GetDB(ctx).AddOrUpdateUser(user)
}

func DeleteUser(ctx context.Context, user User) error {
	return GetDB(ctx).DeleteUser(user)
}

func GetUser(ctx context.Context) User {
	return ctx.Value(router.ContextKey("tg-db-user")).(User)
}

func IsUserProfileCompleted(u User) bool {
	if u.ShownName != "" && u.Bio != "" {
		return true
	}
	return false
}

func GetSettings(ctx context.Context) *Settings {
	u := GetUser(ctx)
	s, err := GetDB(ctx).GetSettings(u.Id)
	if err != nil {
		log.WithError(err).WithField("user-id", u.Id).Error("unable to read user settings")
	}
	return s
}

func SetSettingValue(ctx context.Context, name string, val string) error {
	u := GetUser(ctx)
	return GetDB(ctx).SetSettingValue(u.Id, name, val)
}

func DeleteSetting(ctx context.Context, name string) error {
	u := GetUser(ctx)
	return GetDB(ctx).DeleteSetting(u.Id, name)
}

func GetRandomUser(ctx context.Context) *User {
	u := GetUser(ctx)
	var s int
	ids, vis, err := GetDB(ctx).GetMatchingIdsWithVisibility(u.Id)
	if err != nil {
		log.WithError(err).Error("unable to read matching user ids")
		return nil
	}
	s = 0
	for i := 0; i < len(vis); i++ {
		s += vis[i]
	}

	if len(ids) == 0 {
		return nil
	}
	r := rand.Intn(s)
	s = 0
	for i := 0; i < len(vis); i++ {
		if r >= s && r < s+vis[i] {
			u, err = GetDB(ctx).GetUser(ids[i])
			if err != nil {
				log.WithError(err).Error("unable to read matching user")
				return nil
			}

			return &u
		}
		s += vis[i]
	}
	return nil
}

func RejectUser(ctx context.Context, contactId int64) error {
	u := GetUser(ctx)
	return GetDB(ctx).SetContactsStatus(u.Id, contactId, ContactStatusRejected)
}

func DeinviteUser(ctx context.Context, contactId int64) error {
	u := GetUser(ctx)
	return GetDB(ctx).SetContactsStatus(contactId, u.Id, ContactStatusNone)
}

func DoCoffeeUser(ctx context.Context, contactId int64) (*User, error) {
	u := GetUser(ctx)
	db := GetDB(ctx)
	// check if reversed invitation already exists
	st, err := db.GetContactsStatus(contactId, u.Id)
	if err != nil {
		return nil, err
	}
	if st == ContactStatusCoffee {
		c, err := db.GetUser(contactId)
		if err != nil {
			return nil, err
		}
		return &c, db.SetContactsStatus(u.Id, contactId, ContactStatusMatched, ContactStatusMatched)
	}
	return nil, db.SetContactsStatus(u.Id, contactId, ContactStatusCoffee)
}
