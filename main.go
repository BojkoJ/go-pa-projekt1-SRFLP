package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

// Globální proměnné - minimální synchronizace
var (
	// bestCost - atomická proměnná pro nejlepší cenu (používáme uint64 pro atomic operace s float64)
	bestCost uint64
	// bestPermutation - mutex chráněná
	bestPermutation []int
	// Mutex (Mutual Exclusion) pro ochranu bestPermutation
	// Je to vlastně obdoba zámku (lock)
	// Když jedna goroutine drží mutex, ostatní musí čekat, až ho uvolní
	bestMutex sync.Mutex
	// Statistiky
	// Proč atomic.Int64 - Go 1.19+ podporuje atomic typy přímo
	// Je to hardware atomic typ - používá hardware atomické instrukce
	// To znamená, že je garantováno, že operace proběhne celá bez přerušení - menší šance deadlocku
	// Je dobré to používat, když máme více goroutines a chceme minimalizovat zámky (mutexy)
	totalVisited atomic.Int64
	totalPruned  atomic.Int64
)

// load_data načte data ze souboru Y-t_10.txt
// Vrací: pole šířek zařízení a matici vah přechodů mezi zařízeními
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
			// ParseFloat vrací float64 (64-bitový float) - druhý parametr 64 specifikuje přesnost
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

	// DŮLEŽITÉ: Doplnění dolní části matice (Dataset dolní část pod diagonálou má hodnotu 0)
	// Matice musí být symetrická - c_ij = c_ji
	n := len(matrix)
	for i := 0; i < n; i++ {
		for j := i + 1; j < n; j++ {
			// Zkopírujeme hodnotu z horní části (nad diagonálou) do dolní části (pod diagonálou)
			matrix[j][i] = matrix[i][j]
		}
	}

	return widths, matrix
}

// calculate_distance vypočítá vzdálenost mezi dvěma pozicemi v permutaci
// Vzorec POZMĚNĚNÝ DLE ZADÁNÍ: d(π_i, π_j) = (l_πi + l_πj)/2 + Σ(l_πk) pro i ≤ k ≤ j
// POZOR: Toto je modifikovaný vzorec! Standardní SRFLP používá i < k < j
// Parametry:
//   - permutation: aktuální uspořádání zařízení
//   - widths: šířky jednotlivých zařízení
//   - i: index první pozice v permutaci
//   - j: index druhé pozice v permutaci
//
// Vrací: vzdálenost mezi pozicemi i a j
func calculate_distance(permutation []int, widths []int, i int, j int) float64 {
	// Zjistíme, která zařízení jsou na pozicích i a j
	// permutation[i] je INDEX zařízení (0-9), ne jeho pozice
	facilityI := permutation[i]
	facilityJ := permutation[j]

	// První část vzorce: polovina šířky obou krajních zařízení
	// (l_πi + l_πj) / 2
	distance := float64(widths[facilityI]+widths[facilityJ]) / 2.0

	// Druhá část vzorce: suma šířek VŠECH zařízení včetně krajních (i ≤ k ≤ j)
	// Σ(l_πk)
	// !!! Toto je vzorec ze zadání. Správně by to mělo být i < k < j, ale zadání říká včetně krajních. !!!
	for k := i; k <= j; k++ {
		facilityK := permutation[k]
		// Přičteme jeho šířku k celkové vzdálenosti
		distance += float64(widths[facilityK])
	}

	return distance
}

// calculateCostIncrement vypočítá PŘÍRŮSTEK ceny při přidání nového zařízení
// Toto je OPTIMALIZOVANÁ verze - místo přepočítání celé ceny (O(depth²))
// spočítáme jen zvýšení ceny způsobené novým zařízením (O(depth))
// Parametry:
//   - perm: částečná permutace (jen prvních depth pozic je vyplněno)
//   - depth: kolik pozic je už obsazeno
//   - newFacility: zařízení, které chceme přidat na pozici depth
//   - widths: šířky všech zařízení
//   - costMatrix: matice vah přechodů
//
// Vrací: přírůstek ceny (kolik se přičte k currentCost)
func calculateCostIncrement(perm []int, depth int, newFacility int,
	widths []int, costMatrix [][]float64) float64 {

	costIncrement := 0.0

	// Dočasně přidáme nové zařízení na pozici depth pro výpočet vzdáleností
	perm[depth] = newFacility

	// Pro každé zařízení už v permutaci spočítáme cost s novým zařízením
	for i := 0; i < depth; i++ {
		facilityI := perm[i]

		// Použijeme modularizovanou funkci calculate_distance
		// Pro výpočet vzdálenosti mezi pozicí i a novou pozicí depth
		distance := calculate_distance(perm, widths, i, depth)

		// Přičteme cost: c_ij * distance
		costIncrement += costMatrix[facilityI][newFacility] * distance
	}

	return costIncrement
}

func main() {
	// Načteme data
	widths, costMatrix := load_data()
	if widths == nil || costMatrix == nil {
		fmt.Println("Failed to load data.")
		return
	}

	// výpis dat
	fmt.Println("Widths:", widths)
	fmt.Println("Cost Matrix:")
	for _, row := range costMatrix {
		fmt.Println(row)
	}

}
