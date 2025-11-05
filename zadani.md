# Zadání projektu: Paralelní řešení problému rozložení zařízení (SRFLP)

## Úvod

Tento projekt se zabývá implementací paralelního algoritmu pro řešení **Single-Row Facility Layout Problem** (SRFLP), česky "problém rozložení zařízení v jedné řadě". Jde o optimalizační problém, který má široké využití v průmyslu, logistice a automatizaci.

---

## Co je SRFLP? - Vysvětlení

### Základní myšlenka

Představte si, že máte **robotické rameno** uprostřed výrobní haly a kolem něj je několik **dopravních pásů** (nebo jiných zařízení), ze kterých robot odebírá součástky. Robot se musí neustále pohybovat mezi těmito pásy, aby vybíral díly a skládal z nich výrobky.

**Problém:** V jakém pořadí (uspořádání) rozmístíme pásy kolem robota, aby robot jezdil co nejméně a byl co nejefektivnější?

### Praktický příklad ze života

Máte 4 pásy označené jako **m₁, m₂, m₃, m₄**:

-   Z pásu **m₁** robot odebírá díly velmi často (např. 100× za hodinu)
-   Z pásu **m₂** jen občas (10× za hodinu)
-   Z pásu **m₃** často (80× za hodinu)
-   Z pásu **m₄** velmi často (90× za hodinu)

Pokud bychom umístili pásy, mezi kterými robot jezdí často, daleko od sebe (např. m₁ na jeden konec a m₄ na druhý konec), robot by musel stále projíždět dlouhé vzdálenosti. To je **neefektivní** - spotřebuje více energie, opotřebuje se rychleji a výroba je pomalejší.

**Lepší řešení:** Umístit často využívané pásy (m₁, m₃, m₄) blízko sebe a méně využívaný pás (m₂) stranou.

---

## Matematická formulace problému

Nyní si problém popíšeme přesně pomocí matematiky.

### Vstupní data

1. **Množina zařízení:**

    $$F = \{1, 2, 3, ..., n\}$$

    **Vysvětlení:** Máme $n$ zařízení (pásů). Každé má své číslo od 1 do $n$. Například pokud máme 10 pásů, pak $n = 10$ a $F = \{1, 2, 3, 4, 5, 6, 7, 8, 9, 10\}$.

2. **Šířky zařízení:**

    $$L = (l_1, l_2, l_3, ..., l_n)$$

    **Vysvětlení:** Každé zařízení má svou šířku. Symbol $l_i$ znamená "šířka zařízení číslo $i$".

    - $l_1$ = šířka prvního zařízení
    - $l_2$ = šířka druhého zařízení
    - atd.

    **Příklad:** Pokud máme 3 pásy se šířkami 2 metry, 1.5 metru a 3 metry, pak:
    $$L = (2, 1.5, 3)$$

3. **Matice vah přechodů:**

    $$C = \{c_{ij}\} \in \mathbb{R}^{n \times n}$$

    **Vysvětlení:** Toto je **tabulka** (matice) velikosti $n \times n$ (n řádků a n sloupců). Každé políčko $c_{ij}$ říká, jak moc často (nebo jak důležité je), aby robot přecházel mezi zařízením $i$ a zařízením $j$.

    - $c_{12}$ = jak často robot jezdí mezi zařízením 1 a 2
    - $c_{34}$ = jak často robot jezdí mezi zařízením 3 a 4
    - atd.

    **Poznámka:** Symbol $\mathbb{R}^{n \times n}$ znamená, že je to matice reálných čísel (tedy můžeme mít i desetinná čísla).

### Hledané řešení - Permutace

**Cíl:** Najít nejlepší **pořadí** (uspořádání) zařízení.

$$\pi = (\pi_1, \pi_2, \pi_3, ..., \pi_n)$$

**Vysvětlení symbolu $\pi$ (pí):** Toto je **permutace** - což je matematický název pro "přeházené pořadí".

**Praktický příklad:**

-   Máme 4 zařízení: $\{1, 2, 3, 4\}$
-   Jedno možné uspořádání: $\pi = (2, 4, 1, 3)$
    -   Na první pozici dáme zařízení číslo 2
    -   Na druhou pozici dáme zařízení číslo 4
    -   Na třetí pozici dáme zařízení číslo 1
    -   Na čtvrtou pozici dáme zařízení číslo 3

