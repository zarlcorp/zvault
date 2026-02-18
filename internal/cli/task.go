package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/zarlcorp/zvault/internal/task"
)

func runTask(args []string) {
	if len(args) == 0 {
		printTaskUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "add":
		runTaskAdd(args[1:])
	case "list", "ls":
		runTaskList(args[1:])
	case "done":
		runTaskDone(args[1:])
	case "edit":
		runTaskEdit(args[1:])
	case "rm":
		runTaskRm(args[1:])
	case "clear":
		runTaskClear()
	case "help", "--help", "-h":
		printTaskUsage()
	default:
		// bare ID: show task detail
		runTaskDetail(args[0])
	}
}

func printTaskUsage() {
	fmt.Fprint(os.Stderr, `Usage: zvault task <command>

Commands:
  add     add a new task
  list    list tasks (alias: ls)
  done    mark task(s) complete
  edit    rename a task
  rm      delete task(s)
  clear   delete all completed tasks
  <id>    show task detail

Add flags:
  -p <h|m|l>        priority (high, medium, low)
  -d <date>          due date (YYYY-MM-DD, tomorrow, next week, +3d)
  --tags tag1,tag2   optional tags

List flags:
  --pending          show only pending tasks
  --done             show only completed tasks
  -p <h|m|l>        filter by priority
  --tag <tag>        filter by tag
`)
}

