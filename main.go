package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/jung-kurt/gofpdf"
	"github.com/shopspring/decimal"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"image"
	"image/color"
	"image/png"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/tools/bezier"
)

var database = "movies_metadata.csv"
var columns = map[int]interface{}{
	2:  "budget",
	3:  "genres",
	14: "release_date",
	15: "revenue",
	20: "title",
}

type Genre struct {
	Name  string
	Count int
}

type FilterAndFunc struct {
	Filter string
	Func   func(film map[string]interface{}) bool
}
type Point struct {
	x, y float64
}

func main() {

	start := time.Now()
	// a - mean, p - positive profit, n - negative profit
	var output []string
	var pngs []image.Image

	output = append(output, "^Notation \r\n")
	output = append(output, repeatingBox("\r\n", 2))
	output = append(output, repeatingBox("\r\n", 2))

	output = append(output, "A - mean, p. - Positive, n. - Negative, \r")
	output = append(output, repeatingBox("\r", 2))
	output = append(output, repeatingBox("\r", 2))
	output = append(output, "C - Costs (only budget), R - Revenue, Pr - Profit, RtoC - Revenue to Costs \r\n")
	output = append(output, repeatingBox("\r\n", 2))
	output = append(output, repeatingBox("\r\n", 2))

	films := readCSV(database, columns)
	for _, film := range films {
		film["genres"] = parseGenres(film["genres"].(string))
	}

	output = append(output, "^General \r\n")
	output = append(output, repeatingBox("\r\n", 2))
	output = append(output, repeatingBox("\r\n", 2))

	//fmt.Println("General films: ", films)

	// Filter films
	films = filter(films, func(film map[string]interface{}) bool {
		if _, err := strconv.Atoi(film["budget"].(string)); err != nil {
			return false
		}
		return film["budget"] != "0"
	})
	films = filter(films, func(film map[string]interface{}) bool {
		if _, err := strconv.Atoi(film["revenue"].(string)); err != nil {
			return false
		}
		return film["revenue"] != "0"
	})
	films = filter(films, func(film map[string]interface{}) bool {
		for _, column := range columns {
			if film[column.(string)] == "" {
				return false
			}
		}
		return true
	})

	// Mean by all films
	_, _, _, _, _, temp, pngTemp := outputItem(films, func(film map[string]interface{}) bool {
		return true
	}, "All films", "")
	pngs = append(pngs, pngTemp)
	output = append(output, temp)
	_, _, _, _, _, temp, pngTemp = outputItem(films, func(film map[string]interface{}) bool {
		return film["release_date"].(string)[0:4] >= "2000"
	}, "Films 2000+", "*")
	pngs = append(pngs, pngTemp)
	output = append(output, temp)
	output = append(output, repeatingBox("\r\n", 18))

	// Years
	output = append(output, "^Years \r\n")
	output = append(output, repeatingBox("\r\n", 2))
	output = append(output, repeatingBox("\r\n", 2))
	var yFilmMap = []FilterAndFunc{
		{"2015-2017", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "2015" && film["release_date"].(string)[6:10] <= "2017"
		}},
		{"2010-2014", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "2010" && film["release_date"].(string)[6:10] < "2014"
		}},
		{"2005-2009", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "2005" && film["release_date"].(string)[6:10] < "2010"
		}},
		{"2000-2004", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "2000" && film["release_date"].(string)[6:10] < "2005"
		}},
		{"1995-1999", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "1995" && film["release_date"].(string)[6:10] < "2000"
		}},
		{"1990-1994", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "1990" && film["release_date"].(string)[6:10] < "1995"
		}},
		{"1985-1989", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "1985" && film["release_date"].(string)[6:10] < "1990"
		}},
		{"1980-1984", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "1980" && film["release_date"].(string)[6:10] < "1985"
		}},
		{"1975-1979", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "1975" && film["release_date"].(string)[6:10] < "1980"
		}},
		{"1970-1974", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "1970" && film["release_date"].(string)[6:10] < "1975"
		}},
		{"1965-1969", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "1965" && film["release_date"].(string)[6:10] < "1970"
		}},
		{"1964-", func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] < "1965"
		}},
	}
	for _, j := range yFilmMap {
		_, _, _, _, _, temp, pngTemp = outputItem(films, j.Func, j.Filter, "*")
		pngs = append(pngs, pngTemp)
		output = append(output, temp)
	}

	// Budget 2000+
	y2000PlusFilms := filter(films, func(film map[string]interface{}) bool {
		return film["release_date"].(string)[6:10] >= "2000"
	})

	output = append(output, "^Budget 2000+ \r\n")
	output = append(output, repeatingBox("\r\n", 2))
	output = append(output, repeatingBox("\r\n", 2))
	var bFilmMap = []FilterAndFunc{
		{"100k-", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) < 100000
		}},
		{"100k-250k", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 100000 && toFloat(film["budget"].(string)) < 250000
		}},
		{"250k-500k", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 250000 && toFloat(film["budget"].(string)) < 500000
		}},
		{"500k-750k", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 500000 && toFloat(film["budget"].(string)) < 750000
		}},
		{"750k-1m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 750000 && toFloat(film["budget"].(string)) < 1000000
		}},
		{"1m-2m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 1000000 && toFloat(film["budget"].(string)) < 2000000
		}},
		{"2m-5m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 2000000 && toFloat(film["budget"].(string)) < 5000000
		}},
		{"2m-3m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 2000000 && toFloat(film["budget"].(string)) < 3000000
		}},
		{"3m-4m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 3000000 && toFloat(film["budget"].(string)) < 4000000
		}},
		{"4m-5m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 4000000 && toFloat(film["budget"].(string)) < 5000000
		}},
		{"5m-10m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 5000000 && toFloat(film["budget"].(string)) < 10000000
		}},
		{"10m-20m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 10000000 && toFloat(film["budget"].(string)) < 20000000
		}},
		{"20m-50m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 20000000 && toFloat(film["budget"].(string)) < 50000000
		}},
		{"50m-100m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 50000000 && toFloat(film["budget"].(string)) < 100000000
		}},
		{"100m+", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 100000000
		}},
	}
	for _, j := range bFilmMap {
		_, _, _, _, _, temp, pngTemp = outputItem(y2000PlusFilms, j.Func, j.Filter, "*")
		pngs = append(pngs, pngTemp)
		output = append(output, temp)
	}

	// Genres 2000+
	output = append(output, "^Genres 2000+ \r\n")
	output = append(output, repeatingBox("\r\n", 2))
	output = append(output, repeatingBox("\r\n", 2))

	genres := getGenres(y2000PlusFilms)
	for _, genre := range genres {
		gFilms := filter(y2000PlusFilms, func(film map[string]interface{}) bool {
			for _, g := range film["genres"].([]string) {
				if g == genre.Name {
					return true
				}
			}
			return false
		})
		_, _, _, _, _, temp, pngTemp = outputItem(gFilms, func(film map[string]interface{}) bool {
			return film["release_date"].(string)[6:10] >= "2000"
		}, genre.Name, "*")
		pngs = append(pngs, pngTemp)
		output = append(output, temp)
	}
	output = append(output, repeatingBox("\r\n", 18))
	output = append(output, repeatingBox("\r\n", 18))

	// Genres + Budget 2000+
	output = append(output, "^Genres + Budget 2000+ \r\n")
	output = append(output, repeatingBox("\r\n", 2))
	output = append(output, repeatingBox("\r\n", 2))
	var gbFilmMap = []FilterAndFunc{
		{"500k-", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) < 500000
		}},
		{"500k-1m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 500000 && toFloat(film["budget"].(string)) < 1000000
		}},
		{"1m-2m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 1000000 && toFloat(film["budget"].(string)) < 2000000
		}},
		{"2m-3m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 2000000 && toFloat(film["budget"].(string)) < 3000000
		}},
		{"3m-4m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 3000000 && toFloat(film["budget"].(string)) < 4000000
		}},
		{"4m-5m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 4000000 && toFloat(film["budget"].(string)) < 5000000
		}},
		{"5m-7.5m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 5000000 && toFloat(film["budget"].(string)) < 7500000
		}},
		{"7.5m-10m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 7500000 && toFloat(film["budget"].(string)) < 10000000
		}},
		{"10m-20m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 10000000 && toFloat(film["budget"].(string)) < 20000000
		}},
		{"10m-50m", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 10000000 && toFloat(film["budget"].(string)) < 50000000
		}},
		{"50m+", func(film map[string]interface{}) bool {
			return toFloat(film["budget"].(string)) >= 50000000
		}},
	}

	for _, genre := range genres {
		gFilms, _, _, _, _, _, _ := outputItem(y2000PlusFilms, func(film map[string]interface{}) bool {
			for _, g := range film["genres"].([]string) {
				if g == genre.Name {
					return true
				}
			}
			return false
		}, genre.Name, "")
		for _, j := range gbFilmMap {
			_, _, _, _, _, temp, pngTemp = outputItem(gFilms, j.Func, genre.Name+" "+j.Filter, "**")
			pngs = append(pngs, pngTemp)
			output = append(output, temp)
		}
	}

	timeProcess := time.Since(start)
	fmt.Printf("Data processing time : %v \n", timeProcess)

	// Save all the pngs
	for i, _png := range pngs {
		if _png == nil {
			_png = image.NewRGBA(image.Rect(0, 0, 1, 1))
		}

		if _, err := os.Stat("pngs"); os.IsNotExist(err) {
			os.Mkdir("pngs", 0755)
		} else {
			os.RemoveAll("pngs")
			os.Mkdir("pngs", 0755)
		}
		f, err := os.Create("pngs/" + strconv.Itoa(i) + ".png")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		err = png.Encode(f, _png)
		if err != nil {
			panic(err)
		}
	}

	// Save the pdf
	makePdf3(output, pngs, "output.pdf")

	fmt.Printf("File writing time : %v \n", time.Since(start)-timeProcess)
	fmt.Printf("Total time : %v \n", time.Since(start))

}

