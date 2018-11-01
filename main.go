package main

import (
	"fmt"
	"log"
	"log/syslog"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hpcloud/tail"
	"github.com/jessevdk/go-flags"
	"github.com/jroimartin/gocui"
)

// CommonLog is a struct collecting fields for http logs in common format.
type CommonLog struct {
	IP         string
	Identifier string
	User       string
	Date       string
	Method     string
	Request    string
	Proto      string
	Status     int
	Bytes      int64
}

// StatItem is a struct collecting log information during execution.
type StatItem struct {
	Timestamp   time.Time
	Hits        int
	Status2xx   int
	Status3xx   int
	Status4xx   int
	Status5xx   int
	TopSections map[string]int
	TopStatus   map[string]int
}

type StatsTotal struct {
	TotalHits      int
	Total2xx       int
	Total3xx       int
	Total4xx       int
	Total5xx       int
	TopSectionsMsg string
	TopStatusMsg   string
}

type StatsAvg struct {
	AvgHits int
	Avg2xx  int
	Avg3xx  int
	Avg4xx  int
	Avg5xx  int
}

// Logwatcher is the struct launching the application.
type Logwatcher struct {
	StartTime     time.Time
	AlertMsg      []string
	AlertState    bool
	CollectionNum int
	*Config
	*StatsTotal
	*StatsAvg
}

var (
	wg sync.WaitGroup
	mu sync.Mutex

	done     = make(chan bool)
	logTailC = make(chan CommonLog)
	logDumpC = make(chan string)

	margin = "\t\n\t\t\t\t\t"
	tab    = "\t\t\t\t\t"
)

func sortMap(m map[string]int) (msg string) {

	n := map[int][]string{}
	a := make([]int, 0)
	msg = margin + tab

	for k, v := range m {
		n[v] = append(n[v], k)
	}
	for k := range n {
		a = append(a, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(a)))
	for _, k := range a {
		for _, s := range n[k] {
			msg += fmt.Sprintf("%s%s%s : %d", margin, tab, s, k)
		}
	}
	return msg
}

func (lw *Logwatcher) TimeElapsed() string {
	return time.Duration(time.Duration(time.Now().Unix()-lw.StartTime.Unix()) * time.Second).String()
}

func (lw *Logwatcher) Date() string {
	return time.Now().Format(time.StampMilli)
}

func (lw *Logwatcher) LogReader() error {
	start := tail.SeekInfo{
		Offset: 0,
		Whence: 2,
	}

	stream, err := tail.TailFile(lw.LogFile, tail.Config{
		Follow:    true,
		ReOpen:    true,
		Location:  &start,
		MustExist: true,
		Logger:    tail.DiscardingLogger,
	})

	if err != nil {
		log.Println(err)
		return err
	}

	//127.0.0.1 - - [11/May/2016:22:02:21 +0200] "GET /assets/avatars/avatar4.png HTTP/1.1" 304 0
	re := regexp.MustCompile(`^(?P<Ip>[\d\.]+) (?P<identifier>.*) (?P<user>.*) \[(?P<date>.*)\] "(?P<method>.*) (?P<request>.*) (?P<proto>.*)" (?P<status>\d+) (?P<bytes>\d+)`)

	for item := range stream.Lines {
		res := re.FindStringSubmatch(item.Text)
		bytes, _ := strconv.ParseInt(res[9], 10, 64)
		status, _ := strconv.Atoi(res[8])

		statitem := CommonLog{
			IP:         res[1],
			Identifier: res[2],
			User:       res[3],
			Date:       res[4],
			Method:     res[5],
			Request:    res[6],
			Proto:      res[7],
			Status:     status,
			Bytes:      bytes,
		}

		logTailC <- statitem
		logDumpC <- res[0]
	}

	return nil
}

func (lw *Logwatcher) LoadOnRefresh(item *StatItem, tmpStat *StatsAvg) {
	tmpStat.AvgHits += item.Hits
	tmpStat.Avg2xx += item.Status2xx
	tmpStat.Avg3xx += item.Status3xx
	tmpStat.Avg4xx += item.Status4xx
	tmpStat.Avg5xx += item.Status5xx

	lw.TotalHits += item.Hits
	lw.Total2xx += item.Status2xx
	lw.Total3xx += item.Status3xx
	lw.Total4xx += item.Status4xx
	lw.Total5xx += item.Status5xx

	lw.TopSectionsMsg = sortMap(item.TopSections)
	lw.TopStatusMsg = sortMap(item.TopStatus)
}

