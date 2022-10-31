package main

import (
	"fmt"
	aud "github.com/go-audio/wav"
	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"log"
	"math"
	"math/cmplx"
	"os"
	"strconv"
	"time"
)

func track(msg string) (string, time.Time) {
	return msg, time.Now()
}

func duration(msg string, start time.Time) {
	log.Printf("%v: %v\n", msg, time.Since(start))
}

func C(xi, xi_1, yj float64) float64 {
	var c float64 = 1 //стоимость
	if (xi_1 <= xi && xi <= yj) || (xi_1 >= xi && xi >= yj) {
		return c
	} else {
		return c + math.Min(math.Abs(xi-xi_1), math.Abs(xi-yj))
	}
}

func Msm(x, y []float64) float64 {
	//defer duration(track("MSM"))
	cost := make([][]float64, len(x))
	for i := 0; i < len(x); i++ {
		cost[i] = make([]float64, len(y))
	}
	//initialization
	cost[0][0] = math.Abs(x[0] - y[0])
	for i := 1; i < len(x); i++ {
		cost[i][0] = cost[i-1][0] + C(x[i], x[i-1], y[0])
	}
	for j := 1; j < len(y); j++ {
		cost[0][j] = cost[0][j-1] + C(y[j], x[0], y[j-1])
	}
	//main loop
	for i := 1; i < len(x); i++ {
		for j := 1; j < len(y); j++ {
			cost[i][j] = math.Min(cost[i-1][j-1]+math.Abs(x[i]-y[j]), math.Min(cost[i-1][j]+C(x[i], x[i-1], y[j]), cost[i][j-1])+C(y[j], x[i], y[j-1]))
		}
	}
	return cost[len(x)-1][len(y)-1]
}

func ZeroPaddingPeriodogramm(x, y []float64) float64 {
	//len(x) должно быть больше = len(y)
	//longer periodogramm
	nx := len(x)
	ny := len(y)
	var mx int = nx / 2
	var my int = ny / 2
	var sum1 float64
	var w1 float64
	for i := 0; i < len(x); i++ {
		if i < mx {
			w1 = 2 * math.Pi * float64(i) / float64(nx)
		}

		sum1 += math.Abs(x[i] * math.Pow(math.E, imag(cmplx.Sqrt(-1))*w1))
	}
	//zero padding
	for i := 0; i < nx-ny; i++ {
		y = append(y, 0)
	}
	//shorter periodogramm
	var sum2 float64
	var w2 float64
	for i := 0; i < len(y); i++ {
		if i < my {
			w2 = 2 * math.Pi * float64(i) / float64(len(y))
		}
		sum2 += math.Abs(y[i] * math.Pow(math.E, imag(cmplx.Sqrt(-1))*w2))
	}

	p1 := sum1 * sum1 / float64(nx)
	p2 := sum2 * sum2 / float64(len(y))
	zeroPadding := (1 / float64(mx)) * math.Pow(p1-p2, 2)
	return math.Sqrt(zeroPadding)
}

func ReducedPeriodogram(x, y []float64) float64 {
	//len(x) должно быть больше = len(y)
	nx := len(x)
	ny := len(y)
	//var mx float64 = float64(nx / 2)
	var my float64 = float64(ny / 2)
	var sum1 float64
	var w1 float64
	//longer periodogramm
	for i := 0; i < nx; i++ {
		//frequency on len(y) for longer time series
		if i < int(my) {
			w1 = 2 * math.Pi * float64(i) / float64(ny)
		}
		sum1 += math.Abs(x[i] * math.Pow(math.E, imag(cmplx.Sqrt(-1))*w1))
	}

	//shorter periodogramm
	var sum2 float64
	var w2 float64
	for i := 0; i < ny; i++ {
		if i < int(my) {
			w2 = 2 * math.Pi * float64(i) / float64(len(y))
		}
		sum2 += math.Abs(y[i] * math.Pow(math.E, imag(cmplx.Sqrt(-1))*w2))
	}

	prp1 := sum1 * sum1 / float64(nx)
	p2 := sum2 * sum2 / float64(ny)
	reduced := (1 / my) * math.Pow(prp1-p2, 2)
	return math.Sqrt(reduced)
}

