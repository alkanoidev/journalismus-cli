package entryUtils

import (
	"os"
	"time"
)

func WriteEntryToFile(body string) (string, error) {
	filename := "./" + time.Now().Format("January-02-2006") + ".md"
	err := os.WriteFile(filename, []byte(body), 0644)
	if err != nil {
		return "An error occured", err
	}
	return "Entry saved", err
}

func ReadOrCreateEntry() (string, error) {
	filename := "./" + time.Now().Format("January-02-2006") + ".md"
	entry, err := os.ReadFile(filename)
	if err == nil {
		return string(entry), nil
	} else {
		// create todays entry if it doesnt exist
		f, err := os.Create(filename)
		f.Close()
		return "", err
	}

}

func DeleteEntry() error {
	filename := "./" + time.Now().Format("January-02-2006") + ".md"
	err := os.Remove(filename)
	return err
}
