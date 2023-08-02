package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
)

type Flashcard struct {
	Front      string `json:"card"`
	Back       string `json:"definition"`
	ErrorCount int    `json:"errorCount"`
}

var ioLog strings.Builder

func readInputString() string {
	reader := bufio.NewReader(os.Stdin)

	text, _ := reader.ReadString('\n')
	text = strings.TrimSpace(text)
	ioLog.WriteString(text + "\n")
	return text
}

func createFlashcard(existingFlashcards []Flashcard) Flashcard {
	var flashcardFront = ""
	var flashcardBack = ""

	for !validateFlashcard("term", existingFlashcards, flashcardFront) {
		if flashcardFront != "" {
			printAndLog("The card \"%s\" already exists. Try again:\n", flashcardFront)
		} else {
			printAndLog("The card:\n")
		}
		flashcardFront = readInputString()
	}

	for !validateFlashcard("definition", existingFlashcards, flashcardBack) {
		if flashcardBack != "" {
			printAndLog("The definition \"%s\" already exists. Try again:\n", flashcardBack)
		} else {
			printAndLog("The definition of the card:\n")
		}
		flashcardBack = readInputString()
	}

	return Flashcard{Front: flashcardFront, Back: flashcardBack}
}

func validateFlashcard(option string, existingFlashcards []Flashcard, input string) bool {
	if input == "" {
		return false
	}

	for _, flashcard := range existingFlashcards {
		if (option == "term" && flashcard.Front == input) || (option == "definition" && flashcard.Back == input) {
			return false
		}
	}

	return true
}

func findCardForTheDefinition(existingFlashcards []Flashcard, currentFlashcardIndex int, providedDefinition string) Flashcard {
	for i, flashcard := range existingFlashcards {
		if i == currentFlashcardIndex {
			continue
		}

		if flashcard.Back == providedDefinition {
			return flashcard
		}
	}

	return existingFlashcards[currentFlashcardIndex]
}

func answerFlashcard(flashcard Flashcard, existingFlashcards *[]Flashcard, flashcardIndex int) {
	printAndLog("Print the definition of \"%s\":\n", flashcard.Front)
	answer := readInputString()

	if flashcard.Back == answer {
		printAndLog("Correct!\n\n")
	} else {
		flashcardForDefinition := findCardForTheDefinition(*existingFlashcards, flashcardIndex, answer)
		if flashcardForDefinition == flashcard {
			printAndLog("Wrong. The right answer is \"%s\".\n", flashcard.Back)
		} else {
			printAndLog("Wrong. The right answer is \"%s\", but your definition is correct for \"%s\".\n", flashcard.Back, flashcardForDefinition.Front)
		}
		(*existingFlashcards)[flashcardIndex].ErrorCount += 1
	}
}

func addFlashcard(flashcardsSetPointer *[]Flashcard) {
	newFlashcard := createFlashcard(*flashcardsSetPointer)
	*flashcardsSetPointer = append(*flashcardsSetPointer, newFlashcard)
	printAndLog("The pair (\"%s\":\"%s\") has been added.\n\n", newFlashcard.Front, newFlashcard.Back)
}

func removeFlashcard(flashcardsSetPointer *[]Flashcard) {
	var flashcardIndex = -1
	var cardToRemove string

	printAndLog("Which card?\n")
	cardToRemove = readInputString()

	for i, flashcard := range *flashcardsSetPointer {
		if flashcard.Front == cardToRemove {
			flashcardIndex = i
			break
		}
	}

	if flashcardIndex == -1 {
		printAndLog("Can't remove \"%s\": there is no such card.\n\n", cardToRemove)
	} else {
		*flashcardsSetPointer = append((*flashcardsSetPointer)[:flashcardIndex], (*flashcardsSetPointer)[flashcardIndex+1:]...)
		printAndLog("The card has been removed.\n\n")
	}
}

func importFlashcards(flashcardsSetPointer *[]Flashcard, fileToImport string) {
	var filename = fileToImport
	var importedFlashcards []Flashcard
	var indexisToSkip = make([]int, 0, len(*flashcardsSetPointer))
	if filename == "" {
		printAndLog("File name:\n")
		filename = readInputString()
	}

	fileContent, err := os.ReadFile(filename)
	if err != nil {
		printAndLog("File not found.\n\n")
	} else {
		json.Unmarshal(fileContent, &importedFlashcards)
		for existingIndex, existingFlashcard := range *flashcardsSetPointer {
			for newIndex, newFlashcard := range importedFlashcards {
				if existingFlashcard.Front == newFlashcard.Front || existingFlashcard.Back == newFlashcard.Back {
					flashcardSliceTail := (*flashcardsSetPointer)[existingIndex+1:]
					*flashcardsSetPointer = append((*flashcardsSetPointer)[:existingIndex], newFlashcard)
					*flashcardsSetPointer = append(*flashcardsSetPointer, flashcardSliceTail...)
					indexisToSkip = append(indexisToSkip, newIndex)
					break
				}
			}
		}

		for i, flashcard := range importedFlashcards {
			skip := false
			for _, indexToSkip := range indexisToSkip {
				if i == indexToSkip {
					skip = true
					break
				}
			}
			if skip {
				continue
			}
			*flashcardsSetPointer = append(*flashcardsSetPointer, flashcard)
		}
		printAndLog("%d cards have been loaded.\n\n", len(importedFlashcards))
	}
}

