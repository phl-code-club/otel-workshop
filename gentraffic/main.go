package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-faker/faker/v4"
)

var (
	authURL    = os.Getenv("AUTH_URL")
	profileURL = os.Getenv("PROFILE_URL")
)

var userCount = func() int {
	countVar, exists := os.LookupEnv("USER_COUNT")
	if !exists {
		return 100
	}
	count, err := strconv.Atoi(countVar)
	if err != nil {
		panic("Invalid USER_COUNT " + countVar)
	}
	return count
}()

type TokenResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type User struct {
	Email     string    `json:"email" faker:"email"`
	Password  string    `json:"password" faker:"loginPassword"`
	CreatedAd time.Time `json:"created_at"`
}

type Profile struct {
	ID       int    `json:"id"`
	Username string `json:"username" faker:"username"`
	Bio      string `json:"bio" faker:"paragraph"`
	Location string `json:"location" faker:"addressString"`
}

func main() {
	_ = faker.AddProvider("loginPassword", func(_v reflect.Value) (any, error) {
		pass := faker.Password()
		return fmt.Sprintf("%s%d", pass, rand.IntN(10)), nil
	})
	_ = faker.AddProvider("addressString", func(_v reflect.Value) (any, error) {
		addr := faker.GetRealAddress()
		return fmt.Sprintf("%s, %s, %s, %s", addr.Address, addr.City, addr.State, addr.PostalCode), nil
	})
	var wg sync.WaitGroup
	for i := range userCount {
		var user User
		err := faker.FakeData(&user)
		wg.Go(func() {
			err = testSuite(user)
			if err != nil {
				slog.Error("error running test suite", "error", err, "iteration", i, "email", user.Email)
			}
		})
	}

	wg.Wait()
}

func timer() func() time.Duration {
	start := time.Now()
	return func() time.Duration {
		return time.Since(start)
	}
}

func testSuite(user User) error {
	t := timer()
	requestBody := fmt.Sprintf(`{"email": "%s", "password": "%s"}`, user.Email, user.Password)
	signupResponse, err := http.Post(authURL+"/signup", "application/json", strings.NewReader(requestBody))
	if err != nil {
		return err
	}
	if signupResponse.StatusCode > 300 {
		return errors.New("invalid signup request: " + requestBody)
	}

	signupDecoder := json.NewDecoder(signupResponse.Body)
	var signupBody TokenResponse
	err = signupDecoder.Decode(&signupBody)
	if err != nil {
		return err
	}
	userRequest, err := http.NewRequest(http.MethodGet, authURL+"/user", nil)
	if err != nil {
		return err
	}
	userRequest.Header.Set("authorization", "bearer "+signupBody.Token)
	client := http.Client{}
	userResponse, err := client.Do(userRequest)
	if err != nil {
		return err
	}
	if userResponse.StatusCode > 300 {
		return errors.New("invalid user request: " + signupBody.Token)
	}
	userDecoder := json.NewDecoder(userResponse.Body)
	var userBody User
	err = userDecoder.Decode(&userBody)
	if err != nil {
		return err
	}
	signinResponse, err := http.Post(authURL+"/signin", "application/json", strings.NewReader(requestBody))
	if err != nil {
		return err
	}
	if signinResponse.StatusCode > 300 {
		return errors.New("invalid signin request: " + requestBody)
	}

	signinDecoder := json.NewDecoder(signinResponse.Body)
	var signinBody TokenResponse
	err = signinDecoder.Decode(&signinBody)
	if err != nil {
		return err
	}

	badProfileRequest, err := http.NewRequest(http.MethodGet, profileURL, nil)
	if err != nil {
		return err
	}
	badProfileRequest.Header.Set("authorization", "bearer "+signinBody.Token)
	badProfileResponse, err := client.Do(badProfileRequest)
	if err != nil {
		return err
	}
	if badProfileResponse.StatusCode < 400 {
		return errors.New("incorrect status code on bad profile request")
	}
	var profile Profile
	err = faker.FakeData(&profile)
	if err != nil {
		return err
	}
	profileBody := fmt.Sprintf(`{"username": "%s", "bio": "%s", "location": "%s"}`, profile.Username, profile.Bio, profile.Location)
	createProfileRequest, err := http.NewRequest(http.MethodPut, profileURL, strings.NewReader(profileBody))
	if err != nil {
		return err
	}
	createProfileRequest.Header.Set("authorization", "bearer "+signinBody.Token)
	createProfileResponse, err := client.Do(createProfileRequest)
	if err != nil {
		return err
	}
	if createProfileResponse.StatusCode != 200 {
		return errors.New("incorrect status code on profile create request")
	}
	profileRequest, err := http.NewRequest(http.MethodGet, profileURL, nil)
	if err != nil {
		return err
	}
	profileRequest.Header.Set("authorization", "bearer "+signinBody.Token)
	profileResponse, err := client.Do(profileRequest)
	if err != nil {
		return err
	}
	if profileResponse.StatusCode != 200 {
		return errors.New("incorrect status code on profile request")
	}
	slog.Info("finished running test suite", "user", user.Email, "duration", t())
	return nil
}
