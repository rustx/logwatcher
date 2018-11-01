package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
)

func (lw *Logwatcher) UpdateMainView(g *gocui.Gui) error {
	g.Update(func(g *gocui.Gui) error {
		mainV, err := g.View("main")
		if err != nil {
			return err
		}
		mainV.Clear()
		fmt.Fprintf(mainV,
			"%sDate Now: %s\n%sTime Elapsed : %s\n\n%s"+
				"Refresh Interval : %d s%sAlert Interval : %d s%sLog Interval : %d ms%sAlert Threshold : %d\n",
			margin, lw.Date(), tab, lw.TimeElapsed(), tab,
			lw.RefreshInterval, tab, lw.AlertInterval, tab, lw.LogInterval, tab, lw.AlertThreshold)
		return nil

	})
	return nil
}

func (lw *Logwatcher) UpdateStatsTotalView(g *gocui.Gui) error {
	g.Update(func(g *gocui.Gui) error {
		statsTotalV, err := g.View("stats_total")
		if err != nil {
			return err
		}
		statsTotalV.Clear()
		fmt.Fprintf(statsTotalV,
			"%sTotal Hits : %d\n%sTotal 2XX  : %v\n%sTotal 3XX  : %d\n%sTotal 4XX  : %d\n%sTotal 5XX  : %d\n\n",
			margin, lw.TotalHits, margin, lw.Total2xx, margin, lw.Total3xx, margin, lw.Total4xx, margin, lw.Total5xx)
		return nil

	})
	return nil
}

func (lw *Logwatcher) UpdateStatsAvgView(g *gocui.Gui) error {
	g.Update(func(g *gocui.Gui) error {
		statsAvgV, err := g.View("stats_avg")
		if err != nil {
			return err
		}
		statsAvgV.Clear()
		fmt.Fprintf(statsAvgV,
			"%sAvg Hits : %d\n%sAvg 2XX  : %d\n%sAvg 3XX  : %d\n%sAvg 4XX  : %d\n%sAvg 5XX  : %d\n\n",
			margin, lw.AvgHits, margin, lw.Avg2xx, margin, lw.Avg3xx, margin, lw.Avg4xx, margin, lw.Avg5xx)
		return nil

	})
	return nil
}

func (lw *Logwatcher) UpdateLogTailView(g *gocui.Gui, logEvents []string) error {
	g.Update(func(g *gocui.Gui) error {
		logTailV, err := g.View("log_tail")
		if err != nil {
			return err
		}
		logTailV.Clear()
		for _, line := range logEvents {
			fmt.Fprintf(logTailV, margin+line)
		}
		logEvents = logEvents[0:0]
		return nil
	})
	return nil
}

func (lw *Logwatcher) UpdateTopSectionsView(g *gocui.Gui) error {
	g.Update(func(g *gocui.Gui) error {
		topSectionsV, err := g.View("top_sections")
		if err != nil {
			return err
		}
		topSectionsV.Clear()
		fmt.Fprintf(topSectionsV, "%sTop Sections :\n%v%s",
			margin, lw.TopSectionsMsg, margin)
		return nil

	})
	return nil
}

func (lw *Logwatcher) UpdateTopStatusView(g *gocui.Gui) error {
	g.Update(func(g *gocui.Gui) error {
		topStatusV, err := g.View("top_status")
		if err != nil {
			return err
		}
		topStatusV.Clear()
		fmt.Fprintf(topStatusV, "%sTop Status :\n%v%s",
			margin, lw.TopStatusMsg, margin)
		return nil
	})
	return nil
}

func (lw *Logwatcher) UpdateAlertView(g *gocui.Gui) error {
	g.Update(func(g *gocui.Gui) error {
		alertV, err := g.View("alert")
		if err != nil {
			return err
		}
		alertV.Clear()
		if lw.AvgHits > lw.AlertThreshold {
			lw.AlertMsg = append(lw.AlertMsg,
				fmt.Sprintf("%sHigh traffic generated an alert - average hits = %d, triggered at %s",
					margin, lw.AvgHits, lw.Date()))
			alertV.BgColor = gocui.ColorRed
			lw.AlertState = true
		} else {
			if lw.AvgHits < lw.AlertThreshold {
				if lw.AlertState == true {
					log.Println("Recover generated at : ", lw.Date())
					lw.AlertMsg = append(lw.AlertMsg,
						fmt.Sprintf("%sLow traffic generated a recover - average hits = %d, triggered at %s",
							margin, lw.AvgHits, lw.Date()))
					alertV.BgColor = gocui.ColorGreen
					lw.AlertState = false
				} else {
					alertV.BgColor = gocui.ColorDefault
				}
			}
		}
		for _, msg := range lw.AlertMsg {
			fmt.Fprintf(alertV, msg)
		}
		return nil
	})
	return nil
}