func exportFlashcards(flashcardsSetPointer *[]Flashcard, exportFile string) {
	var filename = exportFile
	var serializedFlashcards []byte
	if filename == "" {
		printAndLog("File name:\n")
		filename = readInputString()
	}

	serializedFlashcards, _ = json.Marshal(*flashcardsSetPointer)
	os.WriteFile(filename, serializedFlashcards, 0644)

	printAndLog("%d cards have been saved.\n\n", len(*flashcardsSetPointer))
}

func askQuestions(flashcardsSetPointer *[]Flashcard) {
	if len(*flashcardsSetPointer) < 1 {
		printAndLog("No cards available!\n")
	} else {
		var numberOfCardsToAskString string
		printAndLog("How many times to ask?\n")
		numberOfCardsToAskString = readInputString()
		numberOfCardsToAsk, _ := strconv.Atoi(numberOfCardsToAskString)
		askedQuestions := 0
		for askedQuestions < numberOfCardsToAsk {
			for i, flashcard := range *flashcardsSetPointer {
				if askedQuestions >= numberOfCardsToAsk {
					break
				}
				answerFlashcard(flashcard, flashcardsSetPointer, i)
				askedQuestions++
			}
		}
	}
}

func printAndLog(text string, arguments ...any) {
	formattedText := fmt.Sprintf(text, arguments...)
	ioLog.WriteString(formattedText)
	fmt.Printf(formattedText)
}

func saveLog() {
	var filename string
	printAndLog("File name:\n")
	filename = readInputString()

	os.WriteFile(filename, []byte(ioLog.String()), 0644)
	printAndLog("The log has been saved.\n\n")
}

func evaluateHardestCard(flashcardsSetPointer *[]Flashcard) {
	hardestFlashcards := make([]Flashcard, 0)
	topMistakesCount := 0

	for _, flashcard := range *flashcardsSetPointer {
		if flashcard.ErrorCount == topMistakesCount && topMistakesCount > 0 {
			hardestFlashcards = append(hardestFlashcards, flashcard)
		} else if flashcard.ErrorCount > topMistakesCount {
			hardestFlashcards = []Flashcard{flashcard}
			topMistakesCount = flashcard.ErrorCount
		}
	}

	hardestFlashcardsLength := len(hardestFlashcards)

	if hardestFlashcardsLength == 0 {
		printAndLog("There are no cards with errors.\n\n")
	} else if hardestFlashcardsLength == 1 {
		flashcard := hardestFlashcards[0]
		printAndLog("The hardest card is \"%s\". You have %d errors answering it.\n\n", flashcard.Front, flashcard.ErrorCount)
	} else {
		var output strings.Builder
		output.WriteString("The hardest cards are ")
		for i, flashcard := range hardestFlashcards {
			output.WriteString(fmt.Sprintf("\"%s\"", flashcard.Front))
			if i != hardestFlashcardsLength-1 {
				output.WriteString(", ")
			} else {
				output.WriteString(fmt.Sprintf(". You have %d errors answering them.\n\n", flashcard.ErrorCount))
			}
		}

		printAndLog(output.String())
	}
}

func resetStatistics(flashcardsSetPointer *[]Flashcard) {
	for i, _ := range *flashcardsSetPointer {
		(*flashcardsSetPointer)[i].ErrorCount = 0
	}

	printAndLog("Card statistics have been reset.\n\n")
}

func main() {
	initialImportFile := flag.String("import_from", "", "Enter filename for initial cards import")
	exitExportFile := flag.String("export_to", "", "Enter filename for exporting cards on exit")
	flag.Parse()

	var action string
	flashcards := make([]Flashcard, 0)

	log.Println(*initialImportFile)

	if *initialImportFile != "" {
		importFlashcards(&flashcards, *initialImportFile)
	}

	for action != "exit" {
		printAndLog("Input the action (add, remove, import, export, ask, exit, log, hardest card, reset stats):\n")
		action = readInputString()

		switch action {
		case "add":
			addFlashcard(&flashcards)
		case "remove":
			removeFlashcard(&flashcards)
		case "import":
			importFlashcards(&flashcards, "")
		case "export":
			exportFlashcards(&flashcards, "")
		case "ask":
			askQuestions(&flashcards)
		case "log":
			saveLog()
		case "hardest card":
			evaluateHardestCard(&flashcards)
		case "reset stats":
			resetStatistics(&flashcards)
		case "exit":
			if *exitExportFile != "" {
				exportFlashcards(&flashcards, *exitExportFile)
			}
			printAndLog("Bye bye!\n")
		}
	}
}