func InterpolatedPeriodogram(x, y []float64) float64 {
	//len(x) должно быть больше len(y)
	var sum1 float64
	var sum11 float64
	var sum111 float64
	var wr float64 = 1
	var wr1 float64 = 1
	//var wx float64
	var wy float64 = 1
	var r int
	//longer periodogramm
	for i := 0; i < len(x); i++ {
		//frequency on len(y) for longer time series
		if i <= len(y)/2 {
			r = i * ((len(x) / 2) / (len(y) / 2))
		}
		//fmt.Println(r)
		if i <= len(y)/2 && i != 0 {
			wy = 2 * math.Pi * float64(i) / float64(len(y))
		}
		if i <= r && i != 0 {
			wr = 2 * math.Pi * float64(i) / float64(len(x))
		}
		if i <= r+1 && i != 0 {
			wr1 = 2 * math.Pi * float64(i) / float64(len(x))
		}
		//fmt.Println(wy, wr, wr1)
		sum1 = math.Abs(x[i] * math.Pow(math.E, imag(cmplx.Sqrt(-1))*wr))
		sum11 = math.Abs(x[i] * math.Pow(math.E, imag(cmplx.Sqrt(-1))*wr1))
		sum111 += sum1*(1-(wy-wr)/(wr1-wr)) + sum11*((wy-wr)/(wr1-wr))
		//sum111 += sum1 + sum11
	}
	//shorter periodogramm
	var sum2 float64
	var w2 float64

	for i := 0; i < len(y); i++ {
		if i > len(y)/2 {
			w2 = 2 * math.Pi * float64(i) / float64(len(y))
		}

		sum2 += math.Abs(y[i] * math.Pow(math.E, imag(cmplx.Sqrt(-1))*w2))
	}

	prp1 := sum111
	p2 := sum2 * sum2 / float64(len(y))
	reduced := (1 / float64(len(y)) / 2) * math.Pow(prp1-p2, 2)
	return reduced
}

func Lcss(x, y []float64) float64 {
	//defer duration(track("LCSS"))
	lenX := len(x)
	lenY := len(y)
	lenLcss := make([][]float64, lenX+1)
	for i := 0; i < lenX+1; i++ {
		lenLcss[i] = make([]float64, lenY+1)
	}

	for i := 0; i < lenX+1; i++ {
		for j := 0; j < lenY+1; j++ {
			if i == 0 || j == 0 {
				lenLcss[i][j] = 0
			} else if x[i-1] == y[j-1] {
				lenLcss[i][j] = lenLcss[i-1][j-1] + 1
			} else {
				lenLcss[i][j] = math.Max(lenLcss[i-1][j], lenLcss[i][j-1])
			}
		}
	}
	return lenLcss[lenX][lenY]
}

func MinMax(array []float64) (float64, float64) {
	var max float64 = array[0]
	var min float64 = array[0]
	for _, value := range array {
		if max < value {
			max = value
		}
		if min > value {
			min = value
		}
	}
	return min, max
}

func Edr(x, y []float64, t float64) float64 {
	//defer duration(track("EDR"))
	d := func(x, y float64) float64 { return math.Abs(x - y) }
	lenX := len(x)
	lenY := len(y)
	edr := make([][]float64, lenX+1)
	for i := 0; i < lenX+1; i++ {
		edr[i] = make([]float64, lenY+1)
	}

	edr[0][0] = math.Abs(x[0] - y[0])

	for i := 1; i < lenX+1; i++ {
		edr[i][0] = float64(i - 1)
	}
	for j := 1; j < lenY+1; j++ {
		edr[0][j] = float64(j - 1)
	}

	for i := 1; i < lenX+1; i++ {
		for j := 1; j < lenY+1; j++ {
			if d(x[i-1], y[j-1]) < t {
				edr[i][j] = edr[i-1][j-1]
			} else {
				edr[i][j] = math.Min(edr[i-1][j-1]+0.1, math.Min(edr[i-1][j]+0.1, edr[i][j-1]+0.1))
			}
		}
	}
	return edr[lenX][lenY]
}

func Dtw(x, y []float64) float64 {
	lenX := len(x)
	lenY := len(y)
	dtw := make([][]float64, lenX+1)
	for i := 0; i < lenX+1; i++ {
		dtw[i] = make([]float64, lenY+1)
	}

	for i := 0; i < lenX+1; i++ {
		for j := 0; j < lenY+1; j++ {
			dtw[i][j] = math.Inf(1)
		}
	}
	dtw[0][0] = 0

	for i := 1; i < lenX+1; i++ {
		for j := 1; j < lenY+1; j++ {
			cost := math.Abs(x[i-1] - y[j-1])
			lastMin := math.Min(dtw[i-1][j], math.Min(dtw[i][j-1], dtw[i-1][j-1]))
			dtw[i][j] = cost + lastMin
		}
	}

	//fmt.Println(dtw)
	return dtw[len(x)][len(y)]
}

