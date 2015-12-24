/*
 * This is an unpublished work copyright 2015 Jens-Uwe Mager
 * 30177 Hannover, Germany, jum@anubis.han.de
 */

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"unicode"

	"gopkg.in/gomail.v2"
)

type Config struct {
	EMailAddr  string
	SMTPServer string
	UserName   string
	PassWord   string
	FaxSecret  string
}

const (
	ConfigFile = ".dusfax"
)

var (
	conf      Config
	faxNumber = flag.String("n", "", "fax number to use (required)")
)

func init() {
	flag.StringVar(&conf.EMailAddr, "email", "", "the email address to send the email from")
	flag.StringVar(&conf.SMTPServer, "smtp", "", "the smtp server to use")
	flag.StringVar(&conf.UserName, "user", "", "user name for smtp auth")
	flag.StringVar(&conf.PassWord, "pass", "", "password for smtp auth")
	flag.StringVar(&conf.FaxSecret, "secret", "", "DUS.net fax secret")
}

func main() {
	user, err := user.Current()
	if err != nil {
		panic(err.Error())
	}
	configPath := filepath.Join(user.HomeDir, ConfigFile)
	fmt.Printf("configPath = %v\n", configPath)
	f, err := os.Open(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", configPath, err)
	} else {
		dec := json.NewDecoder(f)
		err = dec.Decode(&conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", configPath, err)
		}
		f.Close()
	}
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [flags] pdf_file_to_send\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}
	flag.Parse()
	//fmt.Printf("conf %#v\n", conf)
	f, err = os.OpenFile(configPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", configPath, err)
	} else {
		enc := json.NewEncoder(f)
		err := enc.Encode(&conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", configPath, err)
		}
		f.Close()
	}
	if flag.NArg() != 1 {
		flag.Usage()
	}
	if len(*faxNumber) == 0 {
		flag.Usage()
	}
	// massage the fax number to omit some special characters
	*faxNumber = strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		if r == '-' || r == '/' || r == '(' || r == ')' {
			return -1
		}
		return r
	}, *faxNumber)
	if (*faxNumber)[0] == '+' {
		*faxNumber = "00" + (*faxNumber)[1:]
	}
	d := gomail.NewPlainDialer(conf.SMTPServer, 587, conf.UserName, conf.PassWord)
	s, err := d.Dial()
	if err != nil {
		panic(err)
	}
	m := gomail.NewMessage()
	m.SetHeader("From", conf.EMailAddr)
	m.SetHeader("To", fmt.Sprintf("%s@fax.dus.net", *faxNumber))
	m.SetHeader("Subject", conf.FaxSecret)
	m.SetBody("text/plain", "fax send via dusfax.go\n")
	m.Attach(flag.Arg(0))
	if false {
		_, err = m.WriteTo(os.Stdout)
		if err != nil {
			panic(err)
		}
	}
	if err := gomail.Send(s, m); err != nil {
		panic(err)
	}
}
