package main

import (
	"fmt"
	"os"
	"unicode"

	"github.com/asdine/storm/v3"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	flag "github.com/spf13/pflag"

	"github.com/ajaxray/geek-life/model"
	"github.com/ajaxray/geek-life/repository"
	repo "github.com/ajaxray/geek-life/repository/storm"
	"github.com/ajaxray/geek-life/util"
	"github.com/ajaxray/geek-life/config"
)

var (
	app              *tview.Application
	layout, contents, taskview *tview.Flex

	statusBar         *StatusBar
	projectPane       *ProjectPane
	taskPane          *TaskPane
	taskDetailPane    *TaskDetailPane
	projectDetailPane *ProjectDetailPane

	db          *storm.DB
	projectRepo repository.ProjectRepository
	taskRepo    repository.TaskRepository

	// Flag variables
	dbFile string
    vertical bool
    dynamiclist bool
)

func init() {
	flag.StringVarP(&dbFile, "db-file", "d", "", "Specify DB file path manually.")
    flag.BoolVarP(&vertical, "vertical", "v", false, "Vertical task detail layout.")
    flag.BoolVarP(&dynamiclist, "dynamic", "D", false, "Enables the dynamic list")
}

func main() {
	app = tview.NewApplication()
	flag.Parse()

    config.Init(flag.CommandLine)
    config.SaveConfig()
    dbFile = config.GetDbFile()
    vertical = config.GetVertical()
    dynamiclist = config.GetDynamic()

	db = util.ConnectStorm(dbFile)
	defer func() {
		if err := db.Close(); err != nil {
			util.LogIfError(err, "Error in closing storm Db")
		}
	}()

	if flag.NArg() > 0 && flag.Arg(0) == "migrate" {
		migrate(db)
		fmt.Println("Database migrated successfully!")
	} else {
		projectRepo = repo.NewProjectRepository(db)
		taskRepo = repo.NewTaskRepository(db)

		layout = tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(makeTitleBar(), 2, 1, false).
			AddItem(prepareContentPages(), 0, 2, true).
			AddItem(prepareStatusBar(app), 1, 1, false)

		setKeyboardShortcuts()

		if err := app.SetRoot(layout, true).EnableMouse(true).Run(); err != nil {
			panic(err)
		}
	}

}

func migrate(database *storm.DB) {
	util.FatalIfError(database.ReIndex(&model.Project{}), "Error in migrating Projects")
	util.FatalIfError(database.ReIndex(&model.Task{}), "Error in migrating Tasks")

	fmt.Println("Migration completed. Start geek-life normally.")
	os.Exit(0)
}

func setKeyboardShortcuts() *tview.Application {
	return app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if ignoreKeyEvt() {
			return event
		}

		// Global shortcuts
		switch unicode.ToLower(event.Rune()) {
		case 'p':
			app.SetFocus(projectPane)
			taskview.RemoveItem(taskDetailPane)
			return nil
		case 'q':
		case 't':
			app.SetFocus(taskPane)
			taskview.RemoveItem(taskDetailPane)
			return nil
		}

		// Handle based on current focus. Handlers may modify event
		switch {
		case projectPane.HasFocus():
			event = projectPane.handleShortcuts(event)
		case taskPane.HasFocus():
			event = taskPane.handleShortcuts(event)
			if event != nil && projectDetailPane.isShowing() {
				event = projectDetailPane.handleShortcuts(event)
			}
		case taskDetailPane.HasFocus():
			event = taskDetailPane.handleShortcuts(event)
		}

		return event
	})
}

func prepareContentPages() *tview.Flex {
	projectPane = NewProjectPane(projectRepo)
	taskPane = NewTaskPane(projectRepo, taskRepo)
	projectDetailPane = NewProjectDetailPane()
	taskDetailPane = NewTaskDetailPane(taskRepo)

    taskview = tview.NewFlex()
    if(vertical) {
        taskview.SetDirection(tview.FlexRow)
    }
    taskview.AddItem(taskPane, 0, 1, false)

	contents = tview.NewFlex().
		AddItem(projectPane, 25, 1, true).
		AddItem(taskview, 0, 2, false)

	return contents

}

func makeTitleBar() *tview.Flex {
	titleText := tview.NewTextView().SetText("[lime::b]Geek-life [::-]- Task Manager for geeks!").SetDynamicColors(true)
	versionInfo := tview.NewTextView().SetText("[::d]Version: 0.1.2b").SetTextAlign(tview.AlignRight).SetDynamicColors(true)

	return tview.NewFlex().
		AddItem(titleText, 0, 2, false).
		AddItem(versionInfo, 0, 1, false)
}

func AskYesNo(text string, f func()) {

	activePane := app.GetFocus()
	modal := tview.NewModal().
		SetText(text).
		AddButtons([]string{"Yes", "No"}).
		SetDoneFunc(func(buttonIndex int, buttonLabel string) {
			if buttonLabel == "Yes" {
				f()
			}
			app.SetRoot(layout, true).EnableMouse(true)
			app.SetFocus(activePane)
		})

	pages := tview.NewPages().
		AddPage("background", layout, true, true).
		AddPage("modal", modal, true, true)
	_ = app.SetRoot(pages, true).EnableMouse(true)
}

func InputPopup(title string, f func(input string)) {
	activePane := app.GetFocus()
	var name string

	form := tview.NewForm().
		AddInputField("", "", 40, nil, func(text string) {
			name = text
		})
	form.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyEnter:
			f(name)
		case tcell.KeyEsc:
		default:
			return event
		}

		app.SetRoot(layout, true).EnableMouse(true)
		app.SetFocus(activePane)

		return event
	}).SetBorder(true).SetTitle(title)
	modal := tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(form, 5, 1, true).
			AddItem(nil, 0, 1, false), 40, 1, true).
		AddItem(nil, 0, 1, false)

	pages := tview.NewPages().
		AddPage("background", layout, true, true).
		AddPage("modal", modal, true, true)
	_ = app.SetRoot(pages, true).EnableMouse(true)
}
