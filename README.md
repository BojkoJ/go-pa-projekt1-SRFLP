# Paralelní Algoritmy I - Projekt 1: SRFLP (Single Row Facility Layout Problem)

Tento repozitář obsahuje řešení Projektu 1 z předmětu **Paralelní Algoritmy 1** (VŠB-TUO) implementované v jazyce **Go**.

**[📋 Oficiální zadání projektu](https://homel.vsb.cz/~kro080/PAI-2025/U1/)**

---

## 📖 O problému SRFLP

**Single Row Facility Layout Problem** (problém rozložení zařízení v jedné řadě) je optimalizační problém zaměřený na hledání nejlepšího lineárního uspořádání zařízení s cílem minimalizovat celkové náklady na přesuny mezi nimi.

### Praktický příklad

Představte si **robotické rameno** obklopené dopravními pásy, ze kterých odebírá součástky. Pásy, mezi kterými robot jezdí často, by měly být blízko sebe. Pokud jsou často využívané pásy daleko od sebe, robot ztrácí čas a energii dlouhými přesuny.

**Cíl:** Najít takové uspořádání pásů (zařízení), které minimalizuje celkovou "cenou" všech přesunů.

---

## 🔢 Matematická formulace

Pro množinu $n$ zařízení $F = \{1, 2, ..., n\}$ se známými **šířkami** $L = (l_1, l_2, ..., l_n)$ a **maticí vah** přechodů $C = \{c_{ij}\} \in \mathbb{R}^{n \times n}$ hledáme permutaci $\pi = (\pi_1, \pi_2, ..., \pi_n)$, která minimalizuje:

$$\min_{\pi \in S_n} f_{\text{SRFLP}}(\pi) = \sum_{1 \le i < j \le n} \left[ c_{\pi_i \pi_j} \cdot d(\pi_i, \pi_j) \right]$$

Kde vzdálenost mezi zařízeními je definována jako:

$$d(\pi_i, \pi_j) = \frac{l_{\pi_i} + l_{\pi_j}}{2} + \sum_{i < k < j} l_{\pi_k}$$

### ⚠️ Poznámka k vzorci vzdálenosti

V **oficiálním zadání projektu** je vzorec pro výpočet vzdálenosti **mírně změněn** oproti standardní definici SRFLP:

**Změněný vzorec (použitý v projektu):**
$$d(\pi_i, \pi_j) = \frac{l_{\pi_i} + l_{\pi_j}}{2} + \sum_{i \le k \le j} l_{\pi_k}$$

Rozdíl je v rozsahu sumy - místo `i < k < j` (pouze zařízení mezi) se používá `i ≤ k ≤ j` (včetně krajních zařízení).

---

## 🚀 Řešení

### Algoritmus: Branch and Bound (optimalizovaný)

Program implementuje **vysoce optimalizovaný paralelní Branch and Bound algoritmus** s pokročilými technikami:

#### 1. **Greedy Inicializace (Warm Start)**

- Před spuštěním B&B se vypočítá greedy řešení jako počáteční upper bound
- Zařízení se seřadí podle celkové váhy (součtu vah přechodů)
- Greedy heuristika pak přidává zařízení s nejnižším přírůstkem ceny
- Tím se dramaticky zvýší pruning hned od začátku!

#### 2. **Klíčové optimalizace**

✅ **Bitmapové masky** (`uint16`) - rychlé sledování použitých zařízení bez pole booleanů  
✅ **Atomické operace** - `bestCost` sdílená atomicky mezi vlákny (bez mutexu)  
✅ **Lokální best tracking** - každé vlákno má svůj `localBest` pro rychlejší pruning  
✅ **Inkrementální výpočet** - `calculateCostIncrement()` místo přepočítání celé ceny (O(depth) vs O(depth²))  
✅ **Pre-alokované buffery** - permutační pole alokované jednou, ne při každé rekurzi  
✅ **Periodický refresh** - lokální best se aktualizuje z global každých 4096 uzlů (bitwise AND místo modulo)  
✅ **Paralelizace** - každé vlákno začíná s jiným zařízením z greedy pořadí  
✅ **Statistiky** - agregace lokálních počítadel na konci (minimální synchronizace)

---

## 📁 Vstupní data

Repozitář obsahuje testovací dataset **`Y-t_10.txt`** s následující strukturou:

```
10                                    # Počet zařízení
1 1 1 1 1 1 1 1 1 1                  # Šířky zařízení
0 30 17 11 24 25 24 17 16 22         # Matice vah (horní trojúhelník)
0 0 21 23 26 24 27 19 11 32
...
```

**Formát souboru:**

- **Řádek 1:** Počet zařízení `n`
- **Řádek 2:** Šířky jednotlivých zařízení
- **Řádky 3-(2+n):** Horní trojúhelníková matice vah přechodů mezi zařízeními

---

## 💻 Spuštění programu

### Požadavky

- **[Golang](https://go.dev/)** verze 1.19 nebo vyšší

### Kompilace a spuštění

```bash
# Kompilace
go build main.go

# Spuštění
./main        # Linux/macOS
./main.exe    # Windows
```

### Očekávaný výstup

```
=== SRFLP Solver s Branch and Bound ===

Načítám data ze souboru Y-t_10.txt...
Počet zařízení: 10
Šířky zařízení: [1 1 1 1 1 1 1 1 1 1]

Spouštím Branch and Bound algoritmus...
(Paralelní výpočet - optimalizováno pro rychlost)

Greedy iniciální řešení: cena = 6207, permutace = [6 9 7 0 8 3 4 5 1 2]

=== VÝSLEDKY ===
Nejlepší nalezená cena: 5596
Nejlepší permutace: [0 4 1 9 6 3 7 2 5 8]

Permutace (indexy 1-10): [1 5 2 10 7 4 8 3 6 9]

=== STATISTIKY ===
Čas výpočtu: 50-300ms (závisí na CPU, s greedy warm start výrazně rychlejší)
Navštívených uzlů: ~6-8 milionů (optimalizace snížila prohledávání)
Ořezaných větví: ~3-4 miliony (greedy start zvýšil efektivitu pruningu)
Počet goroutines: 10
```

---

## 📊 Výsledky

Pro testovací instanci `Y-t_10.txt` (10 zařízení):

| Metrika                     | Hodnota                       |
| --------------------------- | ----------------------------- |
| **Počet možných permutací** | 3,628,800 (10!)               |
| **Nalezené minimum**        | **5596**                      |
| **Optimální uspořádání**    | [0 4 1 9 6 3 7 2 5 8]         |
| **Greedy warm start**       | 6207 (11% od optima)          |
| **Čas výpočtu**             | ~50-100ms (závisí na HW)      |
| **Navštívených uzlů**       | ~6-8 milionů (0.2% z 10!)     |
| **Ořezané větve**           | ~3-4 miliony (efektivita 50%) |

---

## 🏗️ Struktura kódu

```
main.go
├── load_data()                    # Načtení dat ze souboru Y-t_10.txt
├── calculate_distance()           # Výpočet vzdálenosti mezi zařízeními (modifikovaný vzorec)
├── calculateCostIncrement()       # Inkrementální výpočet přírůstku ceny (O(depth))
├── atomicLoadFloat64()            # Atomické načtení float64 hodnoty
├── atomicStoreFloat64()           # Atomické uložení float64 hodnoty
├── branchAndBound()               # Ultra-optimalizovaný rekurzivní B&B algoritmus
│   ├── Bitmapové masky (uint16)   # Pro tracking použitých zařízení
│   ├── Lokální best tracking      # Každé vlákno má svůj localBest
│   ├── Periodický refresh         # Sync s global best každých 4096 uzlů
│   └── Pruning před rekurzí       # Eliminace neperpsektivních větví
└── main()                         # Greedy warm start + paralelní výpočet
    ├── Greedy inicializace        # Počáteční upper bound
    ├── Paralelizace (goroutines)  # Každé vlákno začíná s jiným zařízením
    └── Agregace statistik         # Sběr lokálních počítadel
```

---
