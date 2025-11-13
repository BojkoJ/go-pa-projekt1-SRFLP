package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
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

// load_data načte data ze souboru dataset2.txt
// Vrací: pole šířek zařízení a matici vah přechodů mezi zařízeními
func load_data() ([]int, [][]float64) {
	// Uložíme si cestu k souboru do proměnné
	var path string = "dataset2.txt"

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

// calculateDistance vypočítá vzdálenost mezi dvěma pozicemi v permutaci
// Vzorec POZMĚNĚNÝ DLE ZADÁNÍ: d(π_i, π_j) = (l_πi + l_πj)/2 + Σ(l_πk) pro i ≤ k ≤ j
// Toto je modifikovaný vzorec. Standardní SRFLP používá i < k < j
// Parametry:
//   - permutation: aktuální uspořádání zařízení - např. [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]
//   - widths: šířky jednotlivých zařízení - např. [4, 5, 3, 6, 2, 7, 4, 5, 3, 6]
//   - i: index první pozice v permutaci
//   - j: index druhé pozice v permutaci
//
// Vrací: vzdálenost mezi pozicemi i a j
func calculateDistance(permutation []int, widths []int, i int, j int) float64 {
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
	for k := i; k <= j; k++ { // normálně by to tedy balo k := i + 1; k < j; k++
		facilityK := permutation[k]
		// Přičteme jeho šířku k celkové vzdálenosti
		distance += float64(widths[facilityK])
	}

	return distance
}

// calculateCostIncrement vypočítá PŘÍRŮSTEK ceny při přidání nového zařízení
// Implementuje VNITŘNÍ část dvojitého vzorce:
// f_SRFLP = Σ(i) Σ(j>i) [c_π[i],π[j] * d(π[i], π[j])]
//
//	^
//	Pouze tato část pro pevné j=new
//
// Konkrétně počítá (pro všechna již umístěná zařízení):
// increment = Σ(i=0 to depth-1) [c_π[i],π[new] * d(π[i], π[new])]
//
// Poznámka: Vnější cyklus (přes všechna j) je v branchAndBound()
// Ten pak volá tuto funkci pro každé j, čímž se vytváří efektivní paralelní výpočet
//
// Toto je OPTIMALIZOVANÁ verze - místo přepočítání celé ceny (O(depth²))
// spočítáme jen přírůstek zavedený novým zařízením (O(depth))
//
// Parametry:
//   - perm []int: částečná permutace (jen prvních <depth> pozic je vyplněno)
//   - depth int: kolik pozic je už obsazeno
//   - newFacility int: zařízení, které chceme přidat na pozici depth
//   - widths []int: šířky všech zařízení
//   - costMatrix [][]float64: matice vah přechodů
//
// Vrací: přírůstek ceny (kolik se přičte k currentCost)
func calculateCostIncrement(perm []int, depth int, newFacility int,
	widths []int, costMatrix [][]float64) float64 {

	costIncrement := 0.0 // Proměnná pro přírůstek ceny

	// Dočasně přidáme nové zařízení na pozici depth pro výpočet vzdáleností
	perm[depth] = newFacility

	// Pro každé zařízení už v permutaci spočítáme cost s novým zařízením
	for i := 0; i < depth; i++ {
		facilityI := perm[i]

		// Použijeme modularizovanou funkci calculateDistance
		// Pro výpočet vzdálenosti mezi pozicí i a novou pozicí depth
		distance := calculateDistance(perm, widths, i, depth)

		// Přičteme cost: c_ij * distance
		costIncrement += costMatrix[facilityI][newFacility] * distance
	}

	return costIncrement
}

// atomicky: tak, že je to nepřerušitelná instrukce pro procesor
// Funkce dělá to, že dostane ukazatel na uint64 (reprezentující float64 hodnotu)
// a vrátí odpovídající float64 hodnotu atomicky.
func atomicLoadFloat64(addr *uint64) float64 {
	return math.Float64frombits(atomic.LoadUint64(addr))
}

// Funkce dělá to, že dostane ukazatel na uint64 (reprezentující float64 hodnotu)
// a uloží odpovídající float64 hodnotu do předané proměnné <val> atomicky
func atomicStoreFloat64(addr *uint64, val float64) {
	atomic.StoreUint64(addr, math.Float64bits(val))
}

// branchAndBound implementuje algoritmus Branch and Bound (backtracking)
//
// # OPTIMALIZOVANÁ verze
//
// Argumenty:
//   - perm []int - permutace, pole intů, předáno do funkce, pole je referenční typ, takže se touto funkcí bude přímo měnit dané pole
//   - used uint16 - bitmapa použitých zařízení (1 = použité, 0 = nepoužité), bude to vypadat třeba: 0000000000001011 (binárně)
//   - depth int - aktuální hloubka rekurze, počet již umístěných zařízení
//   - currentCost float64 - aktuální cena pro danou částečnou permutaci
//   - widths []int - šířky všech zařízení
//   - costMatrix [][]float64 - matice vah přechodů, určuje náklady mezi zařízeními
//   - n int - počet zařízení
//   - localBest *float64 - ukazatel na nejlepší lokální cenu (pro pruning)
//   - visited *int64 - ukazatel na počet navštívených uzlů (pro statistiky)
//   - pruned *int64 - ukazatel na počet ořezaných uzlů (pro statistiky)
func branchAndBound(perm []int, used uint16, depth int, currentCost float64,
	widths []int, costMatrix [][]float64, n int,
	localBest *float64, visited, pruned *int64) {

	// branchAndBound se spouští rekurzivně, takže je třeba inkrementovat počet navštívených uzlů
	*visited++

	// BOUND - rychlý check, že aktuální cena už není lepší než nejlepší nalezená
	if currentCost >= *localBest {
		// Pokud prošlo, tak aktuální cena je větší než globální nejlepší cena
		// Takže ořežeme tento uzel, protože nemá smysl pokračovat dál
		*pruned++
		return
	}

	// Kompletní permutace
	if depth == n {
		// Pokud jsme lepší, update atomicky
		for {
			current := atomicLoadFloat64(&bestCost)
			if currentCost >= current {
				break
			}
			if atomic.CompareAndSwapUint64(&bestCost,
				math.Float64bits(current),
				math.Float64bits(currentCost)) {
				*localBest = currentCost
				bestMutex.Lock()
				bestPermutation = make([]int, n)
				copy(bestPermutation, perm[:n])
				bestMutex.Unlock()
				// Odstraněn fmt.Printf - zpomaluje výkon!
				break
			}
		}
		return
	}

	// Občas refresh (bitwise AND je rychlejší než modulo)
	if *visited&0xFFF == 0 { // každých 4096 uzlů (méně často = rychlejší)
		global := atomicLoadFloat64(&bestCost)
		if global < *localBest {
			*localBest = global
		}
	}

	// BRANCH - zkusíme všechna dosud nepoužitá zařízení
	for facility := 0; facility < n; facility++ {
		// Rychlý check bitmapy - je toto zařízení už použité?
		if used&(1<<facility) != 0 {
			continue
		}

		// Vypočítáme přírůstek ceny při přidání tohoto zařízení
		// Používáme optimalizovanou funkci - O(depth) místo O(depth²)
		costIncrement := calculateCostIncrement(perm, depth, facility, widths, costMatrix)
		newCost := currentCost + costIncrement

		// Pruning před rekurzí
		if newCost >= *localBest {
			*pruned++
			continue
		}

		// In-place update
		perm[depth] = facility

		// Rekurze
		branchAndBound(perm, used|(1<<facility), depth+1, newCost,
			widths, costMatrix, n, localBest, visited, pruned)
	}
}

func main() {
	// Spustíme měření času
	startTime := time.Now()

	fmt.Println("=== SRFLP Solver s Branch and Bound ===")
	fmt.Println()

	// Načteme data ze souboru
	fmt.Println("Načítám data ze souboru dataset2.txt...")
	widths, matrix := load_data()

	// Ověříme, že se data načetla správně
	if widths == nil || matrix == nil {
		fmt.Println("Chyba při načítání dat!")
		return
	}

	n := len(widths)
	fmt.Printf("Počet zařízení: %d\n", n)
	fmt.Printf("Šířky zařízení: %v\n", widths)
	fmt.Println()

	fmt.Println("Spouštím Branch and Bound algoritmus...")
	fmt.Println("(Paralelní výpočet - optimalizováno pro rychlost)")
	fmt.Println()

	// Inicializace
	// OPTIMALIZACE: Nejprve najdeme greedy řešení jako počáteční upper bound
	// Toto dramaticky zvýší pruning hned od začátku!

	// Spočítáme celkovou váhu každého zařízení (součet všech costs)
	type facilityWeight struct {
		id     int
		weight float64
	}
	facilities := make([]facilityWeight, n)
	for i := 0; i < n; i++ {
		weight := 0.0
		for j := 0; j < n; j++ {
			if i != j {
				weight += matrix[i][j]
			}
		}
		facilities[i] = facilityWeight{id: i, weight: weight}
	}

	// Seřadíme podle váhy (nejvyšší první)
	sort.Slice(facilities, func(i, j int) bool {
		return facilities[i].weight > facilities[j].weight
	})

	greedyPerm := make([]int, n)
	greedyUsed := make([]bool, n)
	greedyPerm[0] = facilities[0].id // Začneme s nejvyšší váhou
	greedyUsed[greedyPerm[0]] = true
	greedyCost := 0.0

	// Greedy: vždy přidáme zařízení s nejnižším přírůstkem
	for depth := 1; depth < n; depth++ {
		bestFacility := -1
		bestIncrement := math.Inf(1)

		for _, fw := range facilities {
			if greedyUsed[fw.id] {
				continue
			}

			// Spočítáme přírůstek pomocí modifikovaného vzorce (i ≤ k ≤ j)
			increment := 0.0
			for i := 0; i < depth; i++ {
				facilityI := greedyPerm[i]
				// První část: (l_i + l_new) / 2
				distance := float64(widths[facilityI]+widths[fw.id]) / 2.0
				// Druhá část: suma VŠECH šířek včetně krajních (i ≤ k ≤ depth)
				for k := i; k <= depth; k++ {
					if k < depth {
						distance += float64(widths[greedyPerm[k]])
					} else {
						distance += float64(widths[fw.id])
					}
				}
				increment += matrix[facilityI][fw.id] * distance
			}

			if increment < bestIncrement {
				bestIncrement = increment
				bestFacility = fw.id
			}
		}

		greedyPerm[depth] = bestFacility
		greedyUsed[bestFacility] = true
		greedyCost += bestIncrement
	}

	// Nastavíme greedy řešení jako počáteční upper bound
	atomicStoreFloat64(&bestCost, greedyCost)
	bestPermutation = make([]int, n)
	copy(bestPermutation, greedyPerm)
	fmt.Printf("Greedy iniciální řešení: cena = %.2f, permutace = %v\n", greedyCost, greedyPerm)

	totalVisited.Store(0)
	totalPruned.Store(0)

	// PARALELIZACE - každé vlákno má své lokální počítadla
	var wg sync.WaitGroup

	// Kanál pro agregaci lokálních statistik
	type stats struct {
		visited int64
		pruned  int64
	}
	statsChan := make(chan stats, n)

	// Spustíme n goroutines - každá začne s jiným zařízením jako první pozici
	for i := 0; i < n; i++ {
		wg.Add(1)

		go func(startIdx int) {
			defer wg.Done()

			start := facilities[startIdx].id // Použijeme seřazené ID z greedy

			// Lokální proměnné - žádná synchronizace!
			localBest := math.Inf(1)
			var visited, pruned int64

			// Buffer pro permutaci (pre-alokovaný)
			perm := make([]int, n)
			perm[0] = start
			used := uint16(1 << start)

			// Spustíme backtracking
			branchAndBound(perm, used, 1, 0.0, widths, matrix, n,
				&localBest, &visited, &pruned)

			// Na konci pošleme statistiky
			statsChan <- stats{visited, pruned}
		}(i)
	} // Počkáme na všechny goroutines
	wg.Wait()
	close(statsChan)

	// Agregujeme statistiky
	for s := range statsChan {
		totalVisited.Add(s.visited)
		totalPruned.Add(s.pruned)
	}

	// Výpočet skončil, zastavíme měření času
	elapsed := time.Since(startTime)

	// Vypíšeme výsledky
	fmt.Println()
	fmt.Println("=== VÝSLEDKY ===")
	finalCost := atomicLoadFloat64(&bestCost)
	fmt.Printf("Nejlepší nalezená cena: %.2f\n", finalCost)
	fmt.Printf("Nejlepší permutace: %v\n", bestPermutation)
	fmt.Println()

	// Pro lepší čitelnost přidáme i permutaci s indexy od 1
	permutationPlusOne := make([]int, len(bestPermutation))
	for i, v := range bestPermutation {
		permutationPlusOne[i] = v + 1
	}
	fmt.Printf("Permutace (indexy 1-%d): %v\n", n, permutationPlusOne)
	fmt.Println()

	// Statistiky
	fmt.Println("=== STATISTIKY ===")
	fmt.Printf("Čas výpočtu: %v\n", elapsed)
	fmt.Printf("Navštívených uzlů: %d\n", totalVisited.Load())
	fmt.Printf("Ořezaných větví: %d\n", totalPruned.Load())
	fmt.Printf("Počet goroutines: %d\n", n)
	fmt.Println()

}