**Symbol $S_n$:** Označuje množinu všech možných permutací $n$ prvků.

-   Pro $n = 3$ existuje $3! = 6$ permutací: (1,2,3), (1,3,2), (2,1,3), (2,3,1), (3,1,2), (3,2,1)
-   Pro $n = 10$ existuje $10! = 3\,628\,800$ permutací!

---

## Cenová funkce - Jak hodnotíme kvalitu uspořádání

Každé uspořádání $\pi$ má svou "cenu" - číslo, které říká, jak moc je toto uspořádání dobré nebo špatné. **Čím nižší cena, tím lepší uspořádání!**

### Hlavní vzorec

$$\min_{\pi \in S_n} f_{\text{SRFLP}}(\pi)$$

**Vysvětlení:**

-   Symbol $\min$ znamená "minimalizuj" - hledáme nejmenší možnou hodnotu
-   $\pi \in S_n$ znamená "pro všechny možné permutace"
-   Celý zápis říká: "Najdi takovou permutaci $\pi$, pro kterou je funkce $f_{\text{SRFLP}}(\pi)$ co nejmenší"

### Jak se počítá cena uspořádání

$$f_{\text{SRFLP}}(\pi) = \sum_{1 \le i < j \le n} \left[ c_{\pi_i \pi_j} \cdot d(\pi_i, \pi_j) \right]$$

**Rozeberme si to po částech:**

#### Symbol $\sum$ (sigma)

Symbol $\sum_{1 \le i < j \le n}$ znamená **"sečti pro všechny páry"**.

**Konkrétně:**

-   Vezmeme všechny možné dvojice zařízení, kde $i < j$
-   Pro každou dvojici vypočítáme hodnotu ve hranaté závorce
-   Všechny tyto hodnoty sečteme dohromady

**Příklad pro $n = 4$:**

Sčítáme přes tyto páry:

-   $(i=1, j=2)$
-   $(i=1, j=3)$
-   $(i=1, j=4)$
-   $(i=2, j=3)$
-   $(i=2, j=4)$
-   $(i=3, j=4)$

Celkem 6 párů (obecně pro $n$ zařízení je to $\frac{n(n-1)}{2}$ párů).

#### Co se děje uvnitř sumy

Pro každý pár zařízení počítáme:

$$c_{\pi_i \pi_j} \cdot d(\pi_i, \pi_j)$$

**Význam:**

-   $c_{\pi_i \pi_j}$ = **váha přechodu** mezi zařízeními na pozici $i$ a pozici $j$ (z matice $C$)
-   $d(\pi_i, \pi_j)$ = **vzdálenost** mezi zařízeními na pozici $i$ a pozici $j$
-   Násobíme je = **čím častější přechod** (vyšší váha) a **čím delší vzdálenost**, tím horší (vyšší cena)

---

## Výpočet vzdálenosti mezi zařízeními

$$d(\pi_i, \pi_j) = \frac{l_{\pi_i} + l_{\pi_j}}{2} + \sum_{i < k < j} l_{\pi_k}$$

**Toto je klíčový vzorec! Vysvětlíme si ho podrobně.**

### Co znamená tento vzorec?

Vzdálenost mezi dvěma zařízeními se skládá ze dvou částí:

1. **První část:** $\frac{l_{\pi_i} + l_{\pi_j}}{2}$

    **Vysvétlení:** Měříme vzdálenost od **středu** prvního zařízení do **středu** druhého zařízení. Proto bereme polovinu šířky každého zařízení.

    **Proč?** Robot obvykle pracuje uprostřed každého pásu, ne na jeho kraji.

2. **Druhá část:** $\sum_{i < k < j} l_{\pi_k}$

    **Vysvětlení:** Symbol $\sum_{i < k < j}$ znamená "sečti šířky všech zařízení, která jsou MEZI pozicí $i$ a pozicí $j$".

    Robot musí "projet" přes všechna zařízení, která jsou mezi výchozím a cílovým zařízením.

### Praktický příklad výpočtu vzdálenosti

