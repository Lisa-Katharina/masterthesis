/*
 * Source code for generating a plot for satellite orbit usage over time based on the UCS Satellite Database,
 * as supplementary material for the master thesis:
 *
 * "Hacker-Attacks Against Satellites. An Evaluation of Space Law in Regard to the Nature of Hacker-Attacks"
 *
 * written by Lisa-Katharina Hlavica at the Vrije Universiteit Amsterdam, for obtaining the masters degree in International Technology Law.
 *
 * @Copyright 2021 Lisa-Katharina Hlavica
 */

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	// URL to UCS satellite database
	databaseURL = "https://www.ucsusa.org/sites/default/files/2021-02/UCS-Satellite-Database-1-1-2021.txt"
	
	// name for output chart file
	fileName = "orbit-chart.html"

	printHeader = false
)

func main() {

	// download the database into memory
	resp, err := http.Get(databaseURL)
	if err != nil {
		log.Fatal(err)
	}

	// check the response status code
	if resp.StatusCode != http.StatusOK {
		log.Fatal("http request failed: ", resp.Status)
	}

	// read data from response
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// close response body on exit
	defer resp.Body.Close()

	// prepare mapping from years to map from orbit type to number of satellites
	var yearsMap = map[string]map[string]int{}

	// process each line in database
	for index, line := range strings.Split(string(data), "\n") {
		
		// separate fields for each column
		fields := strings.Split(line, "\t")

		// handle first line: the header
		if index == 0 {

			if printHeader {
				
				// print fields and their index
				for i, r := range fields {
					
					// only if not empty
					if r != "" {
						fmt.Println(r, i)	
					}
				}
			}
			
			continue
		}

		if len(fields) > 19 {
			
			orbitType := fields[8]
			dateOfLaunch := fields[19]

			// if date of launch not empty
			if dateOfLaunch != "" {

				// process date format, eg: 12/11/2001
				year := strings.Split(dateOfLaunch, "/")[2]
				
				// initialize map if not present yet
				if _, ok := yearsMap[year]; !ok {
					yearsMap[year] = map[string]int{}
				}
				
				// collect orbit type count for year
				yearsMap[year][orbitType]++
			}
		}
	}

	// declare array for year values
	var yearValues []string
	
	// collect year values from map
	for year := range yearsMap {
		yearValues = append(yearValues, year)
	}

	// sort year values
	sort.Strings(yearValues)
	
	// create output file
	f, err := os.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	// close on exit
	defer f.Close()

	// create bar chart
	bar := createBarChart(yearValues, yearsMap)

	// render chart into file
	bar.Render(f)

	fmt.Println("done! wrote chart into file", fileName)
}

// createBarChart creates a new bar for the given years, and fills it with series for the values of interest.
func createBarChart(years []string, yearMap map[string]map[string]int) *charts.Bar {
	
	// create new bar
	bar := charts.NewBar()
	
	// set options
	bar.SetGlobalOptions(
		
		// set title
		charts.WithTitleOpts(
			opts.Title{
				Title: "Satellite orbit usage per year",
			},
		),

		// show legend
		charts.WithLegendOpts(
			opts.Legend{
				Show: true,
			},
		),
	)

	const (
		leo = "LEO"
		meo = "MEO"
		geo = "GEO"
		elliptical = "Elliptical"
	)

	// set data for X axis and add values for orbit series
	bar.SetXAxis(years).
		AddSeries(leo, generateBarItems(years, yearMap, leo)).
  		AddSeries(meo, generateBarItems(years, yearMap, meo)).
		AddSeries(geo, generateBarItems(years, yearMap, geo)).
		AddSeries(elliptical, generateBarItems(years, yearMap, elliptical))

	return bar
}

// generateBarItems creates an array of bar chart data points for the given seriesName,
// based on the (sorted) array of years and the orbit-usage-per-year map.
func generateBarItems(years []string, yearMap map[string]map[string]int, seriesName string) []opts.BarData {
	
	// prepare result array
	dataPoints := make([]opts.BarData, 0)

	// for every year
	for _, y := range years {

		// collect bar data point for the given seriesName
		dataPoints = append(dataPoints, opts.BarData{
			Value: yearMap[y][seriesName],
		})
	}
	
	return dataPoints
}