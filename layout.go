package main

import (
	"fmt"
	"time"

	"github.com/jroimartin/gocui"
)

func Layout(g *gocui.Gui) error {

	maxX, maxY := g.Size()

	if mainV, err := g.SetView("main", 0, 0, maxX-1, maxY/8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		mainV.Frame = true
		mainV.Autoscroll = true
		mainV.BgColor = gocui.ColorDefault
		mainV.Title = " Log Watcher | Main Information "
		fmt.Fprintf(mainV, "%sDate Now : %s\n%sTime Elapsed : 0s %s%s",
			margin, time.Now().Format(time.StampMilli), tab, tab, tab)
	}

	if statsTotalV, err := g.SetView("stats_total",
		0, maxY/8, maxX/2-1, maxY/4+maxY/8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		statsTotalV.Frame = true
		statsTotalV.Autoscroll = true
		statsTotalV.BgColor = gocui.ColorDefault
		statsTotalV.Title = fmt.Sprintf(" Stats Total | Every %d s ", config.RefreshInterval)
		fmt.Fprintf(statsTotalV,
			"%sTotal Hits : 0\n%sTotal 2XX  : 0\n%sTotal 3XX  : 0\n%sTotal 4XX  : 0\n%sTotal 5XX  : 0\n\n",
			margin, margin, margin, margin, margin)

	}

	if statsAvgV, err := g.SetView("stats_avg",
		maxX/2-1, maxY/8, maxX-1, maxY/4+maxY/8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		statsAvgV.Frame = true
		statsAvgV.Autoscroll = false
		statsAvgV.BgColor = gocui.ColorDefault
		statsAvgV.Title = fmt.Sprintf(" Stats Average | Every %d s ", config.AlertInterval)
		fmt.Fprintf(statsAvgV,
			"%sAvg Hits : 0\n%sAvg 2XX  : 0\n%sAvg 3XX  : 0\n%sAvg 4XX  : 0\n%sAvg 5XX  : 0\n\n",
			margin, margin, margin, margin, margin)

	}

	if logTailV, err := g.SetView("log_tail",
		0, maxY/4+maxY/8, maxX-1, maxY/2+maxY/8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		logTailV.Frame = true
		logTailV.Autoscroll = true
		logTailV.BgColor = gocui.ColorDefault
		logTailV.Title = fmt.Sprintf(" Log Tail | Every %d ms", config.LogInterval)
		fmt.Fprintf(logTailV, "%sLog tail view", margin)

	}

	if topSectionsV, err := g.SetView("top_sections",
		0, maxY/2+maxY/8, maxX/2-1, maxY-maxY/8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		topSectionsV.Frame = true
		topSectionsV.Autoscroll = false
		topSectionsV.BgColor = gocui.ColorDefault
		topSectionsV.Title = fmt.Sprintf("Top Sections | Every %d s", config.RefreshInterval)
		fmt.Fprintf(topSectionsV, "%sTop Sections:\n\n", margin)
	}

	if topStatusV, err := g.SetView("top_status",
		maxX/2-1, maxY/2+maxY/8, maxX-1, maxY-maxY/8); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		topStatusV.Frame = true
		topStatusV.Autoscroll = true
		topStatusV.BgColor = gocui.ColorDefault
		topStatusV.Title = fmt.Sprintf(" Top Status | Every %d s ", config.RefreshInterval)
		fmt.Fprintf(topStatusV, "%sTop StatusCode:\n\n", margin)

	}

	if alertV, err := g.SetView("alert",
		0, maxY-maxY/8, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		alertV.Frame = true
		alertV.Autoscroll = true
		alertV.BgColor = gocui.ColorDefault
		alertV.Title = fmt.Sprintf(" Alerting | Every %d s | Alert Threshold : %d",
			config.AlertInterval, config.AlertThreshold)
		fmt.Fprintf(alertV, "%sNo alert for now (%s)\n\n", margin, time.Now().Format(time.StampMilli))

	}
	return nil
}

func Keybindings(g *gocui.Gui) error {
	return g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, Quit)
}

func Quit(g *gocui.Gui, v *gocui.View) error {
	done <- true
	return gocui.ErrQuit
}
