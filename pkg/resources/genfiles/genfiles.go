package genfiles

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"loggie-test/pkg/resources"
	"loggie-test/pkg/tools/generator"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

const Name = "genfiles"

func init() {
	resources.Register(Name, makeGenFiles)
}

type Config struct {
	Dir       string `yaml:"dir,omitempty" default:"/tmp/loggie"`
	LineBytes int    `yaml:"lineBytes,omitempty" default:"1024"`
	LineCount int    `yaml:"lineCount,omitempty" default:"1024000"`

	FileCount int `yaml:"fileCount" default:"1"`

	// TODO multiline mock
}

var _ resources.Resource = (*GenFiles)(nil)

type GenFiles struct {
	Conf *Config

	setupErr     error
	setupDone    bool
	PendingFiles map[string]struct{}
}

func makeGenFiles() interface{} {
	return &GenFiles{
		Conf:         &Config{},
		PendingFiles: make(map[string]struct{}),
	}
}

func (r *GenFiles) Config() interface{} {
	return r.Conf
}

func (r *GenFiles) Name() string {
	return Name
}

func (r *GenFiles) Setup(ctx context.Context) error {
	// check dir exist
	_, err := os.Stat(r.Conf.Dir)
	if err != nil {
		if !os.IsNotExist(err) {
			r.setupErr = err
			return err
		}
		if err := os.MkdirAll(r.Conf.Dir, 0777); err != nil {
			r.setupErr = err
			return err
		}
	}

	var wg sync.WaitGroup
	wg.Add(r.Conf.FileCount)
	for i := 0; i < r.Conf.FileCount; i++ {
		fileName := genFileName(r.Conf.Dir, i)
		r.PendingFiles[fileName] = struct{}{}
		go func() {
			defer wg.Done()
			r.writeFile(ctx, fileName)
		}()
	}
	wg.Wait()
	r.setupDone = true
	return nil
}

func genFileName(dir string, index int) string {
	return filepath.Join(dir, fmt.Sprintf("test-%d.log", index))
}

func (r *GenFiles) writeFile(ctx context.Context, filename string) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
	if err != nil {
		r.setupErr = err
		return
	}
	defer file.Close()

	err = generator.WriteLines(ctx, file, r.Conf.LineCount, r.Conf.LineBytes)
	if err != nil {
		r.setupErr = err
		return
	}
}

func (r *GenFiles) CleanUp(ctx context.Context) error {
	for i := 0; i < r.Conf.FileCount; i++ {
		fileName := genFileName(r.Conf.Dir, i)
		if err := os.Remove(fileName); err != nil {
			return err
		}
	}
	return nil
}

func (r *GenFiles) Ready() (bool, error) {
	if r.setupErr != nil {
		return false, r.setupErr
	}

	for fileName := range r.PendingFiles {
		cmd := exec.Command("wc", "-l", fileName)
		ret, err := cmd.CombinedOutput()
		if err != nil {
			return false, err
		}

		splits := strings.Split(string(ret), " ")
		if len(splits) < 1 {
			return false, errors.New("exec wc split error")
		}

		retCount, err := strconv.Atoi(splits[0])
		if err != nil {
			return false, err
		}

		if retCount == r.Conf.LineCount {
			delete(r.PendingFiles, fileName)
		}
	}

	if len(r.PendingFiles) == 0 {
		return true, nil
	}

	return false, nil
}

func (r *GenFiles) AllCount() int64 {
	return int64(r.Conf.FileCount * r.Conf.LineCount)
}