Máme 5 zařízení v tomto pořadí: **Z₁, Z₂, Z₃, Z₄, Z₅**

Jejich šířky: $l_1 = 2$, $l_2 = 1$, $l_3 = 3$, $l_4 = 1.5$, $l_5 = 2.5$

**Otázka:** Jaká je vzdálenost mezi Z₂ a Z₅?

```
Pozice:    1      2      3      4      5
Zařízení: |--Z₁--|--Z₂--|--Z₃--|--Z₄--|--Z₅--|
Šířka:     2m     1m     3m    1.5m   2.5m
```

**Výpočet:**

$$d(Z_2, Z_5) = \frac{l_2 + l_5}{2} + l_3 + l_4$$

$$d(Z_2, Z_5) = \frac{1 + 2.5}{2} + 3 + 1.5$$

$$d(Z_2, Z_5) = 1.75 + 3 + 1.5 = 6.25 \text{ metrů}$$

**Interpretace:**

-   $\frac{1 + 2.5}{2} = 1.75$ m = polovina Z₂ + polovina Z₅
-   $3$ m = celá šířka Z₃ (robot projede přes něj)
-   $1.5$ m = celá šířka Z₄ (robot projede přes něj)

---

## Formát vstupních dat

Problém (instance SRFLP) je zadán v textovém souboru. Ukážeme si konkrétní příklad.

### Příklad: Instance s 10 zařízeními

```
10
1 1 1 1 1 1 1 1 1 1
0 30 17 11 24 25 24 17 16 22
0 0 21 23 26 24 27 19 11 32
0 0 0 24 18 23 31 36 28 19
0 0 0 0 19 18 33 25 20 28
0 0 0 0 0 15 37 27 17 16
0 0 0 0 0 0 27 23 29 24
0 0 0 0 0 0 0 27 31 24
0 0 0 0 0 0 0 0 14 18
0 0 0 0 0 0 0 0 0 24
0 0 0 0 0 0 0 0 0 0
```

### Struktura souboru

**Řádek 1:** Počet zařízení

```
10
```

Máme $n = 10$ zařízení.

**Řádek 2:** Šířky zařízení

```
1 1 1 1 1 1 1 1 1 1
```

Všechna zařízení mají šířku 1 (jednotku). Tedy $l_1 = l_2 = ... = l_{10} = 1$.

**Řádky 3-12:** Matice vah $C$

Toto je **horní trojúhelníková matice**. Proč jen polovina? Protože váha přechodu z $i$ do $j$ je stejná jako z $j$ do $i$: $c_{ij} = c_{ji}$.

**Jak číst matici:**

| Od\Do  | 1   | 2   | 3   | 4   | 5   | 6   | 7   | 8   | 9   | 10  |
| ------ | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| **1**  | 0   | 30  | 17  | 11  | 24  | 25  | 24  | 17  | 16  | 22  |
| **2**  | 30  | 0   | 21  | 23  | 26  | 24  | 27  | 19  | 11  | 32  |
| **3**  | 17  | 21  | 0   | 24  | 18  | 23  | 31  | 36  | 28  | 19  |
| **4**  | 11  | 23  | 24  | 0   | 19  | 18  | 33  | 25  | 20  | 28  |
| **5**  | 24  | 26  | 18  | 19  | 0   | 15  | 37  | 27  | 17  | 16  |
| **6**  | 25  | 24  | 23  | 18  | 15  | 0   | 27  | 23  | 29  | 24  |
| **7**  | 24  | 27  | 31  | 33  | 37  | 27  | 0   | 27  | 31  | 24  |
| **8**  | 17  | 19  | 36  | 25  | 27  | 23  | 27  | 0   | 14  | 18  |
| **9**  | 16  | 11  | 28  | 20  | 17  | 29  | 31  | 14  | 0   | 24  |
| **10** | 22  | 32  | 19  | 28  | 16  | 24  | 24  | 18  | 24  | 0   |

**Příklady:**

-   $c_{1,2} = 30$ znamená, že přechod mezi zařízením 1 a 2 má váhu 30
-   $c_{5,6} = 15$ znamená, že přechod mezi zařízením 5 a 6 má váhu 15
-   $c_{8,9} = 14$ znamená, že přechod mezi zařízením 8 a 9 má váhu 14 (nízká váha = méně častý přechod)

