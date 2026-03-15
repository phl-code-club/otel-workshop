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

func main() {
	_ = faker.AddProvider("loginPassword", func(_v reflect.Value) (any, error) {
		pass := faker.Password()
		return fmt.Sprintf("%s%d", pass, rand.IntN(10)), nil
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
	signupResponse, err := http.Post("http://localhost:8000/signup", "application/json", strings.NewReader(requestBody))
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
	userRequest, err := http.NewRequest(http.MethodGet, "http://localhost:8000/user", nil)
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
	signinResponse, err := http.Post("http://localhost:8000/signin", "application/json", strings.NewReader(requestBody))
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
	slog.Info("finished running test suite", "user", user.Email, "duration", t())
	return nil
}
