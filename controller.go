package vegeta

import (
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
	"github.com/labstack/echo"
)

type Controller struct {
	*Vegeta
	DB *gorm.DB
}

type User struct {
}

type Data struct {
	gorm.Model
	IPAddress  string
	Name       string
	Serialized string
}

func NewController(v *Vegeta) (*Controller, error) {
	db, err := gorm.Open("mysql", "root:DHFLSHQ3@/vegeta?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		return nil, err
	}
	return &Controller{Vegeta: v, DB: db}, nil
}

func (c *Controller) Close() error {
	return c.DB.Close()
}

func (c *Controller) Index() echo.HandlerFunc {
	return func(c echo.Context) error {
		arg := c.Param("arg")
		return c.String(http.StatusOK, arg)
	}
}
