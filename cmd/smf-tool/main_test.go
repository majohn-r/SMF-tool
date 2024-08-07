package main

import (
	"testing"

	"github.com/majohn-r/output"
)

func Test_main(t *testing.T) {
	savedExitFunc := exitFunc
	savedFirstYear := firstYear
	savedBus := bus
	defer func() {
		exitFunc = savedExitFunc
		firstYear = savedFirstYear
		bus = savedBus
	}()
	tests := map[string]struct {
		firstYear    string
		execFunc     func(output.Bus, int, string, string, string, []string) int
		wantExitCode int
		output.WantedRecording
	}{
		"bad first year": {
			firstYear:       "",
			wantExitCode:    1,
			WantedRecording: output.WantedRecording{Error: "The value of firstYear \"\" is not valid: strconv.Atoi: parsing \"\": invalid syntax.\n"},
		},
		"failure": {
			firstYear: "2021",
			execFunc: func(_ output.Bus, _ int, _ string, _ string, _ string, _ []string) int {
				return 1
			},
			wantExitCode: 1,
		},
		"success": {
			firstYear: "2021",
			execFunc: func(_ output.Bus, _ int, _ string, _ string, _ string, _ []string) int {
				return 0
			},
			wantExitCode: 0,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			firstYear = tt.firstYear
			exitFunc = func(exitCode int) {
			}
			o := output.NewRecorder()
			bus = o
			main()
		})
	}
}
