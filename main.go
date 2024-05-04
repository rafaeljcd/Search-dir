package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"os/exec"
	"strconv"
	"runtime"
	"github.com/fatih/color"
	"bufio"
)

type Config struct {
	Index []string `json:"index"`
}

func main() {
	execDirPath, err := directoryCall()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	addFlag := flag.String("add", "", "Add a directory path to the search index")
	removeFlag := flag.String("remove", "", "Remove a directory path from the search index")

	flag.Parse()
	if *addFlag != "" {
		addSearchPath(execDirPath, *addFlag)
		return
	}
	if *removeFlag != "" {
		removeSearchPath(execDirPath, *removeFlag)
		return
	}

	searchDirPathList, err := readJsonFile(execDirPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if len(searchDirPathList) == 0 {
		fmt.Println("No search directory path found in the config file")
		return
	}

	fmt.Println("Search Directory Paths:")
	c1 := color.New(color.FgCyan)
	c2 := color.New(color.FgGreen)
	for i, path := range searchDirPathList {
		c1.Printf("%d. ", i+1)
		c2.Println(path)
	}

	entryList := fetchEntryList(searchDirPathList)

	searchForEntries(entryList)

}

func directoryCall() (string, error) {
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("Error:", err)
		return "", err
	}

	dirPath := filepath.Dir(exePath)

	return dirPath, nil
}

func readJsonFile(dirPath string) ([]string, error) {
	configPath := dirPath + "/config.json"
	// check if the file exists
	_, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		fmt.Println("Config file does not exist")
		createIfFileNotExist(configPath)
		addSearchPath(dirPath, dirPath)
		return []string{dirPath}, nil
	} else if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}

	// read the config.json file
	file, err := os.Open(configPath)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	defer file.Close()

	byteValue, _ := io.ReadAll(file)

	var config Config
	json.Unmarshal(byteValue, &config)

	return config.Index, nil
}

func createIfFileNotExist(configFilePath string) {
	data := Config{
		Index: []string{},
	}

	// Marshal the data into a JSON Format
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write the data to the file
	err = os.WriteFile(configFilePath, jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

func addSearchPath(dirPath string, searchPath string) {
	fileInfo, err := os.Stat(searchPath)
	if os.IsNotExist(err) {
		fmt.Printf("Path %s does not exist\n", searchPath)
		return
	} else if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if !fileInfo.IsDir() {
		fmt.Printf("Path %s is not a directory\n", searchPath)
		return
	}

	configPath := dirPath + "/config.json"
	searchDirPathList, err := readJsonFile(dirPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	searchDirPathList = append(searchDirPathList, searchPath)

	data := Config{
		Index: searchDirPathList,
	}

	// Marshal the data into a JSON Format
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write the data to the file
	err = os.WriteFile(configPath, jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Search Path %s added successfully\n", searchPath)
}

func removeSearchPath(dirPath string, searchPath string) {
	fileInfo, err := os.Stat(searchPath)
	if os.IsNotExist(err) {
		fmt.Printf("Path %s does not exist\n", searchPath)
		return
	} else if err != nil {
		fmt.Println("Error:", err)
		return
	}

	if !fileInfo.IsDir() {
		fmt.Printf("Path %s is not a directory\n", searchPath)
		return
	}

	configPath := dirPath + "/config.json"
	searchDirPathList, err := readJsonFile(dirPath)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	searchDirPathList = removeString(searchDirPathList, searchPath)

	data := Config{
		Index: searchDirPathList,
	}

	// Marshal the data into a JSON Format
	jsonData, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write the data to the file
	err = os.WriteFile(configPath, jsonData, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Search Path %s removed successfully\n", searchPath)
}

func removeString(slice []string, target string) []string {
	result := []string{}
	for _, value := range slice {
		if value != target {
			result = append(result, value)
		}
	}
	return result
}

func fetchEntryList(dirList []string) []string {

	entryList := []string{}
	for _, dir := range dirList {
		_, err := os.Stat(dir)
		if os.IsNotExist(err) {
			fmt.Printf("Path %s does not exist\n", dir)
			continue
		} else if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}

		for _, file := range files {
			if file.IsDir() {
				dir_path := filepath.Join(dir, file.Name())
				entryList = append(entryList, dir_path)
			}
		}
	}
	return entryList
}

func searchForEntries(entryList []string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the search query or type 'exit' to exit: ")

		query, _ := reader.ReadString('\n')
		query = strings.TrimSpace(query)

		if query == "exit" || query == "q" {
			break
		}

		searchResultList := []string{}
		for _, entry := range entryList {
			entry_path := filepath.Base(entry)
			if strings.Contains(strings.ToLower(entry_path), strings.ToLower(query)) {
				searchResultList = append(searchResultList, entry)
			}
		}

		if len(searchResultList) == 0 {
			fmt.Println("No search result found")
			continue
		}

		for i, entry := range searchResultList {

			c1 := color.New(color.FgCyan)

			c1.Printf("%d. ", i+1)
			print_search_result(entry, query)
		}

		ChooseDirectoryToOpen(searchResultList)

		clearTerminal()


	}
	clearTerminal()
}

func print_search_result(result string, query string) {
	entry_path := filepath.Base(result)

	c1 := color.New(color.FgYellow)
	c2 := color.New(color.BgRed)

	lowerEntryPath := strings.ToLower(entry_path)
	lowerQuery := strings.ToLower(query)

	index := strings.Index(lowerEntryPath, lowerQuery)

	beforeSearchTerm := entry_path[:index]
    searchTerm := entry_path[index : index+len(query)]
    afterSearchTerm := entry_path[index+len(query):]

	c1.Print(beforeSearchTerm)
	c2.Print(searchTerm)
	c1.Println(afterSearchTerm)
}

func clearTerminal() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		fmt.Println("Error clearing screen:", err)
	}
}

func ChooseDirectoryToOpen(searchResultList []string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter the number of the directory to open or type 'exit' to exit: ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		if choice == "exit" || choice == "q" {
			break
		}

		index, err := strconv.Atoi(choice)
		if err != nil {
			fmt.Println("Invalid input")
			continue
		}

		if index < 1 || index > len(searchResultList) {
			fmt.Println("Invalid input")
			continue
		}

		file := searchResultList[index-1]

		entry_path := filepath.Base(file)
		c2 := color.New(color.FgYellow)

		fmt.Print("Opening directory:")
		c2.Printf(" %s\n", entry_path)

		if runtime.GOOS == "windows" {
			cmd := exec.Command("explorer", file)
			err = cmd.Start()
		} else {
			cmd := exec.Command("open", file)
			err = cmd.Start()
		}
		if err != nil {
			fmt.Println("Error:", err)
		}
	}
}