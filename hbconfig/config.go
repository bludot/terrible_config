package hbconfig

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/jinzhu/configor"
)

type DynamicConfigService struct {
	Config           any
	vaultSecretsPath string
	files            []string
	watcherChan      chan bool
	watcher          *fsnotify.Watcher
}

type DynamicConfig struct{}

func NewDynamicConfig(config any, dirs ...string) *DynamicConfigService {

	files := make([]string, 0)
	for _, dir := range dirs {
		files = append(files, getDirFiles(dir)...)
	}
	if dynamicConfigService == nil {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			log.Fatal(err.Error())
		}
		dynamicConfigService = &DynamicConfigService{
			Config:  config,
			files:   files,
			watcher: watcher,
		}
	}

	dynamicConfigService.files = files
	dynamicConfigService.reload()
	dynamicConfigService.LoadConfig()

	return dynamicConfigService
}

type AutoloadCallback func()

var autoloadCallbacks []*AutoloadCallback

func RegisterAutoloadCallback(callback AutoloadCallback) {
	autoloadCallbacks = append(autoloadCallbacks, &callback)
}

var dynamicConfigService *DynamicConfigService

func getConfigLocation(dir string) string {
	_, filename, _, _ := runtime.Caller(0)

	return path.Join(path.Dir(filename), "../", dir)
}

func getEnv() string {
	val := os.Getenv("APP_ENV")
	// todo: check our stage names and align with them
	switch strings.ToLower(val) {
	case "prod":
		return "prod"
	case "staging":
		return "staging"
	case "test":
		return "test"
	case "qa":
		return "qa"
	default:
		return "dev"
	}
}

func (d *DynamicConfigService) watchFile(file string) {

	d.watcherChan = make(chan bool)
	go func() {
		log.Println("watching env file: ", file)
		for {
			select {
			case event, ok := <-d.watcher.Events:
				if !ok {
					return
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("file changed: ", event.Name)
					err := d.LoadConfig()
					if err != nil {
						log.Println(err.Error())
					}
					for _, callback := range autoloadCallbacks {
						(*callback)()
					}
				}
			case err := <-d.watcher.Errors:
				log.Println("config watcher error: ", err.Error())
			}
		}
	}()

	err := d.watcher.Add(file)
	if err != nil {
		log.Fatal(err.Error())
	}
	<-d.watcherChan
}

func getDirFiles(dir string) []string {
	relativeDir := path.Join(dir)
	// get files in dir
	files, err := ioutil.ReadDir(relativeDir)
	if err != nil {
		panic(err)
	}

	fileNames := []string{}

	for _, file := range files {
		// get file name
		fileName := file.Name()
		// if file contains .json, add to list
		if strings.Contains(fileName, ".json") {
			// add to list of files
			fileNames = append(fileNames, relativeDir+"/"+fileName)
		}
	}

	return fileNames
}

var startedWatch = false

func (d *DynamicConfigService) LoadConfig() error {

	configFiles := d.files

	configFiles = append(configFiles, []string{fmt.Sprintf("%s/config.%s.json", getConfigLocation("config"), getEnv())}...)

	err := configor.
		New(&configor.Config{AutoReload: false}).
		Load(&d.Config, configFiles...)

	if err != nil {
		log.Println(err.Error())
	}

	if !startedWatch {
		startedWatch = true
		go d.watchFile(fmt.Sprintf("%s/config.%s.json", getConfigLocation("config"), getEnv()))
		for _, file := range d.files {
			go d.watchFile(file)
		}
	}

	return nil
}

func GetDynamicConfig() any {
	return dynamicConfigService.Config
}

func (d *DynamicConfigService) reload() {
	oldFiles := d.watcher.WatchList()
	for _, file := range oldFiles {
		d.watcher.Remove(file)
	}
	d.watcher.Add(fmt.Sprintf("%s/config.%s.json", getConfigLocation("config"), getEnv()))

	for _, file := range d.files {
		d.watcher.Add(file)
	}

}