// Data Processing Logic ↓ ↓ ↓
func readCSV(database string, columns map[int]interface{}) (films []map[string]interface{}) {
	// Open the file
	csvfile, err := os.Open(database)
	if err != nil {
		panic(err)
	}

	// Parse the file
	r := csv.NewReader(csvfile)

	// Read the first line
	_, err = r.Read()
	if err != nil {
		panic(err)
	}

	// Read the rest of the file
	for {
		record, err := r.Read()
		if err != nil {
			break
		}

		film := make(map[string]interface{})
		for key, value := range columns {
			film[value.(string)] = record[key]
		}

		films = append(films, film)
	}

	return films
}
func parseGenres(genre string) []string {
	var genreNames []string
	for _, genre := range strings.Split(genre, "},") {
		if strings.Contains(genre, "name") {
			genreName := strings.Split(strings.Split(genre, "name': '")[1], "'")[0]
			genreNames = append(genreNames, genreName)
		}
	}
	return genreNames
}
func filter(films []map[string]interface{}, test func(map[string]interface{}) bool) []map[string]interface{} {
	result := make([]map[string]interface{}, 0)
	for _, film := range films {
		if test(film) {
			result = append(result, film)
		}
	}
	return result
}
func getGenres(films []map[string]interface{}) (genres []Genre) {
	for _, film := range films {
		for _, genre := range film["genres"].([]string) {
			var found bool
			for i, g := range genres {
				if g.Name == genre {
					found = true
					genres[i].Count++
				}
			}
			if !found {
				genres = append(genres, Genre{Name: genre, Count: 0})
			}
		}
	}

	sort.Slice(genres, func(i, j int) bool {
		return genres[i].Count > genres[j].Count
	})

	return genres
}

