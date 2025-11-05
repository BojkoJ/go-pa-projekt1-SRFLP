package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func load_data() {
	// Uložíme si cestu k souboru do proměnné
	var path string = "Y-t_10.txt"

	// Z knihovny os funkce Open otevře soubor
	file, err := os.Open(path)
	// Ošetření, když nastane error (err)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	// Defer dělá to, že soubor se zavře až na konci funkce
	defer file.Close()

	// Z knihovny io funkce ReadAll načte celý obsah souboru do proměnné data
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Převedeme data na string a rozdělíme na řádky
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	// Řádek 1 obsahuje jen jedno číslo a pak newline character
	// Nelze použít numberOfItems = int(lines[0]) protože int() konvertuje pouze numerické typy, ne stringy
	numberOfItems, err := strconv.Atoi(lines[0])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Řádek 2 obsahuje šířky zařízení oddělené mezerami (a pak newline character)
	var widths []int
	widthStrings := strings.Split(strings.TrimSpace(lines[1]), " ")
	for _, wstr := range widthStrings {
		w, err := strconv.Atoi(wstr)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		widths = append(widths, w)
	}
}

func main() {
	fmt.Println("Hello, World!")
}
