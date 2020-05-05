package main

import (
	"DBASim"
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
)


func main() {
	//Initializes the devices
	var Olt DBASim.Olt
	var mon DBASim.Monitor
	var Ont [1000]DBASim.Ont

	//Get the info from the user
	N, contract, rondes, bufferMax := scanInput()

	//Create devices requested
	mon.New(N)
	Olt.NewOlt(&mon, N)
	wg := sync.WaitGroup{}
	wg.Add(1)

	//Start go routines
	go Olt.Routine(&wg, rondes)
	s := 0
	for s < N {
		wg.Add(1)
		Ont[s].NewOnt(contract, bufferMax, &mon)
		go Ont[s].Routine(&wg, rondes)
		s++
	}
	wg.Wait()

	//Code to output the results
	f, err := os.Create("results.txt")
	if err != nil {
		fmt.Println(err)
		return
	}
	var results string = ""
	var utilitzacioC float64
	var utilitzacioO float64
	i := 0
	j := 0
	for j < rondes {
		NumOntsAble := 0
		results += "Ronda: " + strconv.Itoa(j+1) + "\n"
		results += "PortID\tBytesBuffer\tOffset\tInterval\tAble To Send?\n"
		for i < N {
			if Ont[i].GetAbleToSend(j) == "YES"{
				NumOntsAble++
			}
			results += Ont[i].GetPortID() + "\t" + Ont[i].GetBuffer(j) + "\t\t" + Ont[i].GetOffset(j) + "\t" + Ont[i].GetIntervalR(j) + "\t\t" + Ont[i].GetAbleToSend(j) + "\n"
			i++
		}
		results += "\n"
		aux , _ := strconv.ParseFloat(Olt.GetBytesSent(j), 64)
		utilitzacioC += aux

		results += "Utilització del canal: " + Olt.GetBytesSent(j)
		results += "\n"
		utilitzacioO += float64(NumOntsAble)/float64(N)
		results += "% de Onts que han pogut enviar: " + strconv.FormatFloat(float64(NumOntsAble)/float64(N), 'f', 2, 64)
		results += "\n\n"
		j++
		i = 0
	}
	results += "Utilització canal mitjana: " + strconv.FormatFloat(utilitzacioC/float64(rondes), 'f', 4, 32) + "\n"
	results += "Utilització Onts mitjana: " + strconv.FormatFloat(utilitzacioO/float64(rondes), 'f', 4, 32)

	_, err = f.WriteString(results)
	if err != nil {
		fmt.Println(err)
		f.Close()
		return
	}
}

//function that gets the input from the user
func scanInput() (int, int, int, int) {
	file, err := os.Open("input.txt")

	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	var txtlines []string

	for scanner.Scan() {
		txtlines = append(txtlines, scanner.Text())
	}

	file.Close()
	numOnts, _ := strconv.Atoi(strings.Split(txtlines[1], "Numero de ONTs: ")[1])
	vel, _ := strconv.Atoi(strings.Split(txtlines[2], "Velocitat de les ONTs: ")[1])
	rondes, _ := strconv.Atoi(strings.Split(txtlines[3], "Numero de iteracions: ")[1])
	bufferMax, _ := strconv.Atoi(strings.Split(txtlines[4], "Maxim tamany d'un paquet: ")[1])
	return numOnts, vel, rondes, bufferMax
}