func (lw *Logwatcher) LoadOnAlert(tmpStat *StatsAvg) {
	lw.AvgHits = tmpStat.AvgHits / lw.CollectionNum
	lw.Avg2xx = tmpStat.Avg2xx / lw.CollectionNum
	lw.Avg3xx = tmpStat.Avg3xx / lw.CollectionNum
	lw.Avg4xx = tmpStat.Avg4xx / lw.CollectionNum
	lw.Avg5xx = tmpStat.Avg5xx / lw.CollectionNum
}

func (lw *Logwatcher) CollectStatItems(logStats *[]*CommonLog) *StatItem {
	item := StatItem{
		Timestamp:   time.Now(),
		TopSections: make(map[string]int),
		TopStatus:   make(map[string]int),
	}
	for _, event := range *logStats {
		switch event.Status / 100 {
		case 2:
			item.Status2xx++
		case 3:
			item.Status3xx++
		case 4:
			item.Status4xx++
		case 5:
			item.Status5xx++
		}
		item.Hits++
		section := "/" + strings.Split(event.Request, "/")[1]
		item.TopSections[section]++
		item.TopStatus[strconv.Itoa(event.Status)]++
	}
	return &item
}

func (lw *Logwatcher) PurgeTmpStat(tmpStat *StatsAvg) {
	tmpStat.AvgHits = 0
	tmpStat.Avg2xx = 0
	tmpStat.Avg3xx = 0
	tmpStat.Avg4xx = 0
	tmpStat.Avg5xx = 0
}

func (lw *Logwatcher) Run(g *gocui.Gui) error {

	mainTicker := time.NewTicker(time.Duration(1) * time.Second)
	logTicker := time.NewTicker(time.Duration(lw.LogInterval) * time.Millisecond)
	alertTicker := time.NewTicker(time.Duration(lw.AlertInterval) * time.Second)
	refreshTicker := time.NewTicker(time.Duration(lw.RefreshInterval) * time.Second)

	logStats := make([]*CommonLog, 0)
	logEvents := make([]string, 0)

	tmpStat := StatsAvg{}

	defer wg.Done()

	for {
		select {

		case <-done:
			return nil

		case logStat := <-logTailC:
			logStats = append(logStats, &logStat)

		case logEvent := <-logDumpC:
			logEvents = append(logEvents, logEvent)

		case <-logTicker.C:
			lw.UpdateLogTailView(g, logEvents)

		case <-mainTicker.C:
			lw.UpdateMainView(g)

		case <-refreshTicker.C:
			mu.Lock()
			lw.LoadOnRefresh(lw.CollectStatItems(&logStats), &tmpStat)
			//lw.PurgeLogStats(&logStats)
			logStats = logStats[0:0]
			mu.Unlock()

			lw.UpdateStatsTotalView(g)
			lw.UpdateTopSectionsView(g)
			lw.UpdateTopStatusView(g)

		case <-alertTicker.C:
			mu.Lock()
			lw.LoadOnAlert(&tmpStat)
			lw.PurgeTmpStat(&tmpStat)
			mu.Unlock()

			lw.UpdateAlertView(g)
			lw.UpdateStatsAvgView(g)
		}
	}
}

func main() {

	if _, err := flags.Parse(&config); err != nil {
		fmt.Printf("Default :\n")
		fmt.Printf("logwatcher --log-file /var/log/nginx/access.log --refresh-interval 10")
		fmt.Printf("--alert-interval 120 --alert-threshold 400\n")
		os.Exit(1)
	}

	if int(math.Mod(float64(config.AlertInterval), float64(config.RefreshInterval))) != 0 {
		fmt.Printf("Please review your options, or keep default options to run this program.\nThe modulo of " +
			"alertInterval / refreshInterval must be zero for average calculation to work\nTry logwatcher -h\n")
		os.Exit(1)
	}

	logWriter, err := syslog.New(syslog.LOG_NOTICE, "logwatcher")
	if err == nil {
		log.SetOutput(logWriter)
	}

	lw := Logwatcher{
		StartTime:     time.Now(),
		Config:        &config,
		AlertState:    false,
		StatsTotal:    &StatsTotal{},
		StatsAvg:      &StatsAvg{},
		AlertMsg:      make([]string, 0),
		CollectionNum: config.AlertInterval / config.RefreshInterval,
	}

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	defer g.Close()

	g.SetManagerFunc(Layout)
	if err := Keybindings(g); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	wg.Add(1)

	go lw.Run(g)
	go lw.LogReader()

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatal(err)
		os.Exit(1)
	}

	wg.Wait()

}
