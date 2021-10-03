package traq

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"testing"
	"time"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/traQ/model"
	traQ "github.com/traPtitech/traQ/router/v3"
	"github.com/traPtitech/traQ/utils/random"
	"golang.org/x/oauth2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const URL = "http://localhost:3000/api/v3"

func TestMain(m *testing.M) {
	authentication := struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}{
		Name:     "traq",
		Password: "traq",
	}
	body, err := json.Marshal(authentication)
	if err != nil {
		panic(err)
	}

	// traQのサーバーが起動するのを5*3秒間待つ
	cnt := 0
WaitServer:
	req, err := http.NewRequest(http.MethodPost, URL+"/login", bytes.NewBuffer(body))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	jar, _ := cookiejar.New(nil)
	client := &http.Client{
		Jar: jar,
	}
	res, err := client.Do(req)
	if err != nil || res.StatusCode >= 300 {
		if cnt > 2 {
			panic("unexpected status code")
		}
		cnt++
		time.Sleep(5 * time.Second)
		goto WaitServer

	}

	res, err = client.Get(URL + "/users/me")
	if err != nil {
		panic(err)
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}

	user := new(traQ.User)
	err = json.Unmarshal(data, &user)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		DSN:                       fmt.Sprintf("%s:%s@tcp(%s:3306)/%s?charset=utf8mb4&parseTime=true&loc=Local", "root", "password", "localhost", "traq"),
		DefaultStringSize:         256,   // default size for string fields
		DisableDatetimePrecision:  true,  // disable datetime precision, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // `change` when rename column, rename column not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // auto configure based on currently MySQL version
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	token := random.SecureAlphaNumeric(36)
	scopes := model.AccessScopes{}
	scopes.Add("read")
	newToken := &model.OAuth2Token{
		ID:           uuid.Must(uuid.NewV4()),
		UserID:       user.ID,
		AccessToken:  token,
		RefreshToken: random.SecureAlphaNumeric(36),
		CreatedAt:    time.Now(),
		ExpiresIn:    1000,
		Scopes:       scopes,
	}
	err = db.Create(newToken).Error
	if err != nil {
		panic(err)
	}

	repo := &TraQRepository{
		Config: TraQDefaultConfig,
		URL:    URL,
	}
	u, err := repo.GetUserMe(&oauth2.Token{
		AccessToken: token,
		Expiry:      time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(*u)
}
