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

func main() {
	serverAddress := "localhost:8080" // Change to your server's address

	// Connect to the chat server
	conn, err := net.Dial("tcp", serverAddress)
	if err != nil {
		fmt.Printf("Error connecting to server: %s\n", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Setup tview application
	app := tview.NewApplication()

	// TextView for displaying messages
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw() // Ensure updates are reflected immediately
		})

	textView.SetBorder(true).SetTitle("Messages")

	// InputField for typing messages
	var inputField *tview.InputField // Declare as a pointer for proper scoping
	inputField = tview.NewInputField().
		SetLabel("You: ").
		SetDoneFunc(func(key tcell.Key) {
			message := strings.TrimSpace(inputField.GetText())
			if message == "" {
				return
			}

			// Send the message to the server
			fmt.Fprintf(conn, "%s\n", message)
			inputField.SetText("") // Clear the input field
		})

	inputField.SetBorder(true).SetTitle("Type your message")

	// Layout the UI
	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false). // Messages take all available space
		AddItem(inputField, 3, 1, true) // Input field has fixed height

	fmt.Fprint(textView, "[grey]Please enter your username[white]\n")

	// Goroutine to handle incoming messages
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

		// Handle disconnection
		if err := reader.Err(); err != nil {
			app.QueueUpdateDraw(func() {
				fmt.Fprintln(textView, "[red]Disconnected from server.")
			})
		} else {
			app.QueueUpdateDraw(func() {
				fmt.Fprintln(textView, "[red]Server closed the connection.")
			})
		}

		// Stop the app after disconnection
		app.Stop()
	}()

	// Start the app
	if err := app.SetRoot(flex, true).Run(); err != nil {
		fmt.Printf("Error running application: %s\n", err)
	}
}
