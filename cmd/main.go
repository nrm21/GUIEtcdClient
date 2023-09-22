package main

import (
	"errors"
	"os"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/nrm21/support"
)

var version string // to be auto-added with -ldflags at build time
var mw *walk.MainWindow
var clientID, exePath string
var dbValues map[string][]byte
var sendToMsgBoxCh, watchedChangeCh chan map[string][]byte
var closeWatcher chan bool

// Program entry point
func main() {
	var modifyValueBox, modifyKeyBox, resultMsgBox, baseKeyToUseBox *walk.TextEdit
	var importExportDirBox *walk.LineEdit
	sendToMsgBoxCh = make(chan map[string][]byte)
	watchedChangeCh = make(chan map[string][]byte)
	closeWatcher = make(chan bool)
	windowSizeH := 1100
	windowSizeV := 950

	// Get CWD and use it to find if we are in 'cmd' or base of project, then normalize it
	// by removing '/cmd' from end of path so we can find where our support files are located
	exePath, _ = os.Getwd()
	if exePath[len(exePath)-4:] == "\\cmd" || exePath[len(exePath)-4:] == "\\bin" {
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
		Size:     Size{windowSizeH, windowSizeV},
		Layout:   VBox{},
		Children: []Widget{
			HSplitter{
				MaxSize: Size{150, 23},
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
										support.WriteToEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToUse+"/"+
											normalizeKeyNames(modifyKeyBox.Text()), modifyValueBox.Text())

										readValuesAndSendToMsgBox(&config)
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
										numDeleted := support.DeleteFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToUse+"/"+normalizeKeyNames(modifyKeyBox.Text()))
										if numDeleted < 1 {
											walk.MsgBox(nil, "Error", "No records found", walk.MsgBoxIconInformation)
										}
										readValuesAndSendToMsgBox(&config)
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
									readValuesAndSendToMsgBox(&config)
								},
							},
						},
					},
				},
			},
			HSplitter{
				MaxSize: Size{150, 23},
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							TextLabel{
								Text: "Base key to use:",
							},
							TextEdit{
								AssignTo: &baseKeyToUseBox,
								Text:     config.Etcd.BaseKeyToUse,
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Apply",
								OnClicked: func() {
									closeWatcher <- true
									config.Etcd.BaseKeyToUse = baseKeyToUseBox.Text()
									readValuesAndSendToMsgBox(&config)
									go support.WatchReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToUse, watchedChangeCh, closeWatcher)
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
						MinSize:  Size{windowSizeH, windowSizeV - 175},
						OnBoundsChanged: func() {
							resultMsgBox.SetWidth(mw.Width() - 40)
						},
						VScroll: true,
						Font: Font{
							Family:    "Ariel",
							PointSize: 15,
						},
					},
				},
			},
			HSplitter{
				MaxSize: Size{150, 32},
				Children: []Widget{
					ScrollView{
						Layout: HBox{MarginsZero: true},
						Children: []Widget{
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Export DB",
								OnClicked: func() {
									dbImportExport(&config, importExportDirBox.Text(), "export")
								},
							},
							PushButton{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Import DB",
								OnClicked: func() {
									dbImportExport(&config, importExportDirBox.Text(), "import")
								},
							},
							TextLabel{
								MinSize: Size{100, 20},
								MaxSize: Size{100, 20},
								Text:    "Import/Export Dir: ",
							},
							LineEdit{
								AssignTo: &importExportDirBox,
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

	// These need their own thread since they all loop forever
	go readValuesAndSendToMsgBox(&config)
	go support.WatchReadFromEtcd(&config.Etcd.CertPath, &config.Etcd.Endpoints, config.Etcd.BaseKeyToUse, watchedChangeCh, closeWatcher)
	go updateWatchedChanges()
	go mainLoop(&config, resultMsgBox)

	mw.Run()
}
