package main

import (
	"fmt"
	"net"
	"sort"
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

// Make map data pretty printable, alphabetically sorted and remove base key from from fromt of all keys
func parseMapToString(config *Config, values map[string]string) string {
	orderedMsg := ""
	var msg []string

	for k, v := range values {
		// remove BaseKeyToWrite
		k = strings.Replace(k, config.Etcd.BaseKeyToWrite+"/", "", 1)
		msg = append(msg, k+": "+v+"\r\n")
	}
	sort.Strings(msg)
	for _, v := range msg {
		orderedMsg += v
	}

	// Every time we run this we send to chan to display on screen, so lets reset the update timer here also
	lastUpdate = time.Now()

	return orderedMsg
}

// Run by main(), updates the text box until program exit
func refreshUpdateTime(updateTimeTextBox *walk.TextLabel) {
	for {
		updateTimeTextBox.SetText("Last update: " + fmt.Sprintf("%.0f", time.Since(lastUpdate).Seconds()))
		time.Sleep(500 * time.Millisecond) // just for human readability, dont refresh this too often
	}
}

// Run by main(), continuously prints read variables to screen except the ones we wrote
func readEtcdContinuously(config *Config, sendToMsgBoxCh chan map[string]string) {
	for {
		values, _ := myetcd.ReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToWrite)
		sendToMsgBoxCh <- values
	}
}

// Run by main(), waits for a response to the channel to update the message box until program exit
func mainLoop(config *Config, sendToMsgBoxCh chan map[string]string, resultMsgBox *walk.TextEdit) {
	for {
		msg := parseMapToString(config, <-sendToMsgBoxCh)
		resultMsgBox.SetText(msg)

		// sleep until we haven't updated for more than the sleep duration
		for time.Since(lastUpdate).Seconds() < float64(config.Etcd.SleepSeconds) {
			time.Sleep(1 * time.Second)
		}
	}
}