func Erp(x, y []float64, g float64) float64 {
	d := func(x, y float64) float64 { return math.Abs(x - y) }
	lenX := len(x)
	lenY := len(y)
	//expSol := 0.0
	erp := make([][]float64, lenX+1)
	for i := 0; i < lenX+1; i++ {
		erp[i] = make([]float64, lenY+1)
	}

	for i := 1; i < lenX+1; i++ {
		for j := 1; j < lenY+1; j++ {
			erp[i][j] = math.Min(erp[i-1][j-1]+d(x[i-1], y[j-1]), math.Min(erp[i-1][j]+d(x[i-1], g), erp[i][j-1]+d(y[j-1], g)))
		}
	}

	return erp[lenX][lenY]
}

func ParseAudio(name string) []float64 {
	//defer duration(track("AudioReader"))
	f, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	d := aud.NewDecoder(f)
	if d.Err() != nil {
		log.Fatal(err)
	}

	buf, _ := d.FullPCMBuffer()
	_, err = d.PCMBuffer(buf)
	if err != nil {
		panic(err)
	}
	res := buf.AsFloatBuffer().Data
	return res
}

func FullNormalized(series []float64) []float64 {
	lenS := len(series)
	normSeries := make([]float64, lenS)

	for i := 0; i < lenS; i++ {
		min, max := MinMax(series)
		normSeries[i] = (series[i] - min) / (max - min)
	}
	return normSeries
}

func PlotPcm(data []float64, title opts.Title) *charts.Line {
	x := make([]string, 0)
	y := make([]opts.LineData, 0)
	for i := 0; i < len(data); i++ {
		//x = append(x, fmt.Sprintf("%.1f", float64(i)/signal.SampleRate))
		y = append(y, opts.LineData{Value: data[i], Symbol: "none"})
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		//charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWonderland}),
		//charts.WithYAxisOpts(opts.YAxis{Max: signal.Max(), Min: signal.Min(), SplitNumber: 10}),
		charts.WithXAxisOpts(opts.XAxis{SplitNumber: 100}),
		charts.WithTitleOpts(title))

	line.SetXAxis(x).AddSeries("data", y).SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: false,
		}),
	)

	return line
}

func PlotPcm1(data []float64, data1 []float64, title opts.Title) *charts.Line {
	x := make([]string, 0)
	y := make([]opts.LineData, 0)
	for i := 0; i < len(data); i++ {
		//x = append(x, fmt.Sprintf("%.1f", float64(i)/signal.SampleRate))
		y = append(y, opts.LineData{Value: data[i], Symbol: "none"})
	}

	z := make([]float64, 50)
	data1 = append(z, data1...)

	y1 := make([]opts.LineData, 0)
	for i := 0; i < len(data1); i++ {
		//x = append(x, fmt.Sprintf("%.1f", float64(i)/signal.SampleRate))
		y1 = append(y1, opts.LineData{Value: data1[i], Symbol: "none"})
	}

	line := charts.NewLine()
	line.SetGlobalOptions(
		//charts.WithInitializationOpts(opts.Initialization{Theme: types.ThemeWonderland}),
		//charts.WithYAxisOpts(opts.YAxis{Max: signal.Max(), Min: signal.Min(), SplitNumber: 10}),
		charts.WithXAxisOpts(opts.XAxis{SplitNumber: 100}),
		charts.WithTitleOpts(title))

	line.SetXAxis(x).AddSeries("data", y).SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: false,
		}),
	)

	line.SetXAxis(x).AddSeries("data1", y1).SetSeriesOptions(
		charts.WithLineChartOpts(opts.LineChart{
			Smooth: false,
		}),
	)

	return line
}

func Draw2(x, y []float64, name string) {
	plot1 := PlotPcm1(x, y, opts.Title{
		Title: "Sample1",
		//Subtitle: signal.String(),
	})
	/*
		plot2 := PlotPcm(y, opts.Title{
			Title: "Sample2",
			//Subtitle: signal.String(),
		})

	*/
	file4, _ := os.Create(name)
	plot1.Render(file4)
	//plot2.Render(file4)
}

