package model

import (
	"crypto/sha256"
	"strconv"
	"time"
	"unicode"

	"github.com/Code-Hex/saltissimo"
	"github.com/Code-Hex/vegeta/internal/utils"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type (
	Users []*User
)

type User struct {
	gorm.Model
	Admin    bool   `gorm:"not null"`
	Name     string `gorm:"not null;index:idx_name"`
	Password string `gorm:"not null"`
	Salt     string `gorm:"not null"`
	Token    string `gorm:"not null"`
	Tags     []Tag  `gorm:"ForeignKey:UserID"`
}

type Tag struct {
	gorm.Model
	UserID   uint   `gorm:"not null"`
	Name     string `gorm:"not null;index:idx_name"`
	SomeData []Data `gorm:"ForeignKey:TagID"`
}

type Data struct {
	ID        uint       `json:"-" gorm:"primary_key"`
	CreatedAt time.Time  `json:"-"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"-" sql:"index"`

	TagID      uint   `json:"-" gorm:"not null"`
	RemoteAddr string `json:"remote_addr" gorm:"not null"`
	Hostname   string `json:"hostname" gorm:"not null"`
	Payload    string `json:"payload" gorm:"not null" sql:"type:text;"`
}

// Completed modeles

func CreateUser(db *gorm.DB, name, password string, isAdmin bool) (*User, error) {
	user := &User{}
	if user.AlreadyExist(db, name) {
		return nil, errors.New("User " + name + " already exist")
	}
	hashed, key, err := saltissimo.HexHash(sha256.New, password)
	if err != nil {
		return nil, err
	}
	user.Name = name
	user.Password = hashed
	user.Salt = key
	user.Token = utils.GenerateUUID()
	user.Admin = isAdmin

	tx := db.Begin()
	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return user, nil
}

func EditUser(db *gorm.DB, userID string, isAdmin bool, resetPassword string) (*User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to validate user-id")
	}
	user := &User{}
	if db.First(user, id).RecordNotFound() {
		return nil, errors.Errorf("UserID: %d is not found", id)
	}
	if id != 1 {
		user.Admin = isAdmin
	}

	if resetPassword != "" {
		if _, err := user.UpdatePassword(db, resetPassword); err != nil {
			return nil, err
		}
		return user, nil
	}

	tx := db.Begin()
	if err := tx.Save(user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return user, nil
}

func DeleteUser(db *gorm.DB, userID string) (*User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to validate user-id")
	}
	if id == 1 {
		return nil, errors.New("Can not delete UserID: 1")
	}
	user := &User{}
	if db.First(user, id).RecordNotFound() {
		return nil, errors.Errorf("UserID: %d is not found", id)
	}
	tx := db.Begin()
	if err := db.Delete(user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return user, nil
}

func TokenAuth(db *gorm.DB, uuid string) (*User, error) {
	user := new(User)
	result := db.First(user, "token = ?", uuid)
	if err := result.Related(&user.Tags, "Tags").Error; err != nil {
		return nil, errors.Errorf("Failed to authenticate token: %s", uuid)
	}
	return user, nil
}

func BasicAuth(db *gorm.DB, name, pass string) (*User, error) {
	user := new(User)
	result := db.First(user, "name = ?", name)
	if result.RecordNotFound() {
		return nil, errors.Wrap(errors.New("Username mismatch"), "Invalid user")
	}
	hash, key := user.Password, user.Salt
	ok, err := saltissimo.CompareHexHash(sha256.New, pass, hash, key)
	if err != nil {
		return nil, errors.Wrap(err, "Invalid user")
	}
	if !ok {
		return nil, errors.Wrap(errors.New("Password mismatch"), "Invalid user")
	}
	if err := db.Model(user).Related(&user.Tags).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *User) UpdatePassword(db *gorm.DB, password string) (*User, error) {
	hashed, key, err := saltissimo.HexHash(sha256.New, password)
	if err != nil {
		return nil, err
	}
	u.Password = hashed
	u.Salt = key

	tx := db.Begin()
	if err := tx.Save(u).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return u, nil
}

func (u *User) AlreadyExist(db *gorm.DB, name string) bool {
	return !db.First(u, "name = ?", name).RecordNotFound()
}

func FindUserByName(db *gorm.DB, name string) (*User, error) {
	user := new(User)
	if !user.AlreadyExist(db, name) {
		return nil, errors.New("Not found: " + name)
	}
	if err := db.Model(user).Related(&user.Tags, "Tags").Error; err != nil {
		return nil, err
	}
	return user, nil
}

func IsValidString(str string) bool {
	if str == "" {
		return false
	}
	for _, c := range str {
		if !unicode.IsLetter(c) {
			return false
		}
	}
	return true
}

func (u *User) ReGenerateUserToken(db *gorm.DB) (*User, error) {
	u.Token = utils.GenerateUUID()
	tx := db.Begin()
	if err := tx.Save(u).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()
	return u, nil
}

func (u *User) AddTag(db *gorm.DB, name string) error {
	tag := &Tag{Name: name}
	if !IsValidString(tag.Name) {
		return errors.Errorf("Invalid tag name: %s", tag.Name)
	}
	if !db.Find(&Tag{}, "name = ? and user_id = ?", tag.Name, u.ID).RecordNotFound() {
		return errors.Errorf("Tag %s is already exist", tag.Name)
	}
	tx := db.Begin()
	asn := tx.Model(u).Association("Tags")
	if err := asn.Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := asn.Append(tag).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func (u *User) FindByTagName(db *gorm.DB, name string) (*Tag, error) {
	tag := &Tag{}
	result := db.Model(u).Related(&u.Tags, "Tags").Where("name = ?", name).Find(tag)
	if result.RecordNotFound() {
		return nil, errors.Errorf(`User %s's tag "%s" is not found`, u.Name, name)
	}
	return tag, nil
}

func FindTagByID(db *gorm.DB, id uint) (*Tag, error) {
	tag := new(Tag)
	if err := db.First(tag, id).Error; err != nil {
		return nil, err
	}
	if err := db.Model(tag).Related(&tag.SomeData).Error; err != nil {
		return nil, err
	}
	return tag, nil
}

func (t *Tag) AddData(db *gorm.DB, data Data) error {
	if !utils.IsValidIPAddress(data.RemoteAddr) {
		return errors.Errorf("Invalid ip address format: %s", data.RemoteAddr)
	}
	if !utils.IsValidJSON(data.Payload) {
		return errors.Errorf("Invalid json format: %s", data.Payload)
	}
	tx := db.Begin()
	asn := tx.Model(t).Association("SomeData")
	if err := asn.Error; err != nil {
		tx.Rollback()
		return err
	}
	if err := asn.Append(data).Error; err != nil {
		tx.Rollback()
		return err
	}
	tx.Commit()
	return nil
}

func GetUsers(db *gorm.DB) ([]*User, error) {
	var users []*User
	result := db.Find(&users)
	if err := result.Error; err != nil {
		return nil, err
	}
	return users, nil
}
