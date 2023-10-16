package traq

import (
	//"context"
	"context"
	// "encoding/json"
	"fmt"
	// "net/http"
	// "net/url"
	// "strconv"

	"github.com/gofrs/uuid"
	"github.com/traPtitech/go-traq"
	"golang.org/x/oauth2"
)



func (repo *TraQRepository) GetUser(token *oauth2.Token, userID uuid.UUID) (*traq.User, error) {
	// URL := fmt.Sprintf("%s/users/%s", repo.URL, userID)
	// req, err := http.NewRequest(http.MethodGet, URL, nil)
	// if err != nil {
	// 	return nil, err
	// }
	// data, err := repo.doRequest(token, req)
	// if err != nil {
	// 	return nil, err
	// }

	// user := new(traq.User)
	// err = json.Unmarshal(data, &user)
	
	// return user, err

	// ここから go-traq 書き換え

	traqconf := traq.NewConfiguration()
	conf := TraQDefaultConfig
	traqconf.HTTPClient = conf.Client(context.Background(),token)
	apiClient := traq.NewAPIClient(traqconf)
	userDtail,_,err := apiClient.UserApi.GetUser(context.Background(),userID.String()).Execute()
	if err!=nil{
		fmt.Println("GetUserError")
		return nil,err
	}
	user := convertUserdetailToUser(userDtail)
	fmt.Println("GetUserDetailSuccess")
	return user, err

	// ここまで go-traq 書き換え
}

func (repo *TraQRepository) GetUsers(token *oauth2.Token, includeSuspended bool) ([]*traq.User, error) {
	// URL, err := url.Parse(fmt.Sprintf("%s/users", repo.URL))
	// if err != nil {
	// 	return nil, err
	// }
	// q := URL.Query()
	// q.Set("include-suspended", strconv.FormatBool(includeSuspended))
	// URL.RawQuery = q.Encode()
	// req, err := http.NewRequest(http.MethodGet, URL.String(), nil)
	// if err != nil {
	// 	return nil, err
	// }
	// data, err := repo.doRequest(token, req)
	// if err != nil {
	// 	return nil, err
	// }

	// users := make([]*traq.User, 0)
	// err = json.Unmarshal(data, &users)
	
	// return users, err
	
	traqconf := traq.NewConfiguration()
	conf := TraQDefaultConfig
	traqconf.HTTPClient = conf.Client(context.Background(),token)
	apiClient := traq.NewAPIClient(traqconf)
	users,_,err := apiClient.UserApi.GetUsers(context.Background()).IncludeSuspended(includeSuspended).Execute()
	if err!=nil{
		fmt.Println("GetUsersError")
		return nil,err
	}
	res_users := convertUsersToUsers(users)
	fmt.Println("GetUsersSuccess")
	fmt.Println(res_users)
	return res_users, err
}


func (repo *TraQRepository) GetUserMe(token *oauth2.Token) (*traq.User, error) {
	// URL := fmt.Sprintf("%s/users/me", repo.URL)
	// req, err := http.NewRequest(http.MethodGet, URL, nil)
	// if err != nil {
	// 	return nil, err
	// }
	// data, err := repo.doRequest(token, req)
	// if err != nil {
	// 	return nil, err
	// }

	// user := new(traq.User)
	// err = json.Unmarshal(data, &user)
	// return user, err
	
	// ここから go-traq 書き換え
	traqconf := traq.NewConfiguration()
	conf :=TraQDefaultConfig
	traqconf.HTTPClient=conf.Client(context.Background(),token)
	client := traq.NewAPIClient(traqconf)

	data,_,err:= client.MeApi.GetMe(context.Background()).Execute()
	if err != nil{
		return nil,err
	}
	user := convertMyUserdetailToUser(data)
	fmt.Println("GetUserMeSuccess")
	return user, err
	// ここまで　go-traq 書き換え
}

func convertMyUserdetailToUser (userdetail *traq.MyUserDetail) *traq.User{
	user := new(traq.User)
	user.Id=userdetail.Id
	user.Name=userdetail.Name
	user.DisplayName=userdetail.DisplayName
	user.IconFileId=userdetail.IconFileId
	user.Bot=userdetail.Bot
	user.State=userdetail.State
	user.UpdatedAt=userdetail.UpdatedAt
	return user
}

func convertUserdetailToUser (userdetail *traq.UserDetail) *traq.User{
	user := new(traq.User)
	user.Id=userdetail.Id
	user.Name=userdetail.Name
	user.DisplayName=userdetail.DisplayName
	user.IconFileId=userdetail.IconFileId
	user.Bot=userdetail.Bot
	user.State=userdetail.State
	user.UpdatedAt=userdetail.UpdatedAt
	return user
}
func convertUsersToUsers(users []traq.User) []*traq.User{
	new_users:=make([]*traq.User,len(users))
	for i,_user := range users{
		user:=_user
		new_users[i] = &user
	}
	return new_users
}