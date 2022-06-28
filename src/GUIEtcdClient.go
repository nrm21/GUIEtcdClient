package main

import (
	"fmt"
	"os"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var version string // to be auto-added with -ldflags at build time
var mw *walk.MainWindow
var clientID, exePath string

// Program entry point
func main() {
	var modifyValueBox, modifyKeyBox *walk.TextEdit
	var resultMsgBox *walk.TextEdit
	var chatTextBox *walk.LineEdit

	// Get CWD and use it to find if we are in ./src or base of project, then normalize it
	// by removing '/src' from end of path so we can find where our support files are located
	exePath, _ = os.Getwd()
	if exePath[len(exePath)-4:] == "\\src" {
		exePath = exePath[:len(exePath)-4]
	}

	config, err := getConfigContentsFromYaml(exePath + "/support/config.yml")
	if err != nil {
		walk.MsgBox(nil, "Fatal Error", "Fatal: "+err.Error(), walk.MsgBoxIconError)
		os.Exit(1)
	}

	// If the config file doesn't exist this will run from config file
	// if os.IsNotExist(err) {
	// 	walk.MsgBox(mw, "Error", "The proper config file or registry keys do not exist", walk.MsgBoxIconError)
	// }

	// if localhost is open use that endpoint instead
	if testSockConnect("127.0.0.1", "2379") {
		config.Etcd.Endpoints = []string{"127.0.0.1:2379"}
		println("Localhost open using localhost socket instead")
	} else {
		println("Localhost NOT open using config endpoints list")
	}

	MainWindow{
		AssignTo: &mw,
		Title:    "Etcd Client",
		Size:     Size{1024, 768},
		Layout:   VBox{},
		Children: []Widget{
			HSplitter{
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							TextLabel{
								Text: "Key:",
							},
							TextEdit{
								AssignTo: &modifyKeyBox,
							},
							TextLabel{
								Text: "Value:",
							},
							TextEdit{
								AssignTo: &modifyValueBox,
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Modify",
								OnClicked: func() {
									go func() {
										WriteToEtcd(config, config.Etcd.BaseKeyToWrite+"/"+modifyKeyBox.Text(), modifyValueBox.Text())
										modifyValueBox.SetText("")
									}()
								},
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Delete",
								OnClicked: func() {
									go func() {
										DeleteFromEtcd(config, config.Etcd.BaseKeyToWrite+"/"+modifyKeyBox.Text())
										modifyKeyBox.SetText("")
										modifyValueBox.SetText("")
									}()
								},
							},
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					TextEdit{
						AssignTo: &resultMsgBox,
						ReadOnly: true,
						MinSize:  Size{600, 630},
						Font:     Font{Family: "Ariel", PointSize: 12},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							LineEdit{
								AssignTo: &chatTextBox, Text: "",
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Send",
								OnClicked: func() {
									go func() {
										valueToWrite := fmt.Sprintf("%s", chatTextBox.Text()+"\r\n")
										WriteToEtcd(config, config.Etcd.BaseKeyToWrite, valueToWrite)
										chatTextBox.SetText("")
									}()
								},
							},
							TextLabel{
								Text: "Version: " + version,
							},
						},
					},
				},
			},
		},
	}.Create()

	// Start listening for a response
	go listenForResponse(&config, resultMsgBox, config.Etcd.BaseKeyToWrite)

	mw.Run()
}