---

## Kompletní příklad výpočtu ceny pro malou instanci

Pro lepší pochopení si ukážeme **kompletní výpočet** na velmi malém příkladu.

### Miniaturní instance: 3 zařízení

**Data:**

-   $n = 3$ zařízení
-   Šířky: $l_1 = 1$, $l_2 = 2$, $l_3 = 1$
-   Matice vah:

$$
C = \begin{pmatrix}
0 & 5 & 3 \\
5 & 0 & 8 \\
3 & 8 & 0
\end{pmatrix}
$$

**Uvažované uspořádání:** $\pi = (1, 2, 3)$ (zařízení v původním pořadí)

```
Pozice:    1      2      3
Zařízení: |--1--|----2----|--3--|
Šířka:     1m     2m       1m
```

### Výpočet vzdáleností

**1. Vzdálenost mezi pozicí 1 a 2** (zařízení 1 a 2):

$$d(\pi_1, \pi_2) = \frac{l_1 + l_2}{2} + \sum_{1 < k < 2} l_{\pi_k}$$

Mezi pozicí 1 a 2 není žádné zařízení, takže suma je 0:

$$d(\pi_1, \pi_2) = \frac{1 + 2}{2} + 0 = 1.5 \text{ m}$$

**2. Vzdálenost mezi pozicí 1 a 3** (zařízení 1 a 3):

$$d(\pi_1, \pi_3) = \frac{l_1 + l_3}{2} + l_2$$

$$d(\pi_1, \pi_3) = \frac{1 + 1}{2} + 2 = 1 + 2 = 3 \text{ m}$$

**3. Vzdálenost mezi pozicí 2 a 3** (zařízení 2 a 3):

$$d(\pi_2, \pi_3) = \frac{l_2 + l_3}{2} + 0 = \frac{2 + 1}{2} = 1.5 \text{ m}$$

### Výpočet celkové ceny

$$f_{\text{SRFLP}}(\pi) = c_{12} \cdot d(\pi_1, \pi_2) + c_{13} \cdot d(\pi_1, \pi_3) + c_{23} \cdot d(\pi_2, \pi_3)$$

$$f_{\text{SRFLP}}(\pi) = 5 \cdot 1.5 + 3 \cdot 3 + 8 \cdot 1.5$$

$$f_{\text{SRFLP}}(\pi) = 7.5 + 9 + 12 = 28.5$$

**Cena tohoto uspořádání je 28.5.**

### Porovnání s jiným uspořádáním

**Uspořádání:** $\pi = (2, 1, 3)$ (prohodíme první dvě zařízení)

```
Pozice:    1        2      3
Zařízení: |----2----|--1--|--3--|
Šířka:      2m       1m    1m
```

**Vzdálenosti:**

-   $d(\pi_1, \pi_2) = \frac{2+1}{2} + 0 = 1.5$ m
-   $d(\pi_1, \pi_3) = \frac{2+1}{2} + 1 = 2.5$ m
-   $d(\pi_2, \pi_3) = \frac{1+1}{2} + 0 = 1$ m

**Cena:**

$$f_{\text{SRFLP}}(\pi) = c_{21} \cdot 1.5 + c_{23} \cdot 2.5 + c_{13} \cdot 1$$

$$f_{\text{SRFLP}}(\pi) = 5 \cdot 1.5 + 8 \cdot 2.5 + 3 \cdot 1$$

$$f_{\text{SRFLP}}(\pi) = 7.5 + 20 + 3 = 30.5$$

**Cena tohoto uspořádání je 30.5** - horší než předchozí!

---

## Metoda řešení: Branch and Bound (Backtracking)

### Co je Branch and Bound?

**Branch and Bound** je algoritmus pro hledání optimálního řešení v problémech, kde existuje obrovské množství možností. Místo toho, abychom zkoušeli úplně všechny možnosti, algoritmus:

1. **Branch (Větvení):** Postupně vytváří možná řešení
2. **Bound (Ohraničení):** Odhaduje, zda má smysl danou větev dále zkoumat
3. **Prune (Ořezání):** Zahodí větve, které nemohou vést k lepšímu řešení

