package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dimashiro/service/app/services/retail-api/handlers"
	"github.com/dimashiro/service/business/data/tests"
)

type UserTests struct {
	app        http.Handler
	userToken  string
	adminToken string
}

func TestUsers(t *testing.T) {
	test := tests.NewIntegration(t, c, "inttestusers")
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	tests := UserTests{
		app: handlers.APIMux(handlers.APIMuxConfig{
			Shutdown: shutdown,
			Log:      test.Log,
			Auth:     test.Auth,
			DB:       test.DB,
		}),
		userToken:  test.Token("user@example.com", "gophers"),
		adminToken: test.Token("admin@example.com", "gophers"),
	}

	t.Run("getToken404", tests.getToken404)
	t.Run("getToken200", tests.getToken200)
}

func (ut *UserTests) getToken404(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("unknown@example.com", "some-password")
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to deny tokens to unknown users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching a token with an unrecognized email.", testID)
		{
			if w.Code != http.StatusNotFound {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 404 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 404 for the response.", tests.Success, testID)
		}
	}
}

func (ut *UserTests) getToken200(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/v1/users/token", nil)
	w := httptest.NewRecorder()

	r.SetBasicAuth("admin@example.com", "gophers")
	ut.app.ServeHTTP(w, r)

	t.Log("Given the need to issues tokens to known users.")
	{
		testID := 0
		t.Logf("\tTest %d:\tWhen fetching a token with valid credentials.", testID)
		{
			if w.Code != http.StatusOK {
				t.Fatalf("\t%s\tTest %d:\tShould receive a status code of 200 for the response : %v", tests.Failed, testID, w.Code)
			}
			t.Logf("\t%s\tTest %d:\tShould receive a status code of 200 for the response.", tests.Success, testID)

			var got struct {
				Token string `json:"token"`
			}
			if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
				t.Fatalf("\t%s\tTest %d:\tShould be able to unmarshal the response : %v", tests.Failed, testID, err)
			}
			t.Logf("\t%s\tTest %d:\tShould be able to unmarshal the response.", tests.Success, testID)

			// TODO(jlw) Should we ensure the token is valid?
		}
	}
}
