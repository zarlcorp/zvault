package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/zarlcorp/zvault/internal/task"
)

func runExport(args []string) {
	exportTasks := hasFlag(args, "--tasks")
	exportSecrets := hasFlag(args, "--secrets")
	pending := hasFlag(args, "--pending")
	done := hasFlag(args, "--done")

	// default: export everything
	if !exportTasks && !exportSecrets {
		exportTasks = true
		exportSecrets = true
	}

	v := openVault()
	defer v.Close()

	if exportSecrets {
		secrets, err := v.Secrets().List()
		if err != nil {
			errf("list secrets: %v", err)
			os.Exit(1)
		}

		fmt.Println("# Secrets")
		fmt.Println()
		if len(secrets) == 0 {
			fmt.Println("No secrets stored.")
		} else {
			fmt.Println("| ID | Type | Name | Tags |")
			fmt.Println("|---|---|---|---|")
			for _, sec := range secrets {
				tags := ""
				if len(sec.Tags) > 0 {
					var parts []string
					for _, t := range sec.Tags {
						parts = append(parts, "#"+t)
					}
					tags = strings.Join(parts, " ")
				}
				fmt.Printf("| %s | %s | %s | %s |\n", sec.ID, sec.Type, sec.Name, tags)
			}
		}
		fmt.Println()
	}

	if exportTasks {
		var f task.Filter
		if pending {
			f.Status = task.FilterPending
		} else if done {
			f.Status = task.FilterDone
		}

		tasks, err := v.Tasks().List(f)
		if err != nil {
			errf("list tasks: %v", err)
			os.Exit(1)
		}

		fmt.Println("# Tasks")
		fmt.Println()
		if len(tasks) == 0 {
			fmt.Println("No tasks found.")
		} else {
			for _, tk := range tasks {
				check := "[ ]"
				if tk.Done {
					check = "[x]"
				}

				extra := ""
				if tk.DueDate != nil {
					extra += "  due: " + tk.DueDate.Format("2006-01-02")
				}
				if len(tk.Tags) > 0 {
					var parts []string
					for _, t := range tk.Tags {
						parts = append(parts, "#"+t)
					}
					extra += "  " + strings.Join(parts, " ")
				}

				pri := ""
				switch tk.Priority {
				case task.PriorityHigh:
					pri = "!! "
				case task.PriorityMedium:
					pri = "! "
				}

				fmt.Printf("- %s %s%s%s\n", check, pri, tk.Title, extra)
			}
		}
		fmt.Println()
	}
}
