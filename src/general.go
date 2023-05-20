package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/lxn/walk"
	"github.com/nrm21/support"
	"gopkg.in/yaml.v2"
)

// Config struct
type Config struct {
	Etcd struct {
		// var name has to be uppercase here or it won't work
		Endpoints    []string `yaml:"endpoints"`
		BaseKeyToUse string   `yaml:"baseKeyToUse"`
		Timeout      int      `yaml:"timeout"`
		CertPath     string   `yaml:"certpath"`
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

// Runs when we click either the export or import buttons at the bottom of GUI
func dbImportExport(config *Config, filename, mode string) {
	if mode == "import" {
		if filename == "" {
			walk.MsgBox(nil, "Error", "Please put in a filename", walk.MsgBoxIconError)
		}
		// read the bytes from file
		filebytes, err := os.ReadFile(filename)
		if err != nil {
			fmt.Println(err)
		}
		// and unmarshal the values from JSON
		values := make(map[string]string)
		err = json.Unmarshal(filebytes, &values)
		if err != nil {
			fmt.Println(err)
		}
		// and write them to Etcd
		for key, value := range values {
			support.WriteToEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, key, value)
		}
	} else if mode == "export" {
		path, _ := os.Getwd()
		filename = path + "\\backup.json"

		// read values from Etcd and marshal them into JSON
		values, _ := support.ReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToUse)

		// and convert bytes to string in new map before exporting
		stringifiedValues := make(map[string]string)
		for key, val := range values {
			stringifiedValues[key] = string(val)
		}
		filebytes, err := json.MarshalIndent(stringifiedValues, "", "   ")

		if err != nil {
			fmt.Println(err)
		} else { // and write it to file
			err = os.WriteFile(filename, filebytes, 0644)
			if err != nil {
				fmt.Println(err)
			}
			walk.MsgBox(nil, "Info", "Backup file created", walk.MsgBoxIconInformation)
		}
	}
}

// Make map data pretty printable, alphabetically sorted and remove base key from from fromt of all keys
func parseMapToString(config *Config, values map[string][]byte) string {
	orderedMsg := ""
	var msg []string

	for k, v := range values {
		// remove BaseKeyToUse
		k = strings.Replace(k, config.Etcd.BaseKeyToUse+"/", "", 1)
		// trim null bytes before sending to output (allows safe printing of message)
		v = bytes.ReplaceAll(v, []byte("\x00"), []byte(" "))

		msg = append(msg, k+": "+string(v)+"\r\n")
	}
	sort.Strings(msg)
	for _, v := range msg {
		orderedMsg += v
	}

	return orderedMsg
}

// This function is called after the watcher chan returns with changes and
// compares the changes to what exists and modifies only the needed ones
// then finally it updates the messagebox
func updateWatchedChanges() {
	for {
		newValues := <-watchedChangeCh
		for key, value := range newValues {
			if string(dbValues[key]) != string(value) {
				dbValues[key] = value
			}
		}

		sendToMsgBoxCh <- dbValues
	}
}

// Anytime this is called it will read the current values from etcd for the
// given basekey, and send them to the channel.  It might need to be run async
// depending on where it's used in the codebase since it waits for info forever
// to send to the messagebox until program close.
func readValuesAndSendToMsgBox(config *Config) {
	var err error
	dbValues, err = support.ReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToUse)
	if err != nil {
		walk.MsgBox(nil, "Fatal Error", "Fatal: "+err.Error()+"\nPossible authentication failure", walk.MsgBoxIconError)
		log.Fatal(err.Error())
	}
	sendToMsgBoxCh <- dbValues
}

// Run by main(), waits for a response to the channel to update the message box until program exit
func mainLoop(config *Config, resultMsgBox *walk.TextEdit) {
	for {
		msg := parseMapToString(config, <-sendToMsgBoxCh) // will wait for sending channel
		resultMsgBox.SetText(msg)
	}
}
