package tail

import (
	"bufio"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io"
	"log"
	"os"
	"strings"
	"time"
)
const filename = "/workspace/tmp/openresty-config/logs/access.log"

type Line struct {
	Text string
	Time time.Time
	Err  error
}

type SeekInfo struct {
	Offset int64
	Whence int
}

type Tail struct {
	Filename string
	Lines chan *Line
	Location *SeekInfo
	file   *os.File
	reader *bufio.Reader
	fileLen int64
	watcher *fsnotify.Watcher
}

func(t *Tail)ReadFile() {
	fi, err := os.Stat(t.Filename)
	if err != nil || fi.IsDir() {
		return
	}
	t.fileLen = fi.Size()
	f, err := os.OpenFile(t.Filename, os.O_RDONLY , 0666)
	if err != nil {
		log.Fatal(err)
	}
	t.file = f
}

func (t *Tail)InitReader() {
	t.reader = bufio.NewReader(t.file)
}

func (t *Tail)InitWatcher()  {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("初始化文件watcher失败", err)
	}
	err = watcher.Add(t.Filename)
	if err != nil {
		log.Fatal("notify 文件失败", err)
	}
	t.watcher = watcher
}

func (t *Tail)ReadLine()(string, error) {
	line, err := t.reader.ReadString('\n')
	if err != nil {
		return line, err
	}
	line = strings.TrimRight(line, "\n")
	return line, nil
}

func (t *Tail) Tell() (offset int64, err error) {
	if t.file == nil {
		return
	}
	offset, err = t.file.Seek(0, 1)
	//fmt.Printf("当前位置%d ", offset)
	if err != nil {
		return
	}
	if t.reader == nil {
		return
	}
	offset -= int64(t.reader.Buffered())
	//fmt.Printf("可读buffer长度%d, 当前位置 %d\n", t.reader.Buffered(), offset)
	return
}

func SeekFile() {
	tail := &Tail{
		Filename: filename,
		Lines:    make(chan *Line),
		Location:&SeekInfo{
			Offset: int64(0),
			Whence: 0,
		},
	}
	tail.ReadFile()
	tail.InitWatcher()
	tail.InitReader()
	var offset int64
	for {
		offset, _ = tail.Tell()
		line, err := tail.ReadLine()
		if err == nil {
			fmt.Println(line)
		} else if err == io.EOF {
			_, err = tail.file.Seek(offset, 0)
			if err != nil {
				return
			}
			// 通过chanel 不然死循环一直去读取文件流 会爆炸
			var modifyChan = make(chan bool)
			go func(modifyChan chan bool) {
				for {
					select {
					case event, ok := <-tail.watcher.Events:
						if !ok {
							return
						}
						if event.Op&fsnotify.Write == fsnotify.Write {
							modifyChan <- true
						}
					case e, ok := <- tail.watcher.Errors:
						if !ok {
							return
						}
						fmt.Println(e)
					}
				}
			}(modifyChan)
			<- modifyChan
		} else {
			log.Fatal(err)
		}
	}
}