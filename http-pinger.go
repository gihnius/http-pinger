package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/smtp"
	"os"
	"strings"
	"time"
)

var (
	c      = make(chan string)
	Config ConfigJson
)

type ConfigJson struct {
	Lag          int      `json:"lag"`
	Interval     int      `json:"interval"`
	UrlsFile     string   `json:"urls_file"`
	SmtpUsername string   `json:"smtp_username"`
	SmtpPassword string   `json:"smtp_password"`
	SmtpHost     string   `json:"smtp_host"`
	SmtpPort     string   `json:"smtp_port"`
	EmailSubject string   `json:"email_subject"`
	FromEmail    string   `json:"from_email"`
	ToEmails     []string `json:"to_emails"`
}

type Msg struct {
	Server    string `json:"server_url"`
	Type      string `json:"message_type"`
	Date      string `json:"date"`
	Lag       int    `json:"lag_threshold_in_second"`
	Interval  int    `json:"check_interval_in_second"`
	Status    int    `json:"http_status_code"`
	Responsed string `json:"server_responsed_in"`
	Error     string `json:"error"`
}

func main() {
	ParseConfig()
	file_data, err := ioutil.ReadFile(Config.UrlsFile)
	if err != nil {
		panic(err)
	}
	var urls []string
	for _, line := range strings.Split(string(file_data), "\n") {
		url := strings.TrimSpace(line)
		if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
			urls = append(urls, url)
		}
	}

	for _, url := range urls {
		fmt.Println(url)
		go ping(url)
	}

	for {
		select {
		case msg := <-c:
			fmt.Println(msg)
		}
	}
}

func ping(url string) {
	for {
		start := time.Now()
		res, err := http.Get(url)
		msg := Msg{}
		if err != nil {
			msg.Server = url
			msg.Type = "Fatal"
			msg.Date = time.Now().Format("2006-01-02 15:04 -0700")
			msg.Lag = Config.Lag
			msg.Interval = Config.Interval
			msg.Status = -1
			msg.Responsed = "NaN"
			msg.Error = err.Error()
			b, _ := json.Marshal(msg)
			str := string(b)
			EmailMsg(str)
			c <- str
		} else {
			lag := time.Since(start)
			msg.Server = url
			msg.Date = time.Now().Format("2006-01-02 15:04 -0700")
			msg.Lag = Config.Lag
			msg.Interval = Config.Interval
			msg.Status = res.StatusCode
			msg.Responsed = lag.String()
			msg.Type = "OK" // expected
			if res.StatusCode != 200 {
				msg.Type = "Warning"
				msg.Error = "Unexpected http status code!"
				b, _ := json.Marshal(msg)
				str := string(b)
				EmailMsg(str)
			}
			if lag > time.Duration(Config.Lag)*time.Second {
				msg.Type = "Warning"
				msg.Error = "Responsed times over lag threshold!"
				b, _ := json.Marshal(msg)
				str := string(b)
				EmailMsg(str)
			}
			b, _ := json.Marshal(msg)
			str := string(b)
			c <- str
			res.Body.Close()
		}
		time.Sleep(time.Duration(Config.Interval) * time.Second)
	}
}

func EmailMsg(msg string) {
	sendMail(Config.EmailSubject, msg, Config.FromEmail, Config.ToEmails)
}

func sendMail(subject string, message string, from string, to []string) {
	auth := smtp.PlainAuth(
		"",
		Config.SmtpUsername,
		Config.SmtpPassword,
		Config.SmtpHost,
	)
	msg := fmt.Sprintf("To: %s\r\nFrom: %s\r\nSubject: %s\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s", strings.Join(to, ";"), from, subject, message)
	err := smtp.SendMail(fmt.Sprintf("%s:%s", Config.SmtpHost, Config.SmtpPort), auth, from, to, []byte(msg))
	if err != nil {
		fmt.Println("[Warning] Send Email failed: ", err.Error())
		return
	}
	fmt.Println("Sent Email Notification to: ", to)
}

func ParseConfig() {
	conf := os.Getenv("CONFIG")
	if conf == "" {
		conf = "config.json"
	}
	file, err := os.Open(conf)
	if err != nil {
		fmt.Println("Read config.json failed: ", err.Error())
		os.Exit(1)
	}
	defer file.Close()
	j := json.NewDecoder(file)
	err = j.Decode(&Config)
	if err != nil {
		fmt.Println("Parse config.json failed: ", err.Error())
		os.Exit(1)
	}
}
