package ui

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"encoding/csv"

	"github.com/common-nighthawk/go-figure"
	"github.com/davidscholberg/go-durationfmt"
	"github.com/tamipangadil/mkpis/pkg/vcs"
	"github.com/olekukonko/tablewriter"
)

func AvgDurationFormater(d time.Duration) string {
	t, err := durationfmt.Format(d, "AVG: %dd %hh %mm")
	if err != nil {
		return "ERROR"
	}
	return t
}

func AvgDurationInSecondsFormater(d time.Duration) string {
	t, err := durationfmt.Format(d, "%s")
	if err != nil {
		return "ERROR"
	}
	return t
}

func DurationFormater(d time.Duration) string {

	if d.Microseconds() == 0 {
		return "--"
	}

	t, err := durationfmt.Format(d, "%hh %mm")
	if err != nil {
		return "ERROR"
	}
	return t
}

func DurationInSecondsFormatter(d time.Duration) string {

	if d.Microseconds() == 0 {
		return "--"
	}

	t, err := durationfmt.Format(d, "%s")
	if err != nil {
		return "ERROR"
	}
	return t
}

type CmdUI struct {
	client       vcs.Client
	owner        string
	repo         string
	develBranch  string
	masterBranch string
}

func NewCmdUI(client vcs.Client, owner, repo, develBranch, masterBranch string) *CmdUI {
	return &CmdUI{
		client:       client,
		owner:        owner,
		repo:         repo,
		develBranch:  develBranch,
		masterBranch: masterBranch,
	}
}

func (u CmdUI) Render(from, to time.Time) error {
	rfb, err := u.getFeatureBranchReport(from, to)
	if err != nil {
		return err
	}
	// rrb, err := u.getReleaseBranchReport(from, to)
	// if err != nil {
	// 	return err
	// }
	u.getFeatureBranchReportCSV(from, to)

	// myFigure := figure.NewColorFigure("Printing the reports...", "standard", "white", true)
	// myFigure.Blink(1000, 300, 300)

	// fmt.Println("\033[2J") //clean previous ouput
	u.PrintPageHeader(from, to)
	// u.PrintRepotHeader("Feature Branch Report")
	fmt.Println("")
	fmt.Println("PR Report")
	fmt.Println("")
	fmt.Println(rfb)
	// u.PrintRepotHeader("Release Branch Report")
	// fmt.Println("Release Branch Report")
	// fmt.Println("")
	// fmt.Println(rrb)
	return nil
}

func (u CmdUI) PrintRepotHeader(text string) {
	figure.NewColorFigure(text, "small", "green", true).Print()
	fmt.Println("")
}

func (u CmdUI) PrintPageHeader(from time.Time, to time.Time) {
	// figure.NewColorFigure("MKPIS", "standard", "red", true).Print()
	fLayout := "2006-01-02"
	fmt.Printf("Repo: %s/%s (%s to %s)", u.owner, u.repo, from.Format(fLayout), to.Format(fLayout))
	fmt.Println("")
}

func (u CmdUI) getFeatureBranchReportCSV(from, to time.Time) (string, error) {
	prs, err := u.client.GetMergedPRList(u.owner, u.repo, from, to, u.develBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering information: %s", err.Error())
		return "", err
	}

	fLayout := "2006-01-02"
	csvFilename := fmt.Sprintf("%s-to-%s.csv", from.Format(fLayout), to.Format(fLayout))
	csvfile, err := os.Create(csvFilename)

	if err != nil {
		fmt.Fprintf(os.Stderr, "failed creating file: %s", err.Error())
	}

	csvwriter := csv.NewWriter(csvfile)
	csvwriter.Write([]string{"PR", "Commits", "Size", "Time To First Review", "Review time", "Last Review To Merge", "Comments", "PR Lead Time", "Time To Merge"})

	for _, pr := range prs {
		_ = csvwriter.Write([]string{
			strconv.Itoa(pr.Number),
			strconv.Itoa(pr.Commits),
			strconv.Itoa(pr.ChangedLines),
			DurationInSecondsFormatter(pr.TimeToFirstReview()),
			DurationInSecondsFormatter(pr.TimeToReview()),
			DurationInSecondsFormatter(pr.LastReviewToMerge()),
			strconv.Itoa(pr.ReviewComments),
			DurationInSecondsFormatter(pr.PRLeadTime()),
			DurationInSecondsFormatter(pr.TimeToMerge()),
		})
	}

	csvwriter.Flush()

	csvfile.Close()

	return fmt.Sprintf("Generated CSV: %s", csvFilename), nil
}

func (u CmdUI) getFeatureBranchReport(from, to time.Time) (string, error) {
	prs, err := u.client.GetMergedPRList(u.owner, u.repo, from, to, u.develBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering information: %s", err.Error())
		return "", err
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"PR", "Commits", "Size", "Time To First Review", "Review time", "Last Review To Merge", "Comments", "PR Lead Time", "Time To Merge"})

	for _, pr := range prs {
		table.Append([]string{
			strconv.Itoa(pr.Number),
			strconv.Itoa(pr.Commits),
			strconv.Itoa(pr.ChangedLines),
			DurationFormater(pr.TimeToFirstReview()),
			DurationFormater(pr.TimeToReview()),
			DurationFormater(pr.LastReviewToMerge()),
			strconv.Itoa(pr.ReviewComments),
			DurationFormater(pr.PRLeadTime()),
			DurationFormater(pr.TimeToMerge()),
		})

	}

	kpi := vcs.NewKPICalculator(prs)

	table.SetFooter([]string{
		fmt.Sprintf("Count: %d", kpi.CountPR()),
		fmt.Sprintf("AVG: %.2f", kpi.AvgCommits()),
		fmt.Sprintf("AVG: %.2f", kpi.AvgChangedLines()),
		AvgDurationFormater(kpi.AvgTimeToFirstReview()),
		AvgDurationFormater(kpi.AvgTimeToReview()),
		AvgDurationFormater(kpi.AvgLastReviewToMerge()),
		fmt.Sprintf("AVG: %.2f", kpi.AvgReviews()),
		AvgDurationFormater(kpi.AvgPRLeadTime()),
		AvgDurationFormater(kpi.AvgTimeToMerge()),
	}) // Add Footer
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(true)
	table.Render() // Send output
	return tableString.String(), nil
}

func (u CmdUI) getReleaseBranchReport(from, to time.Time) (string, error) {
	prs, err := u.client.GetMergedPRList(u.owner, u.repo, from, to, u.masterBranch)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error gathering information: %s", err.Error())
		return "", err
	}

	tableString := &strings.Builder{}
	table := tablewriter.NewWriter(tableString)
	table.SetHeader([]string{"PR", "Commits", "Size", "PR Lead Time", "Time To Merge"})

	for _, pr := range prs {
		table.Append([]string{
			strconv.Itoa(pr.Number),
			strconv.Itoa(pr.Commits),
			strconv.Itoa(pr.ChangedLines),
			DurationFormater(pr.PRLeadTime()),
			DurationFormater(pr.TimeToMerge()),
		})

	}

	kpi := vcs.NewKPICalculator(prs)

	table.SetFooter([]string{
		fmt.Sprintf("Count: %d", kpi.CountPR()),
		fmt.Sprintf("AVG: %.2f", kpi.AvgCommits()),
		fmt.Sprintf("AVG: %.2f", kpi.AvgChangedLines()),
		AvgDurationFormater(kpi.AvgPRLeadTime()),
		AvgDurationFormater(kpi.AvgTimeToMerge()),
	}) // Add Footer
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.Render() // Send output
	return tableString.String(), nil
}