func runTaskAdd(args []string) {
	pri := flagValue(args, "-p")
	dueStr := flagValue(args, "-d")
	tags := parseTags(flagValue(args, "--tags"))

	pos := stripFlags(args, []string{"-p", "-d", "--tags"}, nil)
	if len(pos) == 0 {
		errf("task title required")
		os.Exit(1)
	}

	title := strings.Join(pos, " ")
	tk := task.New(title)
	tk.Tags = tags

	if pri != "" {
		p, ok := parsePriority(pri)
		if !ok {
			errf("invalid priority %q (use h, m, or l)", pri)
			os.Exit(1)
		}
		tk.Priority = p
	}

	if dueStr != "" {
		due, err := parseDate(dueStr)
		if err != nil {
			errf("%v", err)
			os.Exit(1)
		}
		tk.DueDate = &due
	}

	v := openVault()
	defer v.Close()

	if err := v.Tasks().Add(tk); err != nil {
		errf("add task: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", green(tk.ID), title)
}

func runTaskList(args []string) {
	var f task.Filter

	if hasFlag(args, "--pending") {
		f.Status = task.FilterPending
	} else if hasFlag(args, "--done") {
		f.Status = task.FilterDone
	}

	pri := flagValue(args, "-p")
	if pri != "" {
		p, ok := parsePriority(pri)
		if !ok {
			errf("invalid priority %q (use h, m, or l)", pri)
			os.Exit(1)
		}
		f.Priority = p
	}

	f.Tag = flagValue(args, "--tag")

	v := openVault()
	defer v.Close()

	tasks, err := v.Tasks().List(f)
	if err != nil {
		errf("list tasks: %v", err)
		os.Exit(1)
	}

	if len(tasks) == 0 {
		fmt.Fprintln(os.Stderr, muted("no tasks found"))
		return
	}

	for _, tk := range tasks {
		printTaskRow(tk)
	}
}

func runTaskDone(args []string) {
	if len(args) == 0 {
		errf("task ID(s) required")
		os.Exit(1)
	}

	ids := parseIDs(args[0])

	v := openVault()
	defer v.Close()

	now := time.Now()
	for _, id := range ids {
		tk, err := v.Tasks().Get(id)
		if err != nil {
			errf("task %q not found", id)
			os.Exit(1)
		}

		tk.Done = true
		tk.CompletedAt = &now

		if err := v.Tasks().Update(tk); err != nil {
			errf("update task: %v", err)
			os.Exit(1)
		}

		fmt.Printf("%s %s\n", green("[x]"), tk.Title)
	}
}

func runTaskEdit(args []string) {
	if len(args) < 2 {
		errf("usage: zvault task edit <id> <new title>")
		os.Exit(1)
	}

	id := args[0]
	title := strings.Join(args[1:], " ")

	v := openVault()
	defer v.Close()

	tk, err := v.Tasks().Get(id)
	if err != nil {
		errf("task %q not found", id)
		os.Exit(1)
	}

	tk.Title = title

	if err := v.Tasks().Update(tk); err != nil {
		errf("update task: %v", err)
		os.Exit(1)
	}

	fmt.Printf("%s %s\n", muted(tk.ID), title)
}

func runTaskRm(args []string) {
	if len(args) == 0 {
		errf("task ID(s) required")
		os.Exit(1)
	}

	ids := parseIDs(args[0])

	v := openVault()
	defer v.Close()

	for _, id := range ids {
		if err := v.Tasks().Delete(id); err != nil {
			errf("delete task %q: %v", id, err)
			os.Exit(1)
		}
		fmt.Printf("%s deleted\n", muted(id))
	}
}

func runTaskClear() {
	v := openVault()
	defer v.Close()

	count, err := v.Tasks().ClearDone()
	if err != nil {
		errf("clear done: %v", err)
		os.Exit(1)
	}

	if count == 0 {
		fmt.Fprintln(os.Stderr, muted("no completed tasks to clear"))
		return
	}

	fmt.Printf("cleared %d completed task(s)\n", count)
}

func runTaskDetail(id string) {
	v := openVault()
	defer v.Close()

	tk, err := v.Tasks().Get(id)
	if err != nil {
		errf("task %q not found", id)
		os.Exit(1)
	}

	printTaskDetail(tk)
}

func printTaskRow(tk task.Task) {
	check := "[ ]"
	if tk.Done {
		check = green("[x]")
	}

	priStr := "   "
	switch tk.Priority {
	case task.PriorityHigh:
		priStr = boldRed("!!")
	case task.PriorityMedium:
		priStr = boldYellow(" !")
	case task.PriorityLow:
		priStr = muted(" .")
	}

	due := ""
	if tk.DueDate != nil {
		due = "  due: " + formatDueDate(tk.DueDate)
	}

	tags := ""
	if len(tk.Tags) > 0 {
		var parts []string
		for _, t := range tk.Tags {
			parts = append(parts, blue("#"+t))
		}
		tags = "  " + strings.Join(parts, " ")
	}

	title := tk.Title
	if tk.Done {
		title = muted(tk.Title)
	}

	fmt.Printf("%s %s %s %s%s%s\n", check, muted(tk.ID), priStr, title, due, tags)
}

func printTaskDetail(tk task.Task) {
	status := green("pending")
	if tk.Done {
		status = muted("done")
	}

	fmt.Printf("%s %s\n", bold(tk.Title), muted(tk.ID))
	fmt.Printf("  %s %s\n", muted("status:"), status)

	if tk.Priority != task.PriorityNone {
		var priColor func(string) string
		switch tk.Priority {
		case task.PriorityHigh:
			priColor = red
		case task.PriorityMedium:
			priColor = yellow
		default:
			priColor = muted
		}
		fmt.Printf("  %s %s\n", muted("priority:"), priColor(string(tk.Priority)))
	}

	if tk.DueDate != nil {
		fmt.Printf("  %s %s\n", muted("due:"), formatDueDate(tk.DueDate))
	}

	if len(tk.Tags) > 0 {
		var parts []string
		for _, t := range tk.Tags {
			parts = append(parts, blue("#"+t))
		}
		fmt.Printf("  %s %s\n", muted("tags:"), strings.Join(parts, " "))
	}

	fmt.Printf("  %s %s\n", muted("created:"), tk.CreatedAt.Format("2006-01-02 15:04"))

	if tk.CompletedAt != nil {
		fmt.Printf("  %s %s\n", muted("completed:"), tk.CompletedAt.Format("2006-01-02 15:04"))
	}
}

// parsePriority converts shorthand (h, m, l) or full names to Priority.
func parsePriority(s string) (task.Priority, bool) {
	switch strings.ToLower(s) {
	case "h", "high":
		return task.PriorityHigh, true
	case "m", "medium":
		return task.PriorityMedium, true
	case "l", "low":
		return task.PriorityLow, true
	default:
		return task.PriorityNone, false
	}
}
