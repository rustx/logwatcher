package main

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"

	. "gopkg.in/check.v1"
)

func Test(t *testing.T) { TestingT(t) }

type LogwatcherSuite struct {
	dir  string
	file string
}

const letters = "abcdefghijklmnopqrstuvwxyz"

var (
	_      = Suite(&LogwatcherSuite{})
	secret = rand.NewSource(time.Now().UnixNano())
)

func randIp(r *rand.Rand) string {
	return fmt.Sprintf("%d.%d.%d.%d", r.Intn(255), r.Intn(255), r.Intn(255), r.Intn(255))
}

func randReq(r *rand.Rand) string {
	req := make([]byte, 8)
	for i := range req {
		req[i] = letters[r.Intn(len(letters))]
	}
	return string(req)
}

func randStatusList(r *rand.Rand) []int {
	statusList := make([]int, 0)

	for i := 0; i < 100; i++ {
		statusList = append(statusList, r.Intn(99)+100)
		statusList = append(statusList, r.Intn(99)+200)
		statusList = append(statusList, r.Intn(99)+300)
		statusList = append(statusList, r.Intn(99)+400)
	}
	return statusList
}

func generateLogLines(valid bool) (string, *CommonLog) {
	r := rand.New(secret)

	statusList := randStatusList(r)

	cl := &CommonLog{}
	if valid == true {
		cl.Identifier = "\"-\""
		cl.User = "\"-\""
		cl.Method = "GET"
		cl.Proto = "HTTP/1.1"
	} else {
		cl.Identifier = "---- ----"
		cl.User = "---- ----"
		cl.Method = "POUET POUET POUET"
		cl.Proto = "HTTPARTY/2.0 DISCO EDITION *_*"
	}
	cl.Request = randReq(r)
	cl.IP = randIp(r)
	cl.Date = time.Now().Format(time.StampMilli)
	cl.Bytes = int64(r.Intn(500))
	cl.Status = statusList[r.Intn(len(statusList))]

	log := fmt.Sprintf("%s %s %s [%s] \"%s /%s %s\" %d %d\n",
		cl.IP, cl.Identifier, cl.User, cl.Date, cl.Method, cl.Request, cl.Proto, cl.Status, cl.Bytes)

	return log, cl
}

func writeTmpLogFile(file string, lines int, valid bool) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("error opening file")
		return err
	}
	defer f.Close()
	for i := 0; i < lines; i++ {
		log, _ := generateLogLines(valid)
		if err != nil {
			fmt.Println("error writing file")
			return err
		}
		f.WriteString(log)
		f.Sync()
	}
	return nil
}

func (s *LogwatcherSuite) SetUpSuite(c *C) error {
	s.dir = c.MkDir()
	s.file = "access.log"

	logFile := filepath.Join(s.dir, s.file)
	err := writeTmpLogFile(logFile, 100, true)
	if err != nil {
		c.Fatal(err)
	}
	_, err = ioutil.ReadFile(logFile)
	if err != nil {
		c.Fatal(err)
	}

	return nil
}

func (s LogwatcherSuite) SetUpTest(c *C) {
	f, err := os.Create(filepath.Join(s.dir, s.file))
	if err != nil {
		c.Fatal(err)
	}
	f.Sync()
}

func (s *LogwatcherSuite) TearDownTest(c *C) {
	os.Chmod(s.dir, 0755)
	os.Remove(filepath.Join(s.dir, s.file))
}

func (s *LogwatcherSuite) TestLogReaderOk(c *C) {
	mainTimer := time.NewTimer(time.Duration(5) * time.Second)

	startTimer := time.NewTimer(time.Duration(1) * time.Second)

	logStats := make([]*CommonLog, 0)
	logEvents := make([]string, 0)

	lw := Logwatcher{
		StartTime: time.Now(),
		Config:    &config,
	}
	lw.LogFile = filepath.Join(s.dir, s.file)

	go lw.LogReader()

loop:
	for {
		select {
		case logStat := <-logTailC:
			c.Log(logStat)
			logStats = append(logStats, &logStat)
			c.Assert(logStat, FitsTypeOf, CommonLog{})
		case logEvent := <-logDumpC:
			c.Log(logEvent)
			c.Assert(logEvent, FitsTypeOf, "")
			logEvents = append(logEvents, logEvent)
		case <-startTimer.C:
			go writeTmpLogFile(lw.LogFile, 100, true)
			c.Log("Entering writeTmpLogfile valid")
			//go writeTmpLogFile(lw.LogFile, 100, false)
			//c.Log("Entering writeTmpLogfile not valid")
		case <-mainTimer.C:
			break loop
		}
	}

	c.Log("LogStats length ", len(logStats))
	c.Log("LogEvents length ", len(logEvents))
	c.Assert(logStats, HasLen, 100)
	c.Assert(logEvents, HasLen, 100)
}

func (s *LogwatcherSuite) TestLogReaderFileNoExistKo(c *C) {
	logStats := make([]*CommonLog, 0)
	logEvents := make([]string, 0)

	lw := Logwatcher{
		StartTime: time.Now(),
		Config:    &config,
	}
	lw.LogFile = "test"

	if err := lw.LogReader(); err != nil {
		c.Assert(err, Not(IsNil))
		c.Assert(err, ErrorMatches, "open test: no such file or directory")
	}

	c.Assert(logStats, HasLen, 0)
	c.Assert(logEvents, HasLen, 0)
}

func (s *LogwatcherSuite) TestLogReaderFilePermissionKo(c *C) {
	logStats := make([]*CommonLog, 0)
	logEvents := make([]string, 0)

	lw := Logwatcher{
		StartTime: time.Now(),
		Config:    &config,
	}
	lw.LogFile = filepath.Join(s.dir, s.file)

	if err := os.Chmod(lw.LogFile, 0300); err != nil {
		c.Fatal(err)
	}
	if err := lw.LogReader(); err != nil {
		c.Assert(err, Not(IsNil))
		c.Assert(err, ErrorMatches, "open .*: permission denied")
		c.Assert(logStats, HasLen, 0)
		c.Assert(logEvents, HasLen, 0)
	}
	if err := os.Chmod(lw.LogFile, 0644); err != nil {
		c.Fatal(err)
	}
}

func (s *LogwatcherSuite) TestDate(c *C) {
	lw := Logwatcher{
		StartTime: time.Now(),
		Config:    &config,
	}
	for i := 0; i < 100; i++ {
		date := lw.Date()
		c.Assert(date, Not(IsNil))
		c.Assert(date, FitsTypeOf, "")
		c.Assert(date, FitsTypeOf, time.Now().Format(time.StampMilli))
	}
}

func (s *LogwatcherSuite) TestTimeElapsed(c *C) {
	lw := Logwatcher{
		StartTime: time.Now(),
		Config:    &config,
	}
	for i := 0; i < 100; i++ {
		date := lw.TimeElapsed()
		c.Assert(date, Not(IsNil))
		c.Assert(date, FitsTypeOf, "")
		c.Assert(date, FitsTypeOf, time.Duration(time.Duration(time.Now().Unix()-lw.StartTime.Unix())*time.Second).String())
	}
}

func (s *LogwatcherSuite) TestSortMap(c *C) {
	r := rand.New(secret)
	score := make(map[string]int)

	for i := 0; i < 100; i++ {
		score["200"] = r.Intn(100)
		score["300"] = r.Intn(200)
		score["400"] = r.Intn(900)
		score["500"] = r.Intn(500)
		result := sortMap(score)

		c.Log(result)
		c.Assert(result, Not(IsNil))
	}

}
