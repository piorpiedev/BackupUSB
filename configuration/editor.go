package configuration

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/f1bonacc1/glippy"
	"github.com/rivo/tview"
)

const fieldWidth = 50

func clearKey(key string) string {
	return strings.Trim(strings.Trim(key, "\""), "'")
}

func clearPath(path string) string {
	path = strings.ReplaceAll(path, "'", "\"")
	path = strings.ReplaceAll(path, "\"", "") // Remove "
	return path
}

func splitPaths(paths string) []string {
	paths = strings.ReplaceAll(paths, "'", "\"")
	paths = strings.ReplaceAll(paths, "\" \"", "\",\"") // "a" "b" -> "a","b"
	paths = strings.ReplaceAll(paths, "\"", "")         // Remove "
	_paths := strings.Split(clearPath(paths), ",")

	// Filter list
	allKeys := make(map[string]bool)
	list := make([]string, 0, len(_paths))

	for _, item := range _paths {
		path := strings.Trim(item, " ")
		if path == "" {
			continue
		}

		// Don't add repeating paths
		if _, value := allKeys[path]; !value {
			allKeys[path] = true
			list = append(list, path)
		}
	}

	return list
}

func (c *Config) OpenEditor() {
	app := tview.NewApplication()
	form := tview.NewForm()

	// * FIELDS

	// Key:
	keyField := tview.NewInputField().
		SetLabel("Key:").
		SetFieldWidth(fieldWidth).
		SetText(c.Key)
	keyField.SetBlurFunc(func() {
		keyField.SetText(clearKey(keyField.GetText()))
	})
	form.AddFormItem(keyField)

	// Paths (comma separated):
	pathsField := tview.NewInputField().
		SetLabel("Paths (comma separated):").
		SetFieldWidth(fieldWidth).
		SetText("\"" + strings.Join(c.Paths, "\", \"") + "\"")
	pathsField.SetBlurFunc(func() { // Format string when unfocused
		pathsField.SetText("\"" + strings.Join(splitPaths(pathsField.GetText()), "\", \"") + "\"")
	})
	form.AddFormItem(pathsField)

	// Backups Amount:
	amountField := tview.NewInputField().
		SetLabel("Backups Amount:").
		SetFieldWidth(fieldWidth).
		SetText(strconv.Itoa(c.Amount))
	amountField.SetBlurFunc(func() {
		val := amountField.GetText()
		n, _ := strconv.Atoi(val) // Even if it errored (and it shouldn't since it's check with the acceptance func) it would just return n = 0
		amountField.SetText(strconv.Itoa(max(n, 1)))
	})
	amountField.SetAcceptanceFunc(func(textToCheck string, lastChar rune) bool {
		n, _ := strconv.Atoi(textToCheck)
		return n != 0 // err == nil can be removed since if it errors it would also return n = 0
	})
	form.AddFormItem(amountField)

	// Destination:
	destinationField := tview.NewInputField().
		SetLabel("Destination:").
		SetFieldWidth(fieldWidth).
		SetText("\"" + c.Destination + "\"")
	destinationField.SetBlurFunc(func() { // Format string when unfocused
		destinationField.SetText("\"" + strings.Trim(clearPath(destinationField.GetText()), " ") + "\"")
	})
	form.AddFormItem(destinationField)

	// * BUTTONS

	// Paste Key
	form.AddButton("Paste/Replace Key", func() {
		clipboard, _ := glippy.Get() // If err, clipboard will just be ""
		keyField.SetText(clipboard)
	})

	// Save
	form.AddButton("Save", func() {
		c.Key = clearKey(keyField.GetText())
		c.Paths = splitPaths(pathsField.GetText())

		val := amountField.GetText()
		n, err := strconv.Atoi(val)
		if err != nil {
			fmt.Printf("Invalid number: %s\n", val)
			return
		}
		c.Amount = max(n, 1) // At least one

		c.Destination = clearPath(destinationField.GetText())

		c.Save()
		app.Stop()
	})

	// Cancel
	form.AddButton("Cancel", func() {
		app.Stop()
	})

	// * SHOW
	form.SetBorder(true).SetTitle("Configuration").SetTitleAlign(tview.AlignLeft)
	app.SetRoot(form, true).EnableMouse(true).EnablePaste(true).Run()
}
