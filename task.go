package main

import (
	"flag"
	"fmt"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/segmentio/go-prompt"
	"github.com/thedevsaddam/task/taskmanager"
	"os"
	"runtime"
	"strconv"
	"strings"
)

const USAGE = `Usage:
	Name:
		Task manager
	Description:
		Your favourite task list or todo manager
	Version:
		1.0.0
	$ task
		Show all tasks
	$ task p
		Show all pending tasks
	$ task a Watch Games of thrones
		Add a new task [Watch Games of thrones] to list
	$ task del
		Remove latest task from list
	$ task rm ID
		Remove task of ID from list
	$ task s ID
		Show detail view task of ID
	$ task c ID
		Mark task of ID as completed
	$ task m ID Pirates of the Caribbean
		Modify a task
	$ task p ID
		Mark task of ID as pending
	$ task flush
		Flush the database!
`

const (
	COMPLETED_MARK = "\u2713"
	PENDING_MARK   = "\u2613"
)

//task manager instance
var tm = taskmanager.New()

func main() {

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, USAGE)
		flag.PrintDefaults()
	}
	flag.Parse()
	cmd, args, argsLen := flag.Arg(0), flag.Args(), len(flag.Args())

	switch {
	case cmd == "" || cmd == "l" || cmd == "ls" && argsLen == 1:
		showTasksInTable(tm.GetAllTasks())
	case cmd == "a" || cmd == "add" && argsLen >= 1:
		tm.Add(strings.Join(args[1:], " "), "")
		successText(" Added to list: " + strings.Join(args[1:], " ") + " ")
	case cmd == "p" || cmd == "pending" && argsLen == 1:
		showTasksInTable(tm.GetPendingTasks())
	case cmd == "del" || cmd == "delete" && argsLen == 1:
		p := prompt.Choose("Do you want to delete latest task?", []string{"yes", "no"})
		if p == 1 {
			warningText(" Task delete aboarted! ")
			return
		}
		err := tm.RemoveTask(tm.GetLastId())
		if err != nil {
			errorText(err.Error())
			return
		}
		successText(" Removed latest task ")
	case cmd == "r" || cmd == "rm" && argsLen == 2:
		id, _ := strconv.Atoi(flag.Arg(1))
		p := prompt.Choose("Do you want to delete task of id "+flag.Arg(1)+" ?", []string{"yes", "no"})
		if p == 1 {
			warningText(" Task delete aboarted! ")
			return
		}
		err := tm.RemoveTask(id)
		if err != nil {
			errorText(err.Error())
			return
		}
		successText(" Task " + strconv.Itoa(id) + " removed! ")
	case cmd == "e" || cmd == "m" || cmd == "u" && argsLen >= 2:
		id, _ := strconv.Atoi(flag.Arg(1))
		ok, _ := tm.UpdateTask(id, strings.Join(args[2:], " "))
		successText(ok)
	case cmd == "c" || cmd == "d" || cmd == "done" && argsLen >= 2:
		id, _ := strconv.Atoi(flag.Arg(1))
		task, err := tm.MarkAsCompleteTask(id)
		if err != nil {
			errorText(err.Error())
			return
		}
		successText(" " + COMPLETED_MARK + " " + task.Description)
	case cmd == "i" || cmd == "p" || cmd == "pending" && argsLen >= 2:
		id, _ := strconv.Atoi(flag.Arg(1))
		task, err := tm.MarkAsPendingTask(id)
		if err != nil {
			errorText(err.Error())
			return
		}
		successText(" " + pendingMark() + " " + task.Description)
	case cmd == "s" && argsLen == 2:
		id, _ := strconv.Atoi(flag.Arg(1))
		task, err := tm.GetTask(id)
		if err != nil {
			errorText(err.Error())
			return
		}
		showTask(task)
	case cmd == "flush":
		p := prompt.Choose("Do you want to delete all tasks?", []string{"yes", "no"})
		if p == 1 {
			warningText(" Flush aborted! ")
			return
		}
		err := tm.FlushDB()
		if err != nil {
			errorText(err.Error())
			return
		}
		successText(" Database flushed successfully! ")
	case cmd == "h" || cmd == "v":
		fmt.Fprint(os.Stderr, USAGE)
	default:
		errorText(" [No command found by " + cmd + "] ")
		fmt.Fprint(os.Stderr, "\n"+USAGE)
	}

}

//show tasks list in table
func showTasksInTable(tasks taskmanager.Tasks) {
	fmt.Fprintln(os.Stdout, "")
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Description", COMPLETED_MARK + "/" + pendingMark(), "Created"})
	table.SetBorders(tablewriter.Border{Left: true, Top: true, Right: true, Bottom: false})
	table.SetFooter([]string{"", "Total: " + strconv.Itoa(tm.TotalTask()), "", "Pending: " + strconv.Itoa(tm.PendingTask())})
	table.SetCenterSeparator("|")
	table.SetRowLine(true)
	for _, task := range tasks {
		//set completed icon
		status := PENDING_MARK
		if task.Completed != "" {
			status = COMPLETED_MARK
		} else {
			status = pendingMark()
		}
		table.Append([]string{
			strconv.Itoa(task.Id),
			task.Description,
			status,
			task.Created,
		})
	}
	table.Render()
	fmt.Fprintln(os.Stdout, "")
}

//show a single tasks
func showTask(task taskmanager.Task) {
	fmt.Fprintln(os.Stdout, "")
	printText("Task Details view")
	printText("--------------------------------")
	printText("ID: " + strconv.Itoa(task.Id))
	printText("UID: " + task.UID)
	printText("Description: " + task.Description)
	printText("Tag: " + task.Tag)
	printText("Created: " + task.Created)
	printText("Updated: " + task.Updated)
	fmt.Fprintln(os.Stdout, "")
}

func printText(str string) {
	fmt.Fprintf(os.Stdout, str+"\n")
}

func printBoldText(str string) {
	if runtime.GOOS == "windows" {
		fmt.Fprintf(os.Stdout, str+"\n")
	} else {
		bold := color.New(color.Bold).FprintlnFunc()
		bold(os.Stdout, str)
	}
}

func successText(str string) {
	if runtime.GOOS == "windows" {
		fmt.Fprintf(color.Output, color.GreenString(str))
	} else {
		success := color.New(color.Bold, color.BgGreen, color.FgWhite).FprintlnFunc()
		success(os.Stdout, str)
	}
}

func warningText(str string) {
	if runtime.GOOS == "windows" {
		fmt.Fprintf(color.Output, color.YellowString(str))
	} else {
		warning := color.New(color.Bold, color.BgYellow, color.FgBlack).FprintlnFunc()
		warning(os.Stdout, str)
	}
}

func errorText(str string) {
	if runtime.GOOS == "windows" {
		fmt.Fprintf(color.Output, color.RedString(str))
	} else {
		error_ := color.New(color.Bold, color.BgRed, color.FgWhite).FprintlnFunc()
		error_(os.Stdout, str)
	}
}

func pendingMark() string {
	pending := PENDING_MARK
	if runtime.GOOS == "windows" {
		pending = "x"
	}
	return pending
}
