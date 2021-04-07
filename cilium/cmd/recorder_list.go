// Copyright 2021 Authors of Cilium
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/command"

	"github.com/spf13/cobra"
)

// recorderListCmd represents the recorder_list command
var recorderListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List current pcap recorders",
	Run: func(cmd *cobra.Command, args []string) {
		listRecorders(cmd, args)
	},
}

func init() {
	recorderCmd.AddCommand(recorderListCmd)
	command.AddJSONOutput(recorderListCmd)
}

func listRecorders(cmd *cobra.Command, args []string) {
	list, err := client.GetRecorder()
	if err != nil {
		Fatalf("Cannot get recorder list: %s", err)
	}

	if command.OutputJSON() {
		if err := command.PrintOutput(list); err != nil {
			os.Exit(1)
		}
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 5, 0, 3, ' ', 0)
	printRecorderList(w, list)
}

func printRecorderList(w *tabwriter.Writer, recList []*models.Recorder) {
	fmt.Fprintln(w, "ID\tCapture Length\tWildcard Filters\t")
	for _, rec := range recList {
		if rec.Status == nil || rec.Status.Realized == nil {
			fmt.Fprint(os.Stderr, "error parsing recorder: empty state")
			continue
		}
	}
	sort.Slice(recList, func(i, j int) bool {
		return *recList[i].Status.Realized.ID <= *recList[j].Status.Realized.ID
	})
	for _, rec := range recList {
		spec := rec.Status.Realized
		capLen := "full"
		if spec.CaptureLength != 0 {
			capLen = fmt.Sprintf("<= %d", spec.CaptureLength)
		}
		str := fmt.Sprintf("%d\t%s\t%s:%s\t->\t%s:%s\t%s",
			int64(*spec.ID), capLen,
			spec.Filters[0].SrcPrefix, spec.Filters[0].SrcPort,
			spec.Filters[0].DstPrefix, spec.Filters[0].DstPort,
			spec.Filters[0].Protocol)
		fmt.Fprintln(w, str)
		for _, filter := range spec.Filters[1:] {
			str := fmt.Sprintf("\t\t%s:%s\t->\t%s:%s\t%s",
				filter.SrcPrefix, filter.SrcPort,
				filter.DstPrefix, filter.DstPort,
				filter.Protocol)
			fmt.Fprintln(w, str)
		}
	}
	w.Flush()
}
