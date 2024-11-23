package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	color          = "white"
	inputNameStage = true
)

func main() {
	serverAddress := "localhost:8080"
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Printf("Error connecting to server: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	app := tview.NewApplication()

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	textView.SetBorder(true).SetTitle("Messages")
	var inputField *tview.InputField
	inputField = tview.NewInputField().
		SetLabel("You: ").
		SetDoneFunc(func(key tcell.Key) {
			message := strings.TrimSpace(inputField.GetText())
			if message == "" {
				return
			}

			if colorInput, found := strings.CutPrefix(message, "!color "); found {
				color = colorInput
				inputField.SetText("")
				return
			}

			if strings.HasPrefix(message, "!") || inputNameStage {
				fmt.Fprintf(conn, "%s\n", message)
				inputNameStage = false
			} else {
				fmt.Fprintf(conn, "[%s]%s[white]\n", color, message)
			}

			inputField.SetText("")
		})

	inputField.SetBorder(true).SetTitle("Type your message")

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(inputField, 3, 1, true)

	fmt.Fprint(textView, "[grey]Please enter your username[white]\n")

	go func() {
		reader := bufio.NewScanner(conn)
		for reader.Scan() {
			message := reader.Text()
			app.QueueUpdateDraw(func() {
				if msg, found := strings.CutPrefix(message, "[INFO] "); found {
					message = fmt.Sprintf("[grey]%s[white]", msg)
				}
				fmt.Fprintln(textView, message) // Safely update TextView
			})
		}

		if err := reader.Err(); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintln(textView, "[red]Disconnected from server.")
			})
		} else {
			app.QueueUpdateDraw(func() {
				fmt.Fprintln(textView, "[red]Server closed the connection.")
			})
		}

		app.Stop()
	}()

	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Printf("Error running application: %s\n", err)
	}
}
