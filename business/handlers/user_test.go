package handlers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/santiagoh1997/service-template/business/auth"
	"github.com/santiagoh1997/service-template/business/handlers"
	"github.com/santiagoh1997/service-template/business/service"
	"github.com/santiagoh1997/service-template/business/tests"
	"github.com/santiagoh1997/service-template/foundation/web"
)

// UserTests holds methods for each user subtest. This type allows passing
// dependencies for tests while still providing a convenient syntax when
// subtests are registered.
type UserTests struct {
	app        http.Handler
	kid        string
	userToken  string
	adminToken string
}

func TestGetUsers(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	ut := UserTests{
		app:        handlers.NewHTTPHandler("test", shutdown, test.Log, test.Auth, test.DB),
		kid:        test.KID,
		userToken:  test.Token("user@example.com", "password"),
		adminToken: test.Token("admin@example.com", "password"),
	}

	t.Run("Invalid page format", func(tt *testing.T) {
		page := "abc"
		r := httptest.NewRequest(http.MethodGet, "/v1/users/"+page+"/1", nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("\t%s\tShould receive a status code of 400 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

		got := w.Body.String()
		want := fmt.Sprintf(`{"error":"invalid page format: %s"}`, page)
		if got != want {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Invalid rows format", func(tt *testing.T) {
		rows := "abc"
		r := httptest.NewRequest(http.MethodGet, "/v1/users/1/"+rows, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("\t%s\tShould receive a status code of 400 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

		got := w.Body.String()
		want := fmt.Sprintf(`{"error":"invalid rows format: %s"}`, rows)
		if got != want {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Not authenticated", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/1/1", nil)
		w := httptest.NewRecorder()

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("\t%s\tShould receive a status code of 401 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 401 for the response.", tests.Success)

		got := w.Body.String()
		want := `{"error":"expected authorization header format: bearer \u003ctoken\u003e"}`
		if got != want {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Forbidden", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/1/1", nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Fatalf("\t%s\tShould receive a status code of 403 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 403 for the response.", tests.Success)

		got := w.Body.String()
		want := "you are not authorized for that action"
		if !strings.Contains(got, want) {
			if got != want {
				t.Logf("\t\tGot : %v", got)
				t.Logf("\t\tWant: %v", want)
				t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
			}
			t.Logf("\t%s\tShould get the expected result.", tests.Success)
		}
	})

	t.Run("No results", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/1000/1", nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("\t%s\tShould receive a status code of 200 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

		var u []service.User
		if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to unmarshal the response.", tests.Success)

		if len(u) != 0 {
			t.Fatalf("\t%s\tShould return 0 Users. Returned %d Users", tests.Failed, len(u))
		}
		t.Logf("\t%s\tShould return 0 Users.", tests.Success)
	})

	t.Run("Success case", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/1/1", nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("\t%s\tShould receive a status code of 200 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

		var u []service.User
		if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to unmarshal the response.", tests.Success)

		if len(u) == 0 {
			t.Fatalf("\t%s\tShould return 1 or more Users. Returned %d Users", tests.Failed, len(u))
		}
		t.Logf("\t%s\tShould return 1 or more Users.", tests.Success)
	})
}

func TestGetUser(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	ut := UserTests{
		app:        handlers.NewHTTPHandler("test", shutdown, test.Log, test.Auth, test.DB),
		kid:        test.KID,
		userToken:  test.Token("user@example.com", "password"),
		adminToken: test.Token("admin@example.com", "password"),
	}

	t.Run("Invalid userID", func(tt *testing.T) {
		id := "12345"
		r := httptest.NewRequest(http.MethodGet, "/v1/users/"+id, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("\t%s\tShould receive a status code of 400 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

		got := w.Body.String()
		want := fmt.Sprintf(`{"error":"%s"}`, service.ErrInvalidID)
		if got != want {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Not authenticated", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("\t%s\tShould receive a status code of 401 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 401 for the response.", tests.Success)

		got := w.Body.String()
		want := `{"error":"expected authorization header format: bearer \u003ctoken\u003e"}`
		if got != want {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("User not found", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("\t%s\tShould receive a status code of 404 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

		got := w.Body.String()
		want := fmt.Sprintf(`{"error":"%s"}`, service.ErrNotFound)
		if got != want {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Not authorized", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/"+tests.AdminID, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Fatalf("\t%s\tShould receive a status code of 403 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 403 for the response.", tests.Success)

		got := w.Body.String()
		want := fmt.Sprintf(`{"error":"%s"}`, service.ErrForbidden)
		if got != want {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Authorized (Admin requesting user)", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/"+tests.UserID, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("\t%s\tShould receive a status code of 200 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

		var u service.User
		if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to unmarshal the response.", tests.Success)
	})

	t.Run("Authorized (user requesting themself)", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/v1/users/"+tests.UserID, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("\t%s\tShould receive a status code of 200 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

		var u service.User
		if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to unmarshal the response.", tests.Success)
	})
}

func TestPostUser(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	ut := UserTests{
		app:        handlers.NewHTTPHandler("test", shutdown, test.Log, test.Auth, test.DB),
		kid:        test.KID,
		userToken:  test.Token("user@example.com", "password"),
		adminToken: test.Token("admin@example.com", "password"),
	}

	t.Run("Bad request", func(tt *testing.T) {
		nur := service.NewUserRequest{}
		body, err := json.Marshal(&nur)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("\t%s\tShould receive a status code of 400 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

		var got web.ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response to an error type : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to unmarshal the response to an error type.", tests.Success)

		want := web.ErrorResponse{
			Error: "field validation error",
			Fields: []web.FieldError{
				{Field: "name", Error: "name is a required field"},
				{Field: "last_name", Error: "last_name is a required field"},
				{Field: "country", Error: "country is a required field"},
				{Field: "email", Error: "email is a required field"},
				{Field: "roles", Error: "roles is a required field"},
				{Field: "password", Error: "password is a required field"},
			},
		}

		sorter := cmpopts.SortSlices(func(a, b web.FieldError) bool {
			return a.Field < b.Field
		})

		if diff := cmp.Diff(got, want, sorter); diff != "" {
			t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Bad request (duplicated email)", func(t *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			Email:           "email@email.com",
			LastName:        "Hernández",
			Country:         "Argentina",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}
		body, err := json.Marshal(&nur)
		if err != nil {
			t.Fatal(err)
		}

		// Create a user successfully...
		r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("\t%s\tShould receive a status code of 201 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 201 for the response.", tests.Success)

		// Attempt to create the same user (with the same email)
		r = httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
		w = httptest.NewRecorder()
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("\t%s\tShould receive a status code of 400 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)

		got := w.Body.String()
		want := fmt.Sprintf(`{"error":"%s"}`, service.ErrDuplicatedEmail)
		if got != want {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Success case", func(tt *testing.T) {
		nur := service.NewUserRequest{
			Name:            "Santiago",
			Email:           "santiago@santiago.com",
			LastName:        "Hernández",
			Country:         "Argentina",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}
		body, err := json.Marshal(&nur)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("\t%s\tShould receive a status code of 201 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 201 for the response.", tests.Success)

		var got service.User
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to unmarshal the response.", tests.Success)

		want := got
		want.Name = nur.Name
		want.LastName = nur.LastName
		want.Email = nur.Email
		want.Roles = nur.Roles

		if diff := cmp.Diff(got, want); diff != "" {
			t.Fatalf("\t%s\tShould get the expected result. Diff:\n%s", tests.Failed, diff)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})
}

func TestPutUser(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	ut := UserTests{
		app:        handlers.NewHTTPHandler("test", shutdown, test.Log, test.Auth, test.DB),
		kid:        test.KID,
		userToken:  test.Token("user@example.com", "password"),
		adminToken: test.Token("admin@example.com", "password"),
	}

	t.Run("Bad request (malformed userID)", func(tt *testing.T) {
		uur := service.UpdateUserRequest{
			Name:     "Another",
			LastName: "Name",
			Email:    "jorge@porcel.com",
			Country:  "Albania",
		}
		body, err := json.Marshal(&uur)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, "/v1/users/123", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("\t%s\tShould receive a status code of 400 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)
	})

	t.Run("Bad request (empty body)", func(tt *testing.T) {
		uur := service.UpdateUserRequest{}
		body, err := json.Marshal(&uur)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, "/v1/users/"+tests.AdminID, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("\t%s\tShould receive a status code of 400 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)
	})

	t.Run("Forbidden", func(tt *testing.T) {
		uur := service.UpdateUserRequest{
			Name:     "Another",
			LastName: "Name",
			Email:    "jorge@porcel.com",
			Country:  "Albania",
		}
		body, err := json.Marshal(&uur)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, "/v1/users/"+tests.AdminID, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Fatalf("\t%s\tShould receive a status code of 403 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 403 for the response.", tests.Success)
	})

	t.Run("User not found", func(tt *testing.T) {
		u := service.UpdateUserRequest{
			Name:     "Doesn't Exist",
			Email:    "email@email.com",
			LastName: "Last Name",
			Country:  "Argelia",
		}
		body, err := json.Marshal(&u)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, "/v1/users/"+uuid.New().String(), bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Fatalf("\t%s\tShould receive a status code of 404 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 404 for the response.", tests.Success)

		got := w.Body.String()
		want := service.ErrNotFound.Error()
		if !strings.Contains(got, want) {
			t.Logf("\t\tGot : %v", got)
			t.Logf("\t\tWant: %v", want)
			t.Fatalf("\t%s\tShould get the expected result.", tests.Failed)
		}
		t.Logf("\t%s\tShould get the expected result.", tests.Success)
	})

	t.Run("Success case", func(tt *testing.T) {
		uur := service.UpdateUserRequest{
			Name:     "Jorge",
			Email:    "jorge@porcel.com",
			LastName: "Porcel",
			Country:  "Albania",
		}
		body, err := json.Marshal(&uur)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, "/v1/users/"+tests.AdminID, bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("\t%s\tShould receive a status code of 200 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

		r = httptest.NewRequest(http.MethodGet, "/v1/users/"+tests.AdminID, nil)
		w = httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("\t%s\tShould receive a status code of 200 for the retrieve. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 200 for the retrieve.", tests.Success)

		var u service.User
		if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
		}

		if u.Name != "Jorge" {
			t.Fatalf("\t%s\tShould see an updated Name : got %q want %q", tests.Failed, u.Name, "Jorge")
		}
		t.Logf("\t%s\tShould see an updated Name.", tests.Success)

		if u.LastName != "Porcel" {
			t.Fatalf("\t%s\tShould see an updated LastName : got %q want %q", tests.Failed, u.LastName, "Porcel")
		}
		t.Logf("\t%s\tShould see an updated LastName.", tests.Success)

		if u.Country != "Albania" {
			t.Fatalf("\t%s\tShould see an updated Country : got %q want %q", tests.Failed, u.Country, "Porcel")
		}
		t.Logf("\t%s\tShould see an updated Country.", tests.Success)

		if u.Email != "jorge@porcel.com" {
			t.Fatalf("\t%s\tShould see an updated Email : got %q want %q", tests.Failed, u.Email, "jorge@porcel.com")
		}
		t.Logf("\t%s\tShould see an updated Email.", tests.Success)
	})
}

func TestDeleteUser(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	ut := UserTests{
		app:        handlers.NewHTTPHandler("test", shutdown, test.Log, test.Auth, test.DB),
		kid:        test.KID,
		userToken:  test.Token("user@example.com", "password"),
		adminToken: test.Token("admin@example.com", "password"),
	}

	t.Run("User not found", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+uuid.New().String(), nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Fatalf("\t%s\tShould receive a status code of 204 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 204 for the response.", tests.Success)
	})

	t.Run("Bad request", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/v1/users/123", nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("\t%s\tShould receive a status code of 400 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 400 for the response.", tests.Success)
	})

	t.Run("Not authorized", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+tests.UserID, nil)
		w := httptest.NewRecorder()

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("\t%s\tShould receive a status code of 401 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 401 for the response.", tests.Success)
	})

	t.Run("Forbidden", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+tests.AdminID, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Fatalf("\t%s\tShould receive a status code of 403 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 403 for the response.", tests.Success)
	})

	t.Run("Success case (admin deleting user)", func(tt *testing.T) {
		// Create User to be deleted...
		nur := service.NewUserRequest{
			Name:            "Delete",
			Email:           "delete@this.com",
			LastName:        "This",
			Country:         "Argentina",
			Roles:           []string{auth.RoleAdmin},
			Password:        "password",
			PasswordConfirm: "password",
		}
		body, err := json.Marshal(&nur)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPost, "/v1/users", bytes.NewBuffer(body))
		w := httptest.NewRecorder()

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("\t%s\tShould receive a status code of 201 for the response. Received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 201 for the response.", tests.Success)

		var u service.User
		if err := json.NewDecoder(w.Body).Decode(&u); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to unmarshal the response.", tests.Success)

		// Delete the created User...
		r = httptest.NewRequest(http.MethodDelete, "/v1/users/"+u.ID, nil)
		w = httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.adminToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Fatalf("\t%s\tShould receive a status code of 204 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 204 for the response.", tests.Success)
	})

	t.Run("Success case (user deleting themself)", func(tt *testing.T) {
		r := httptest.NewRequest(http.MethodDelete, "/v1/users/"+tests.UserID, nil)
		w := httptest.NewRecorder()

		r.Header.Set("Authorization", "Bearer "+ut.userToken)
		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Fatalf("\t%s\tShould receive a status code of 204 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 204 for the response.", tests.Success)
	})
}

func TestGetToken(t *testing.T) {
	test := tests.NewIntegration(t)
	t.Cleanup(test.Teardown)

	shutdown := make(chan os.Signal, 1)
	ut := UserTests{
		app:        handlers.NewHTTPHandler("test", shutdown, test.Log, test.Auth, test.DB),
		kid:        test.KID,
		userToken:  test.Token("user@example.com", "password"),
		adminToken: test.Token("admin@example.com", "password"),
	}

	t.Run("Not authorized", func(tt *testing.T) {
		lr := service.LoginRequest{
			Email:    "some_email@example.com",
			Password: "some_password",
		}
		body, err := json.Marshal(&lr)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/v1/users/token/"+ut.kid, bytes.NewBuffer(body))

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("\t%s\tShould receive a status code of 401 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 401 for the response.", tests.Success)
	})

	t.Run("Success case", func(tt *testing.T) {
		lr := service.LoginRequest{
			Email:    "admin@example.com",
			Password: "password",
		}
		body, err := json.Marshal(&lr)
		if err != nil {
			t.Fatal(err)
		}

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, "/v1/users/token/"+ut.kid, bytes.NewBuffer(body))

		ut.app.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("\t%s\tShould receive a status code of 200 for the response. Status code received: %v", tests.Failed, w.Code)
		}
		t.Logf("\t%s\tShould receive a status code of 200 for the response.", tests.Success)

		var got struct {
			Token string `json:"token"`
		}
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("\t%s\tShould be able to unmarshal the response : %v", tests.Failed, err)
		}
		t.Logf("\t%s\tShould be able to unmarshal the response.", tests.Success)
	})
}