### Analogie: Hledání nejlevnější cesty

Představte si, že plánujete cestu do 10 měst a chcete najít **nejlevnější** cestu, která všechna města navštíví. Máte před sebou strom možností:

```
Start
├─ Do města A (cena 100)
│  ├─ Pak do B (celkem 250)
│  │  └─ Pak do C (celkem 500)
│  └─ Pak do C (celkem 280)
│     └─ Pak do B (celkem 450)
└─ Do města B (cena 150)
   ├─ Pak do A (celkem 280)
   └─ Pak do C (celkem 300)
```

**Bez Branch and Bound:** Prošli bychom úplně všechny možnosti (10! = 3,628,800 cest).

**S Branch and Bound:** Když už máme nějaké řešení za 400, a vidíme větev, která už teď má cenu 500, nemusíme ji dále zkoumat - nemůže být lepší!

### Backtracking jako speciální případ

**Backtracking** je jednodušší verze Branch and Bound:

1. Postupně stavíme řešení (např. přidáváme zařízení do uspořádání)
2. Když zjistíme, že částečné řešení už je horší než nejlepší známé, **vrátíme se zpět** (backtrack) a zkusíme jinou variantu
3. Pokračujeme, dokud neprozkoumáme všechny smysluplné možnosti

### Sdílení nejlepšího řešení v paralelním výpočtu

**Proč paralelizace?**

-   Pro $n = 10$ zařízení je 10! = 3,628,800 možností
-   Jeden procesor by to počítal dlouho
-   **Řešení:** Rozdělíme práci mezi více procesorů/vláken

**Sdílení informací:**

-   Každé vlákno hledá nejlepší řešení ve své části prostoru
-   Když jedno vlákno najde lepší řešení, **sdílí ho s ostatními**
-   Ostatní vlákna pak mohou ořezat více větví (vědí, co je potřeba překonat)

**Analogie:** Skupina lidí hledá poklad na ostrově:

-   Rozdělí se do různých směrů
-   Když někdo najde poklad za 100 zlatých, řekne to ostatním rádiem
-   Ostatní už nemusí zkoumat cesty, které určitě budou horší než 100 zlatých

---

## Vaše zadání - Konkrétní úkol

### Co máte implementovat

**Vytvořte paralelní program**, který:

1. **Načte instanci problému** ze souboru `Y-t_10.txt`

    - Přečte počet zařízení
    - Přečte šířky zařízení
    - Přečte matici vah

2. **Implementuje algoritmus Branch and Bound**

    - Postupně generuje permutace (uspořádání)
    - Počítá cenu každého uspořádání podle vzorců výše
    - Ořezává nepersp ektivní větve

3. **Využije paralelismus**

    - Rozdělí výpočet mezi více vláken/procesorů
    - Vlákna si **sdílejí** informaci o nejlepším dosud nalezeném řešení
    - Tím se zvýší efektivita ořezávání

4. **Najde optimální řešení**
    - Vrátí nejlepší permutaci $\pi^*$
    - Vrátí její cenu $f_{\text{SRFLP}}(\pi^*)$

### Vstupní soubor

Váš program bude pracovat se souborem **Y-t_10.txt**, který obsahuje:

-   10 zařízení
-   Jejich šířky
-   Matici vah přechodů mezi zařízeními

### Očekávaný výstup

Program by měl vypsat:

-   **Nejlepší nalezené uspořádání:** například `(3, 7, 1, 9, 5, 2, 8, 4, 10, 6)`
-   **Cenu tohoto uspořádání:** například `2847.5`
-   **Čas výpočtu:** kolik sekund/minut to trvalo
-   **Informace o paralelizaci:** kolik vláken bylo použito, kolik větví bylo ořezáno, atd.

---

## Doporučení a nápovědy

### 1. Vycházejte z problému TSP

**Traveling Salesman Problem (TSP)** je velmi podobný problém:

-   Také hledá permutaci (pořadí měst)
-   Také minimalizuje nějakou funkci (celkovou vzdálenost cesty)
-   Také se řeší pomocí Branch and Bound

