package traq

import (
	"testing"
)

func Test_oauth(t *testing.T) {
	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	// url := TraQDefaultConfig.AuthCodeURL("random")
	// fmt.Printf("Visit the URL for the auth dialog: %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	// var code string
	// if _, err := fmt.Scan(&code); err != nil {
	// log.Fatal(err)
	// }
	// ctx := context.TODO()
	// tok, err := TraQDefaultConfig.Exchange(ctx, code)
	// if err != nil {
	// log.Fatal(err)
	// }

	// // sample
	// client := TraQDefaultConfig.Client(ctx, tok)
	// resp, _ := client.Get("https://q.trap.jp/api/v3/users/me")
	// defer resp.Body.Close()

	// fmt.Println(resp.Body)
}