// Format ↓ ↓ ↓
func easyNumber(value float64) string {
	if value < 0 {
		return "-" + easyNumber(-value)
	} else if value >= 1000000000 {
		return fmt.Sprintf("%.2f billion", value/1000000000)
	} else if value >= 1000000 {
		return fmt.Sprintf("%.2f million", value/1000000)
	} else if value >= 1000 {
		return fmt.Sprintf("%.2f thousand", value/1000)
	}
	return fmt.Sprintf("%.2f", value)
}
func nameShorter(name string) string {
	space := 0
	for i, c := range name {
		if c == ' ' {
			space = i
			break
		}
	}
	switch name[:space] {
	case "Documentary":
		name = "Docum." + name[space:]
		break
	case "Science":
		name = "Sci-Fi" + name[space+len(" Fiction"):]
		break
	case "Animation":
		name = "Anim." + name[space:]
		break
	case "Adventure":
		name = "Advent." + name[space:]
		break
	}
	return name
}
func repeatingBox(text string, height int) string {
	var output string
	for i := 0; i < height-1; i++ {
		output += text
	}
	return output
}

// Math ↓ ↓ ↓
func divide(a, b float64) float64 {
	aDec := decimal.NewFromFloat(a)
	bDec := decimal.NewFromFloat(b)
	if bDec.IsZero() {
		return 0
	}
	out, _ := aDec.Div(bDec).Float64()
	return out
}
func divideFromInt(a, b int) float64 {
	aDec := decimal.NewFromInt(int64(a))
	bDec := decimal.NewFromInt(int64(b))
	if bDec.IsZero() {
		return 0
	}
	out, _ := aDec.Div(bDec).Float64()
	return out
}
func toFloat(value string) float64 {
	dec, _ := decimal.NewFromString(value)
	out, _ := dec.Float64()
	return out
}
func mean(films []map[string]interface{}, test func(map[string]interface{}) float64) float64 {
	var total decimal.Decimal
	for _, film := range films {
		total = total.Add(decimal.NewFromFloat(test(film)))
	}
	filmCount := decimal.NewFromInt(int64(len(films)))
	if filmCount.IsZero() {
		return 0
	}
	out, _ := total.Div(filmCount).Float64()
	return out
}
func median(films []map[string]interface{}, test func(map[string]interface{}) float64) float64 {
	var values []float64
	for _, film := range films {
		values = append(values, test(film))
	}
	sort.Float64s(values)
	if len(values) == 0 {
		return 0
	}
	if len(values)%2 == 0 {
		return divide(values[len(values)/2]+values[len(values)/2-1], 2)
	}
	return values[len(values)/2]
}

