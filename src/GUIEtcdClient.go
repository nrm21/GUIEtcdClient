package main

import (
	"errors"
	"os"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/nrm21/EtcdChat/src/myetcd"
)

var version string // to be auto-added with -ldflags at build time
var lastUpdate time.Time
var mw *walk.MainWindow
var clientID, exePath string

// Program entry point
func main() {
	var modifyValueBox, modifyKeyBox, resultMsgBox *walk.TextEdit
	var updateTimeTextBox *walk.TextLabel
	var importExportFileBox *walk.LineEdit

	// Get CWD and use it to find if we are in ./src or base of project, then normalize it
	// by removing '/src' from end of path so we can find where our support files are located
	exePath, _ = os.Getwd()
	if exePath[len(exePath)-4:] == "\\src" || exePath[len(exePath)-4:] == "\\bin" {
		exePath = exePath[:len(exePath)-4]
	}

	config, err := getConfigContentsFromYaml(exePath + "\\config.yml")
	// if the config file doesnt exist
	if err != nil {
		walk.MsgBox(nil, "Fatal Error", "Fatal: "+err.Error(), walk.MsgBoxIconError)
		os.Exit(1)
	}

	// if the cert path doesnt exist
	if _, err := os.Stat(config.Etcd.CertPath); errors.Is(err, os.ErrNotExist) {
		walk.MsgBox(nil, "Fatal Error", "Fatal: "+err.Error(), walk.MsgBoxIconError)
	}

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
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Modify",
								OnClicked: func() {
									go func() {
										myetcd.WriteToEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToWrite+"/"+
											normalizeKeyNames(modifyKeyBox.Text()), modifyValueBox.Text())

										values, _ := myetcd.ReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToWrite)
										resultMsgBox.SetText(parseMapToString(&config, values))
										modifyKeyBox.SetText("")
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
										numDeleted := myetcd.DeleteFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToWrite+"/"+normalizeKeyNames(modifyKeyBox.Text()))
										if numDeleted < 1 {
											walk.MsgBox(nil, "Error", "No records found", walk.MsgBoxIconInformation)
										}
										values, _ := myetcd.ReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToWrite)
										resultMsgBox.SetText(parseMapToString(&config, values))
										modifyKeyBox.SetText("")
										modifyValueBox.SetText("")
									}()
								},
							},
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
								Text:    "Refresh",
								OnClicked: func() {
									go func() {
										values, _ := myetcd.ReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToWrite)
										resultMsgBox.SetText(parseMapToString(&config, values))
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
						Font: Font{
							Family:    "Ariel",
							PointSize: 15,
						},
					},
				},
			},
			HSplitter{
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							TextLabel{
								AssignTo: &updateTimeTextBox,
								Text:     "Last update: ",
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Export DB",
								OnClicked: func() {
									go func() {
										dbImportExport(&config, importExportFileBox.Text(), "export")
									}()
								},
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Import DB",
								OnClicked: func() {
									go func() {
										dbImportExport(&config, importExportFileBox.Text(), "import")
									}()
								},
							},
							TextLabel{
								Text: "Import/Export File: ",
							},
							LineEdit{
								AssignTo: &importExportFileBox,
								Text:     "",
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

	sendToMsgBoxCh := make(chan map[string]string)

	// These need their own thread since they also loop forever
	go refreshUpdateTime(updateTimeTextBox)
	go readEtcdContinuously(&config, sendToMsgBoxCh)
	go mainLoop(&config, sendToMsgBoxCh, resultMsgBox)

	mw.Run()
}
