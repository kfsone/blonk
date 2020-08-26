package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/go-yaml/yaml"
	"github.com/google/uuid"
	"github.com/kfsone/blonk"
)

var optConfig = flag.String("config", ".blonk.rc", "Specify path to the blonk config file")
var optUUID = flag.String("uuid", "", "Specify the UUID to use")
var optEmail = flag.String("email", "", "Specify blink account email")
var optPassword = flag.String("password", "", "Specify blink account password")
var optPin = flag.String("pin", "", "Verification pin (if required)")

// Config describes current configuration.
type Config struct {
	UUID      uuid.UUID `yaml:"uuid"`
	Email     string    `yaml:"email,omitempty"`
	Password  string    `yaml:"password,omitempty"`
	AuthToken string    `yaml:"auth,omitempty"`
}

var config Config

func doRequest(into interface{}, request *blonk.Request, err error) error {
	if err != nil {
		return err
	}
	log.Print("request: ", request.URL)
	log.Print("body: ", string(request.Body))
	httpRequest, err := http.NewRequest(http.MethodPost, request.URL, bytes.NewReader(request.Body))
	for key, value := range request.Headers {
		httpRequest.Header.Set(key, value)
	}
	if err != nil {
		return err
	}
	response, err := http.DefaultClient.Do(httpRequest)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	log.Print(response.Status)

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}
	log.Print(string(body))

	return json.Unmarshal(body, into)
}

func login(session *blonk.Session, email, password string) error {
	loginReply := blonk.LoginReply{}
	request, err := session.NewLogin(email, password)
	err = doRequest(&loginReply, request, err)
	if err != nil {
		return err
	}
	session.Authed(loginReply.Account.ID, loginReply.Client.ID, loginReply.AuthToken.AuthToken)
	if loginReply.Account.Verify || loginReply.Client.Verify {
		if *optPin == "" {
			return errors.New("pin required for verification")
		}
		return verify(session)
	}
	log.Print("Logged in")
	return nil
}

func getPin() (string, error) {
	fmt.Print("Enter verification pin: ")
	var pin string
	_, err := fmt.Scanln(&pin)
	return pin, err
}

func verify(session *blonk.Session) error {
	log.Print("Verification code required")
	pin, err := getPin()
	if err != nil {
		return err
	}
	result := blonk.VerifyPinResult{}
	request, err := session.NewVerifyPin(pin)
	err = doRequest(&result, request, err)
	if err != nil {
		return err
	}
	if result.Valid == false {
		return fmt.Errorf("validation failed (%d): %s", result.Code, result.Message)
	}
	return nil
}

func readConfigFile() bool {
	if file, err := os.Open(*optConfig); err == nil {
		defer file.Close()
		data, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			log.Fatal(err)
		}
		return true
	}
	return false
}

func saveConfig() {
	if data, err := yaml.Marshal(config); err != nil {
		log.Fatal(err)
	} else {
		if err = ioutil.WriteFile(*optConfig, data, 0640); err != nil {
			log.Fatal(err)
		}
	}
}

func configure() bool {
	saved := false
	if config.UUID == uuid.Nil {
		if len(*optUUID) > 0 {
			if uuid, err := uuid.Parse(*optUUID); err != nil {
				panic(err)
			} else {
				config.UUID = uuid
			}
		} else {
			config.UUID = uuid.New()
		}
		saveConfig()
		saved = true
	}
	return saved
}

func either(flagValue *string, configValue string) string {
	if len(*flagValue) > 0 {
		return *flagValue
	}
	return configValue
}

func main() {
	flag.Parse()

	hadRc := readConfigFile()
	savedRc := configure()
	if !hadRc && !savedRc {
		saveConfig()
	}

	session, err := blonk.NewSession(blonk.DefaultHost, config.UUID)
	log.Print("Session UUID: ", session.UUID())
	if err != nil {
		panic(err)
	}
	defer session.Close()

	email := either(optEmail, config.Email)
	password := either(optPassword, config.Password)
	if err := login(session, email, password); err != nil {
		panic(err)
	}
}