// Output ↓ ↓ ↓
func outputItem(films []map[string]interface{}, condition func(film map[string]interface{}) bool, name, clarification string) (filteredFilms []map[string]interface{}, aBudget, aRevenue, aProfit, aRevenueToBudget float64, output string, png image.Image) {
	filteredFilms = filter(films, condition)

	aBudget = mean(filteredFilms, func(film map[string]interface{}) float64 {
		return toFloat(film["budget"].(string))
	})
	aRevenue = mean(filteredFilms, func(film map[string]interface{}) float64 {
		return toFloat(film["revenue"].(string))
	})
	aRevenueToBudget = divide(aRevenue, aBudget)
	aProfit = mean(filteredFilms, func(film map[string]interface{}) float64 {
		return toFloat(film["revenue"].(string)) - toFloat(film["budget"].(string))
	})

	png, _, _ = image.Decode(bytes.NewReader(makeChart(filteredFilms, 25, false, "budget", "revenue")))

	pFilms := filter(filteredFilms, func(film map[string]interface{}) bool {
		return toFloat(film["revenue"].(string)) >= toFloat(film["budget"].(string))
	})
	apBudget := mean(pFilms, func(film map[string]interface{}) float64 {
		return toFloat(film["budget"].(string))
	})
	apRevenue := mean(pFilms, func(film map[string]interface{}) float64 {
		return toFloat(film["revenue"].(string))
	})
	pRevenueToBudget := divide(apRevenue, apBudget)

	nFilms := filter(filteredFilms, func(film map[string]interface{}) bool {
		return toFloat(film["revenue"].(string)) < toFloat(film["budget"].(string))
	})
	anBudget := mean(nFilms, func(film map[string]interface{}) float64 {
		return toFloat(film["budget"].(string))
	})
	anRevenue := mean(nFilms, func(film map[string]interface{}) float64 {
		return toFloat(film["revenue"].(string))
	})
	nRevenueToBudget := divide(anRevenue, anBudget)

	if len(name) > 20 {
		name = nameShorter(name)
	}
	output = fmt.Sprintf("!i"+name+" : %d (%.2f%%"+clarification+") \r\n", len(filteredFilms), float64(len(filteredFilms))/float64(len(films))*100) +
		fmt.Sprintf("AC    : %s USD \r\n", easyNumber(aBudget)) +
		fmt.Sprintf("AR    : %s USD \r\n", easyNumber(aRevenue)) +
		fmt.Sprintf("APr   : %s USD \r\n", easyNumber(aProfit)) +
		fmt.Sprintf("ARtoC : %.2f \r\n", aRevenueToBudget) +
		fmt.Sprintf("P. : %d (%.2f%%) ARtoC : %.3f \r\n", len(pFilms), divideFromInt(len(pFilms), len(filteredFilms))*100, pRevenueToBudget) +
		fmt.Sprintf("N. : %d (%.2f%%) ARtoC : %.3f \r\n", len(nFilms), divideFromInt(len(nFilms), len(filteredFilms))*100, nRevenueToBudget)

	return filteredFilms, aBudget, aRevenue, aProfit, aRevenueToBudget, output, png
}

