package mario

import (
	"context"
	"fmt"
	"math"
	"slices"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datafactory/armdatafactory/v3"
	"github.com/fatih/color"
)

type RunStats struct {
	pipelineName   string
	pipelineResult string
	startTime      time.Time
	endTime        time.Time
	durationMs     int32
}

func AnalyzeRuns(nDays int, name string) {
	defer timer("AnalyzeRuns")()
	factory := getFactoryClient()
	ctx := context.Background()

	pipelineRuns, _ := getPipelineRuns(&factory, ctx, nDays, name)
	runStats, durations := collectPipelineRunStats(pipelineRuns)
	printTimeseries(name, runStats, durations)

}

func printTimeseries(name string, runStats []RunStats, durations []int32) {
	defer timer("printTimeseries")()
	var (
		barCharacter       = "\u25A4"
		maxBarLength int32 = 40
		minBarLength       = maxBarLength / 20
		headerLength       = 80
	)

	header := createHeader(
		"ANALYZE",
		headerLength,
		color.New(color.FgBlue),
		"=",
		true,
	)
	footer := createHeader("", headerLength, color.New(color.FgWhite), "=", true)

	fmt.Println(header)

	if len(runStats) == 0 {
		fmt.Println("No runs found matching", name)
		fmt.Println(footer)
		return
	}

	minDuration := slices.Min(durations)
	maxDuration := slices.Max(durations)

	color.New(color.Underline).Println(name)
	fmt.Println()

	for i, run := range runStats {
		var (
			duration     = run.durationMs
			runDistance  = duration - minDuration
			barDistance  = maxDuration - minDuration
			durationTime = time.Duration(duration) * time.Millisecond

			startTimeFormatted = run.startTime.Format("2006-01-02 15:04:05")
			durationFormatted  = durationTime.Truncate(time.Second).String()

			previousDuration int32   = 0
			pctDiff          float64 = 0
			pctDiffFormatted string  = "0"
		)

		barLengthFloat := float64(
			runDistance,
		) / float64(
			barDistance,
		) * float64(
			maxBarLength,
		)
		barLength := int32(barLengthFloat)

		if barLength < minBarLength {
			barLength = minBarLength
		}

		if barLength > maxBarLength {
			barLength = maxBarLength
		}

		bar := strings.Repeat(barCharacter, int(barLength))

		if i > 0 {
			previousDuration = int32(runStats[i-1].durationMs)
			pctDiff = float64(duration-previousDuration) / float64(previousDuration) * 100
			pctDiffFormatted = fmt.Sprintf("%.2f", math.Abs(pctDiff))
		}

		switch {
		case pctDiff > 0:
			pctDiffFormatted = failureColor()("\u2191", pctDiffFormatted, "%")
		case pctDiff < 0:
			pctDiffFormatted = successColor()("\u2193", pctDiffFormatted, "%")
		default:
			pctDiffFormatted = neutralColor()(pctDiffFormatted, "%")
		}

		switch {
		case run.pipelineResult == "Succeeded":
			bar = successColor()(bar)
		case run.pipelineResult == "Failed":
			bar = failureColor()(bar)
		case run.pipelineResult == "Cancelled":
			bar = color.New(color.FgYellow).Sprint(bar)
		default:
			bar = neutralColor()(bar)
		}

		fmt.Println(startTimeFormatted, bar, durationFormatted, pctDiffFormatted)

	}

	fmt.Println(footer)

}

func collectPipelineRunStats(
	pipelineRuns armdatafactory.PipelineRunsClientQueryByFactoryResponse,
) ([]RunStats, []int32) {
	defer timer("collectPipelineRunStats")()

	runStats := make([]RunStats, len(pipelineRuns.Value))
	durations := make([]int32, len(pipelineRuns.Value))
	for i, run := range pipelineRuns.Value {
		runStats[i] = RunStats{
			pipelineName:   *run.PipelineName,
			pipelineResult: *run.Status,
			startTime:      *run.RunStart,
			endTime:        *run.RunEnd,
			durationMs:     *run.DurationInMs,
		}
		durations[i] = *run.DurationInMs

	}

	return runStats, durations

}
