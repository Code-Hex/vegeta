package vegeta

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"unicode"

	"github.com/Code-Hex/saltissimo"
	"github.com/Code-Hex/vegeta/internal/utils"
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type User struct {
	gorm.Model
	Name     string `gorm:"not null"`
	Password string `gorm:"not null"`
	Salt     string `gorm:"not null"`
	Token    string `gorm:"not null"`
	Tags     []Tag  `gorm:"ForeignKey:UserID"`
}

type Tag struct {
	gorm.Model
	UserID   uint   `gorm:"not null"`
	Name     string `gorm:"not null"`
	SomeData []Data `gorm:"ForeignKey:TagID"`
}

type Data struct {
	gorm.Model
	TagID      uint   `gorm:"not null"`
	RemoteAddr string `gorm:"not null"`
	DeviceID   string `gorm:"not null"`
	Payload    string `gorm:"not null" sql:"type:text;"`
}

func CreateUser(db *gorm.DB, name, password string) (*User, error) {
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
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func TokenAuth(db *gorm.DB, uuid string) (*User, error) {
	user := new(User)
	result := db.First(user, "token = ?", uuid)
	if err := result.Error; err != nil {
		return nil, errors.Wrap(err, "Invalid user")
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

func (u *User) UpdatePassword(db *gorm.DB, name, password string) (*User, error) {
	hashed, key, err := saltissimo.HexHash(sha256.New, password)
	if err != nil {
		return nil, err
	}
	find := db.Model(u).Where("name = ?", name)
	if err := find.Error; err != nil {
		return nil, err
	}
	u.Password = hashed
	u.Salt = key
	if err := db.Save(u).Error; err != nil {
		return nil, err
	}
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

func (u *User) AddTag(db *gorm.DB, tag Tag) error {
	asn := db.Model(u).Association("Tags")
	if err := asn.Error; err != nil {
		return err
	}

	if !IsValidString(tag.Name) {
		// TODO: output message via logger
		return errors.Errorf("Invalid tag name: %s", tag.Name)
	}
	if db.Where("name = ? and user_id = ?", tag.Name, u.ID).Find(&Tag{}).RecordNotFound() {
		if err := asn.Append(tag).Error; err != nil {
			return err
		}
	}

	return nil
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

func IsValidJSON(s string) bool {
	var js interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}

func (t *Tag) AddData(db *gorm.DB, data Data) error {
	asn := db.Model(t).Association("SomeData")
	if err := asn.Error; err != nil {
		return err
	}
	if !IsValidJSON(data.Serialized) {
		// TODO: output message via logger
		fmt.Println("Invalid:", data.Serialized)
		return errors.Errorf("Invalid json format: %s", data.Serialized)
	}
	if err := asn.Append(data).Error; err != nil {
		return err
	}
	return nil
}