// Pdf ↓ ↓ ↓
func makePdf3(output []string, pngs []image.Image, filename string) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	// set line spacing
	pdf.SetFont("Courier", "", 8)
	pdf.AddPage()

	// Output into two columns
	var y1, y2, y3 float64 = pdf.GetY(), pdf.GetY(), pdf.GetY()
	var i, j int
	for _, selection := range output {
		if i%3 == 0 {
			if y1 > 215 {
				pdf.AddPage()
				y1, y2, y3 = pdf.GetY(), pdf.GetY(), pdf.GetY()
				i = 0
			}
			for _, str := range strings.Split(selection, "\r\n") {
				pdf.SetXY(10, y1)
				if strings.HasPrefix(str, "^") {
					pdf.SetFontStyle("B")
					pdf.CellFormat(0, 3.5, str[1:], "", 1, "", false, 0, "")
					pdf.SetFontStyle("")
				} else if strings.HasPrefix(str, "!") {
					pdf.CellFormat(0, 3.5, str[2:], "", 1, "", false, 0, "")
				} else {
					pdf.CellFormat(0, 3.5, str, "", 1, "", false, 0, "")
				}
				y1 = pdf.GetY()
			}
			if j < len(pngs) {
				if strings.HasPrefix(selection, "!i") {
					if y1 > 245 {
						pdf.AddPage()
						y1, y2, y3 = pdf.GetY(), pdf.GetY(), pdf.GetY()
						i = 0
					}
					jStr := strconv.Itoa(j)
					pdf.ImageOptions("pngs/"+jStr+".png", 10, y1, 50, 30, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
					y1 += 35
					j++
				}
			}
		} else if i%3 == 1 {
			if y2 > 215 {
				pdf.AddPage()
				y1, y2, y3 = pdf.GetY(), pdf.GetY(), pdf.GetY()
				i = 0
			}
			for _, str := range strings.Split(selection, "\r\n") {
				pdf.SetXY(73, y2)
				if strings.HasPrefix(str, "^") {
					pdf.SetFontStyle("B")
					pdf.CellFormat(0, 3.5, str[1:], "", 1, "", false, 0, "")
					pdf.SetFontStyle("")
				} else if strings.HasPrefix(str, "!") {
					pdf.CellFormat(0, 3.5, str[2:], "", 1, "", false, 0, "")
				} else {
					pdf.CellFormat(0, 3.5, str, "", 1, "", false, 0, "")
				}
				y2 = pdf.GetY()
			}
			if j < len(pngs) {
				if strings.HasPrefix(selection, "!i") {
					if y2 > 245 {
						pdf.AddPage()
						y1, y2, y3 = pdf.GetY(), pdf.GetY(), pdf.GetY()
						i = 0
					}
					jStr := strconv.Itoa(j)
					pdf.ImageOptions("pngs/"+jStr+".png", 73, y2, 50, 30, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
					y2 += 35
					j++
				}
			}
		} else {
			if y3 > 215 {
				pdf.AddPage()
				y1, y2, y3 = pdf.GetY(), pdf.GetY(), pdf.GetY()
				i = 0
			}
			for _, str := range strings.Split(selection, "\r\n") {
				pdf.SetXY(136, y3)
				if strings.HasPrefix(str, "^") {
					pdf.SetFontStyle("B")
					pdf.CellFormat(0, 3.5, str[1:], "", 1, "", false, 0, "")
					pdf.SetFontStyle("")
				} else if strings.HasPrefix(str, "!") {
					pdf.CellFormat(0, 3.5, str[2:], "", 1, "", false, 0, "")
				} else {
					pdf.CellFormat(0, 3.5, str, "", 1, "", false, 0, "")
				}
				y3 = pdf.GetY()
			}
			if j < len(pngs) {
				if strings.HasPrefix(selection, "!i") {
					if y3 > 245 {
						pdf.AddPage()
						y1, y2, y3 = pdf.GetY(), pdf.GetY(), pdf.GetY()
						i = 0
					}
					jStr := strconv.Itoa(j)
					pdf.ImageOptions("pngs/"+jStr+".png", 136, y3, 50, 30, false, gofpdf.ImageOptions{ImageType: "PNG", ReadDpi: true}, 0, "")
					y3 += 35
					j++
				}
			}
		}
		i++
	}

	err := pdf.OutputFileAndClose(filename)
	if err != nil {
		panic(err)
	}
}

// Chart ↓ ↓ ↓
func makeChart(films []map[string]interface{}, n int, logEnabeled bool, paramX, paramY string) []byte {
	const nBez = 1000
	pointsX, pointsY, n := processToChart(films, paramX, paramY, n)
	if pointsX == nil || pointsY == nil {
		return nil
	}

	p := plot.New()
	p.Add(plotter.NewGrid())
	if logEnabeled {
		p.Y.Scale = plot.LogScale{}
	}

	pts := make(plotter.XYs, n)
	for i := range pts {
		pts[i].X = pointsX[i]
		pts[i].Y = pointsY[i]
	}
	s, _ := plotter.NewScatter(pts)
	s.GlyphStyle.Radius = vg.Points(1)
	p.Add(s)

	// Add line
	line, _ := plotter.NewLine(pts)
	line.Color = color.RGBA{R: 255, A: 255}
	p.Add(line)

	// Bezier
	vgpts := make([]vg.Point, n)
	for i := range vgpts {
		vgpts[i].X = vg.Length(pts[i].X)
		vgpts[i].Y = vg.Length(pts[i].Y)
	}
	bez := bezier.New(vgpts...)
	bezpts := make(plotter.XYs, nBez)
	for i := range bezpts {
		tempVgPoint := bez.Point(float64(i) / float64(nBez))
		bezpts[i].X = float64(tempVgPoint.X)
		bezpts[i].Y = float64(tempVgPoint.Y)
	}

	bezline, _ := plotter.NewLine(bezpts)
	bezline.Color = color.RGBA{B: 255, A: 255}
	p.Add(bezline)

	return saveChart(p)
}
func saveChart(p *plot.Plot) []byte {
	wt, err := p.WriterTo(7*vg.Centimeter, 4*vg.Centimeter, "png")
	if err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	wt.WriteTo(buf)
	return buf.Bytes()
}
func processToChart(films []map[string]interface{}, paramX, paramY string, n int) (pointsX, pointsY []float64, newN int) {
	if n > len(films) && len(films) > 1 {
		n = len(films)
	}
	var elementsInSegment = len(films) / n
	if elementsInSegment == 0 {
		return nil, nil, 0
	}
	var sum float64
	var count int

	for i, film := range films {
		if i%elementsInSegment == 0 && i != 0 {
			pointsX = append(pointsX, toFloat(film[paramX].(string)))
			pointsY = append(pointsY, sum/float64(count))
			sum = 0
			count = 0
		}
		sum += toFloat(film[paramY].(string))
		count++
	}
	pointsX = append(pointsX, toFloat(films[len(films)-1][paramX].(string)))
	pointsY = append(pointsY, sum/float64(count))

	for i := 0; i < len(pointsX)/2; i++ {
		pointsX[i], pointsX[len(pointsX)-1-i] = pointsX[len(pointsX)-1-i], pointsX[i]
		pointsY[i], pointsY[len(pointsY)-1-i] = pointsY[len(pointsY)-1-i], pointsY[i]
	}

	return pointsX, pointsY, n
}
