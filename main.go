package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func load_data() ([]int, [][]float64) {
	// Uložíme si cestu k souboru do proměnné
	var path string = "Y-t_10.txt"

	// Z knihovny os funkce Open otevře soubor
	file, err := os.Open(path)
	// Ošetření, když nastane error (err)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}
	// Defer dělá to, že soubor se zavře až na konci funkce
	defer file.Close()

	// Z knihovny io funkce ReadAll načte celý obsah souboru do proměnné data
	data, err := io.ReadAll(file)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}

	// Převedeme data na string, normalizujeme konce řádků a rozdělíme na řádky
	content := strings.TrimSpace(string(data))
	// Normalizovat Windows/Mac/*nix konce řádků na '\n'
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")

	// Řádek 1 obsahuje jen jedno číslo a pak newline character
	// Nelze použít numberOfItems = int(lines[0]) protože int() konvertuje pouze numerické typy, ne stringy
	numberOfItems, err := strconv.Atoi(lines[0]) // Atoi je ekvivalent ParseInt pro převod stringu na int
	if err != nil {
		fmt.Println("Error:", err)
		return nil, nil
	}

	// Řádek 2 obsahuje šířky zařízení oddělené mezerami (a pak newline character)
	var widths []int
	// Funkce Split - jako první argument bere string, který chceme rozdělit, jako druhý argument bere oddělovač
	// Funkce TrimSpace odstraní whitespace znaky (jako jsou newlines) z začátku a konce stringu
	widthStrings := strings.Split(strings.TrimSpace(lines[1]), " ")
	
	// Takže teď máme ve widthStrings pole stringů, které musíme převést na pole intů:
	for _, wstr := range widthStrings { // Pro každý string v poli widthStrings: (první je _, protože index nepotřebujeme)
		w, err := strconv.Atoi(wstr) // Atoi je ekvivalent ParseInt pro převod stringu na int
		if err != nil {
			fmt.Println("Error:", err)
			return nil, nil
		}
		
		widths = append(widths, w) // Do pole widths přidáme převedené číslo
	}
	
	// Teď máme ve widths pole intů obsahující šířky zařízení, widthStrings je už nepotřebné - takže ho můžeme "smazat"
	widthStrings = nil

	// Řádky 3+: horní trojuhelníková matice vah:
	// Do temp proměnné si uložíme jen ty řádky, které obsahují váhy
	// 2: - od začátku (index 0) přeskočíme první dva řádky
	// (2 + numberOfItems): - až po řádek, který obsahuje poslední váhu (počítáme od nuly, takže to je 2 + numberOfItems)
	var matrix [][]float64
	tmp := lines[2:(2 + numberOfItems)]

	// Pro každý řádek v tmp:
	for _, line := range tmp {
		// Vytvoříme si prázdné pole floatů pro jeden řádek matice
		row := []float64{}
		// Rozdělíme řádek na části podle mezer
		parts := strings.Split(line, " ")

		// Pro každou část:
		for _, part := range parts {
			// Převedeme string na float
			value, err := strconv.ParseFloat(part, 64)
			if err != nil {
				fmt.Println("Error:", err)
				return nil, nil
			}
			// Převedenou hodnotu přidáme do row (pole floatů)
			row = append(row, value)
		}
		// Řádek row máme plný, přidáme ho do matice
		matrix = append(matrix, row)
	}

	// Důležité: Doplnění dolní části matice (Dataset dolní část pod diagonálou má hodnotu 0)
	// Matice musí být symetrická
	n := len(matrix)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			matrix[j][i] = matrix[i][j]
		}
	}

	return widths, matrix
}

func main() {
	// načteme data a logneme je abychom si ověřili, že se načetla správně
	widths, matrix := load_data()
	fmt.Println("Widths:", widths)
	fmt.Println("Matrix:")
	for _, row := range matrix {
		fmt.Println(row)
	}
}