**Rozdíly:**

-   V TSP měříme vzdálenost mezi městy přímo z matice
-   V SRFLP musíme vzdálenost vypočítat podle vzorce (závisí na šířkách zařízení a pozicích)

### 2. Struktura programu

**Doporučená struktura:**

```
1. Načtení dat ze souboru
   - Funkce pro parsing vstupního souboru

2. Výpočet vzdálenosti
   - Funkce d(pi_i, pi_j) podle vzorce výše

3. Výpočet ceny uspořádání
   - Funkce f_SRFLP(pi)

4. Branch and Bound algoritmus
   - Rekurzivní funkce pro generování permutací
   - Ořezávání pomocí aktuálně nejlepší známé ceny

5. Paralelizace
   - Rozdělení prostoru permutací mezi vlákna
   - Sdílená proměnná pro nejlepší řešení (thread-safe!)

6. Výstup
   - Vypsání výsledku
```

### 3. Testování

Než začnete řešit plnou instanci s 10 zařízeními:

1. **Otestujte na malé instanci** (3-4 zařízení) - můžete ručně ověřit správnost
2. **Zkontrolujte výpočet vzdáleností** - vypište si mezivýsledky
3. **Ověřte paralelizaci** - spusťte s 1 vláknem vs. více vlákny - výsledek musí být stejný!

### 4. Optimalizace

**Důležité triky:**

-   **Symetrie:** $c_{ij} = c_{ji}$ - počítejte jen jednou
-   **Ořezávání:** Čím dříve zaktualizujete nejlepší řešení, tím více větví můžete ořezat
-   **Dolní odhad (bound):** Zkuste odhadnout minimální možnou cenu pro částečné řešení

---

## Příklad pseudokódu

```pseudocode
// Globální proměnné (sdílené mezi vlákny)
nejlepší_cena = nekonečno
nejlepší_permutace = prázdná
zámek = mutex pro synchronizaci

funkce branch_and_bound(částečná_permutace, zbývající_zařízení, aktuální_cena):
    // Pokud jsme vytvořili kompletní permutaci
    pokud zbývající_zařízení je prázdná:
        celková_cena = aktuální_cena + dopočítej_zbytek(částečná_permutace)

        pokud celková_cena < nejlepší_cena:
            zamkni(zámek):
                pokud celková_cena < nejlepší_cena:  // double-check
                    nejlepší_cena = celková_cena
                    nejlepší_permutace = částečná_permutace
        návrat

    // Bound - ořezání
    dolní_odhad = aktuální_cena + optimistický_odhad(částečná_permutace, zbývající)
    pokud dolní_odhad >= nejlepší_cena:
        návrat  // Tato větev nemůže vést k lepšímu řešení

    // Branch - větvení
    pro každé zařízení z in zbývající_zařízení:
        nová_permutace = částečná_permutace + [z]
        nová_cena = aktuální_cena + vypočti_přírůstek(z, částečná_permutace)
        nové_zbývající = zbývající_zařízení - {z}

        branch_and_bound(nová_permutace, nové_zbývající, nová_cena)

// Paralelní spuštění
funkce main():
    načti_data("Y-t_10.txt")

    vytvoř_vlákna(počet_vláken):
        každé_vlákno:
            // Rozdělíme první úroveň stromu mezi vlákna
            pro moje_počáteční_zařízení:
                branch_and_bound([zařízení], zbytek, 0)

    počkej_na_dokončení_všech_vláken()

    vypiš(nejlepší_permutace, nejlepší_cena)
```

---

## Reference a další studium

Pro hlubší pochopení problému doporučuji přečíst:

