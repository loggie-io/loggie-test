package main

import (
	"context"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"loggie-test/pkg/resources/genfiles"
	"time"
)

func main() {
	// read config
	fmt.Println("read config")
	defaultConfigPath := "./config.yml"
	content, err := ioutil.ReadFile(defaultConfigPath)
	if err != nil {
		panic(err)
	}
	fmt.Printf("config:\n%s", string(content))

	fmt.Println("unmarshal config")
	config := &genfiles.Config{}
	if err = yaml.Unmarshal(content, config); err != nil {
		panic(err)
	}

	// start generate log files
	files := genfiles.GenFiles{
		Conf:         config,
		PendingFiles: make(map[string]struct{}),
	}

	fmt.Println("start generate log files")
	err = files.Setup(context.Background())
	if err != nil {
		panic(err)
	}

	fmt.Println("end generate log files")
	for {
		fmt.Println("-")
		time.Sleep(1 * time.Minute)
	}
}
