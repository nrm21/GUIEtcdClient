package main

import (
	"net"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/nrm21/EtcdChat/src/myetcd"
	"github.com/nrm21/support"
	"gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Etcd struct {
		// var name has to be uppercase here or it won't work
		Endpoints      []string      `yaml:"endpoints"`
		BaseKeyToWrite string        `yaml:"baseKeyToWrite"`
		Timeout        int           `yaml:"timeout"`
		SleepSeconds   time.Duration `yaml:"sleepSeconds"`
		CertPath       string        `yaml:"certpath"`
	}
}

// Unmarshals the config contents from file into memory
func getConfigContentsFromYaml(filename string) (Config, error) {
	var conf Config
	file, err := support.ReadConfigFileContents(filename)
	if err != nil {
		return conf, err
	}
	err = yaml.Unmarshal(file, &conf)
	if err != nil {
		return conf, err
	}

	return conf, nil
}

// Returns a string of (up to) the nanosecond level of right now (at runtime)
func getMilliTime() string {
	now := time.Now()
	tstamp := now.Format(time.RFC3339Nano)
	tstamp = strings.Replace(tstamp, "T", "  ", 1)

	return tstamp[:len(tstamp)-14] // second resolution
}

// Checks a socket connection and returns bool of if open or not
func testSockConnect(host string, port string) bool {
	conn, _ := net.DialTimeout("tcp", net.JoinHostPort(host, port), 500*time.Millisecond)
	if conn != nil {
		defer conn.Close()

		return true
	} else {
		return false
	}
}

// If the first character in a keyname is a '/' we remove it.  This should provite consistancy
// for us in modifying and deleting subkeys of the base key.
func normalizeKeyNames(value string) string {
	if value[:1] == "/" {
		value = value[1:len(value)]
	}

	return value
}

// Make map data pretty preintable and remove base key from from fromt of all keys
func parseMapToString(config *Config, values map[string]string) string {
	msg := ""
	for k, v := range values {
		// remove BaseKeyToWrite
		k = strings.Replace(k, config.Etcd.BaseKeyToWrite+"/", "", 1)

		msg += k + ": " + v + "\r\n"
	}

	return msg
}

// Continuously prints read variables to screen except the ones we wrote
func readEtcdContinuously(sendToMsgBoxCh chan string, config *Config) {
	for {
		values, _ := myetcd.ReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToWrite)
		sendToMsgBoxCh <- parseMapToString(config, values)
	}
}

// Is run by main(), loops forever waiting for a response to the channel
func listenForResponse(config *Config, resultMsgBox *walk.TextEdit) {
	sendToMsgBoxCh := make(chan string)

	// This needs its own thread since it also loops forever
	go readEtcdContinuously(sendToMsgBoxCh, config)

	for { // loop forever (user expected to break)
		msg := <-sendToMsgBoxCh
		// Append to the end of the message thats already there
		// msg = resultMsgBox.Text() + msg
		resultMsgBox.SetText(msg)
		time.Sleep(config.Etcd.SleepSeconds * time.Second)
	}
}