**Kothari, R., Ghosh, D.** (2012). _The single row facility layout problem: state of the art._ OPSEARCH 49, 442–462.  
[https://doi.org/10.1007/s12597-012-0091-4](https://doi.org/10.1007/s12597-012-0091-4)

Tento článek obsahuje:

-   Historii problému SRFLP
-   Různé varianty problému
-   Přehled algoritmů pro řešení
-   Porovnání efektivity různých přístupů

---

## Shrnutí klíčových bodů

### Co je SRFLP?

Hledání nejlepšího pořadí zařízení v řadě, aby se minimalizovala celková "cena" přechodů mezi nimi.

### Jak se počítá cena?

Pro každý pár zařízení vynásobíme **váhu přechodu** × **vzdálenost** a vše sečteme.

### Jak se počítá vzdálenost?

Polovina šířky obou zařízení + šířky všech zařízení mezi nimi.

### Jak to vyřešit?

Algoritmem Branch and Bound - systematicky zkoušíme možnosti a ořezáváme ty, které nemohou být lepší.

### Proč paralelně?

Pro 10 zařízení je přes 3 miliony možností - více vláken = rychlejší výpočet.

### Co sdílet mezi vlákny?

Nejlepší dosud nalezené řešení - všichni z něj můžou těžit při ořezávání.

---

---

## Praktické implementační tipy

### Očekávaný výsledek

Pro instanci `Y-t_10.txt` je **optimální cena: 5596**

Tento výsledek použijte pro ověření správnosti vaší implementace!

### Implementace vzorců v kódu

#### Implementace první sumy (výpočet celkové ceny)

Vzorec: $f_{\text{SRFLP}}(\pi) = \sum_{1 \le i < j \le n} \left[ c_{\pi_i \pi_j} \cdot d(\pi_i, \pi_j) \right]$

**Implementace pomocí dvou vnořených cyklů:**

```python
# Python příklad
def calculate_cost(permutation, widths, cost_matrix):
    n = len(permutation)
    total_cost = 0.0

    # První suma - dva vnořené cykly
    for i in range(n):
        for j in range(i + 1, n):
            # Vypočítáme vzdálenost mezi pozicí i a j
            d = calculate_distance(permutation, widths, i, j)

            # Přičteme k celkové ceně: váha × vzdálenost
            # permutation[i] a permutation[j] jsou indexy zařízení
            facility_i = permutation[i]
            facility_j = permutation[j]
            total_cost += cost_matrix[facility_i][facility_j] * d

    return total_cost
```

**Vysvětlení:**

-   `i` iteruje od 0 do n-1
-   `j` iteruje od i+1 do n-1 (zajišťuje, že `i < j`)
-   Tím procházíme všechny unikátní páry pozic

#### Implementace druhé sumy (výpočet vzdálenosti)

Vzorec: $d(\pi_i, \pi_j) = \frac{l_{\pi_i} + l_{\pi_j}}{2} + \sum_{i < k < j} l_{\pi_k}$

**Implementace pomocí jednoho cyklu:**

```python
# Python příklad
def calculate_distance(permutation, widths, i, j):
    facility_i = permutation[i]
    facility_j = permutation[j]

    # První část: polovina šířky obou krajních zařízení
    distance = (widths[facility_i] + widths[facility_j]) / 2.0

    # Druhá část: suma šířek všech zařízení mezi i a j
    for k in range(i + 1, j):
        facility_k = permutation[k]
        distance += widths[facility_k]

    return distance
```

**Vysvětlení:**

-   `k` iteruje od `i+1` do `j-1` (zajišťuje `i < k < j`)
-   Sčítáme šířky všech zařízení, která leží mezi pozicemi `i` a `j`

### Načtení dat ze souboru

**Kompletní funkce pro načtení instance SRFLP:**

```python
def load_data():
    path = "Y-t_10.txt"

    with open(path) as f:
        lines = f.readlines()

        # Řádek 1: počet zařízení
        number_of_items = int(lines[0])

        # Řádek 2: šířky zařízení
        widths = lines[1]
        widths = [int(x) for x in widths.split()]

        # Řádky 3+: horní trojúhelníková matice vah
        tmp = lines[2:(2 + number_of_items)]
        data = []

        for line in tmp:
            row = []
            parts = line.split()
            for part in parts:
                row.append(float(part))
            data.append(row)

    # DŮLEŽITÉ: Doplnění dolní části matice
    # Matice je symetrická: c_ij = c_ji
    n = len(data)
    for i in range(n):
        for j in range(i + 1, n):
            data[j][i] = data[i][j]

    return [widths, data]
```

**Vysvětlení důležitých částí:**

1. **Parsing šířek:**

    ```python
    widths = [int(x) for x in widths.split()]
    ```

    Rozdělí řádek podle mezer a převede každou hodnotu na integer.

2. **Načtení horní trojúhelníkové matice:**

    ```python
    tmp = lines[2:(2 + number_of_items)]
    ```

    Vezme řádky od indexu 2 do (2 + n), což je n řádků matice.

3. **⚠️ KRITICKÉ: Doplnění dolní části matice**

    ```python
    for i in range(n):
        for j in range(i + 1, n):
            data[j][i] = data[i][j]
    ```

    **Proč je to potřeba?**

    - Vstupní soubor obsahuje jen horní trojúhelník matice (nad diagonálou)
    - Musíme doplnit dolní část, aby byla matice úplná a symetrická
    - Bez toho by při přístupu k `data[j][i]` pro `j > i` byl buď index mimo rozsah nebo nulová hodnota

    **Příklad:**

    Vstup (horní trojúhelník):

    ```
    0  30  17
    0   0  21
    0   0   0
    ```

    Po doplnění (celá symetrická matice):

    ```
    0  30  17
    30  0  21
    17 21   0
    ```

### Poznámky pro implementaci v Go

Pokud budete implementovat v **Go** místo Pythonu, myslete na tyto rozdíly:

1. **Indexování:** Go používá stejné indexování od 0 jako Python ✅

2. **Typy:** Go je staticky typovaný

    ```go
    var widths []int
    var costMatrix [][]float64
    var permutation []int
    ```

3. **Rozsahy cyklů:**

    ```go
    // Python: for i in range(n)
    for i := 0; i < n; i++ { ... }

    // Python: for j in range(i+1, n)
    for j := i + 1; j < n; j++ { ... }
    ```

4. **Synchronizace vláken:**

    ```go
    // Použijte sync.Mutex pro sdílenou proměnnou
    var mutex sync.Mutex
    var bestCost float64 = math.Inf(1)

    // Při aktualizaci:
    mutex.Lock()
    if newCost < bestCost {
        bestCost = newCost
        bestPermutation = newPerm
    }
    mutex.Unlock()
    ```

5. **Paralelizace:**

    ```go
    // Goroutines pro paralelní výpočet
    var wg sync.WaitGroup

    for startFacility := 0; startFacility < n; startFacility++ {
        wg.Add(1)
        go func(start int) {
            defer wg.Done()
            branchAndBound([]int{start}, remaining, 0.0)
        }(startFacility)
    }

    wg.Wait() // Počkáme na dokončení všech goroutines
    ```

### Kontrolní seznam implementace

Před finálním spuštěním zkontrolujte:

-   ✅ **Matice je správně načtena** - dolní část je doplněna ze symetrických hodnot
-   ✅ **Vzdálenost se počítá správně** - polovina krajních + suma prostředních šířek
-   ✅ **Indexování permutace** - rozlišujete mezi indexem pozice a indexem zařízení
-   ✅ **Paralelizace je thread-safe** - sdílené proměnné jsou chráněny mutexem
-   ✅ **Ořezávání funguje** - větve s cenou ≥ nejlepší_cena se přeskakují
-   ✅ **Výsledek je 5596** - pokud ne, máte chybu v implementaci!

### Testovací mini instance

Pro rychlé ověření správnosti implementace vytvořte malou testovací instanci:

**test_3.txt:**

```
3
1 2 1
0 5 3
0 0 8
0 0 0
```

**Očekávané nejlepší uspořádání:** `(1, 2, 3)` s cenou **28.5** (viz kompletní příklad výše)

Pokud váš program vrátí tento výsledek, máte správně implementované základní funkce!

---

## Závěr

Teď byste měli mít kompletní pochopení problému SRFLP i praktické tipy pro implementaci! Klíčem k úspěchu je:

1. ✅ Správně implementovat výpočet vzdáleností a ceny
2. ✅ **Nezapomenout doplnit dolní část matice!**
3. ✅ Efektivně prořezávat prostor řešení
4. ✅ Správně synchronizovat vlákna při sdílení nejlepšího řešení
5. ✅ Ověřit výsledek - pro `Y-t_10.txt` musí být **5596**

Hodně štěstí při implementaci! 🚀