func Draw(x, y []float64, name string) {
	plot1 := PlotPcm(x, opts.Title{
		Title: "Sample1",
		//Subtitle: signal.String(),
	})

	plot2 := PlotPcm(y, opts.Title{
		Title: "Sample2",
		//Subtitle: signal.String(),
	})

	file4, _ := os.Create(name)
	plot1.Render(file4)
	plot2.Render(file4)
}

//signal envelope function
func E(x []float64, w int) []float64 {
	e := make([]float64, len(x)-w)
	for i := 0; i < len(x)-w; i++ {
		for j := i; j < i+w; j++ {
			e[i] += x[j] * x[j]
		}
		e[i] = e[i] / float64(w)
	}
	return e
}

func Normalization(x, y []float64) ([]float64, []float64) {
	minX, maxX := MinMax(x)
	minY, maxY := MinMax(y)
	min, max := 0.0, 0.0

	if minY <= minX {
		min = minX
	} else {
		min = minY
	}
	if maxY >= maxX {
		max = maxY
	} else {
		max = maxX
	}

	lenX := len(x)
	normSeriesX := make([]float64, lenX)
	lenY := len(y)
	normSeriesY := make([]float64, lenY)

	for i := 0; i < lenX; i++ {
		normSeriesX[i] = (x[i] - min) / (max - min)
	}
	for i := 0; i < lenY; i++ {
		normSeriesY[i] = (y[i] - min) / (max - min)
	}
	return normSeriesX, normSeriesY

}

func main() {
	file, err := os.Create("data.txt")
	if err != nil {
		fmt.Println("Unable to create file:", err)
		os.Exit(1)
	}
	defer file.Close()
	methods := []string{"edr", "erp", "dtw", "lcss", "msm"}

	//example for 5 methods (m), two ideal audio (i) and six distorted (p)
	for m := 0; m < len(methods); m++ {
		file.WriteString("-------" + methods[m] + "-------" + "\n")
		for p := 1; p <= 6; p++ {
			pStr := strconv.Itoa(p)
			file.WriteString("------Папка" + pStr + "-------" + "\n")
			for i := 1; i <= 2; i++ {
				iStr := strconv.Itoa(i)
				for j := 1; j <= 90; j++ {
					jStr := strconv.Itoa(j)
					x := ParseAudio("C:/Users/borov/Desktop/metrics/0" + iStr + "/" + jStr + ".wav")
					y := ParseAudio("C:/Users/borov/Desktop/metrics/" + pStr + "/" + jStr + ".wav")
					switch methods[m] {
					case "edr":
						xNorm, yNorm := Normalization(x, y)
						xNormSmoth := E(xNorm, 100)
						yNormSmoth := E(yNorm, 100)
						edr := Edr(xNormSmoth, yNormSmoth, 0.1)
						strEdr := strconv.FormatFloat(edr, 'f', 6, 64)
						file.WriteString(strEdr + "\n")
					case "erp":
						xNorm, yNorm := Normalization(x, y)
						xNormSmoth := E(xNorm, 100)
						yNormSmoth := E(yNorm, 100)
						erp := Erp(xNormSmoth, yNormSmoth, 0)
						strEdr := strconv.FormatFloat(erp, 'f', 6, 64)
						file.WriteString(strEdr + "\n")
					case "dtw":
						xNorm, yNorm := Normalization(x, y)
						xNormSmoth := E(xNorm, 100)
						yNormSmoth := E(yNorm, 100)
						dtw := Dtw(xNormSmoth, yNormSmoth)
						strEdr := strconv.FormatFloat(dtw, 'f', 6, 64)
						file.WriteString(strEdr + "\n")
					case "lcss":
						xNorm, yNorm := Normalization(x, y)
						xNormSmoth := E(xNorm, 100)
						yNormSmoth := E(yNorm, 100)
						lcss := Lcss(xNormSmoth, yNormSmoth)
						strEdr := strconv.FormatFloat(lcss, 'f', 6, 64)
						file.WriteString(strEdr + "\n")
					case "msm":
						xNorm, yNorm := Normalization(x, y)
						xNormSmoth := E(xNorm, 100)
						yNormSmoth := E(yNorm, 100)
						msm := Msm(xNormSmoth, yNormSmoth)
						strEdr := strconv.FormatFloat(msm, 'f', 6, 64)
						file.WriteString(strEdr + "\n")
					}
				}
				file.WriteString("------------------" + "\n")
			}
		}
	}

}
