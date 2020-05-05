package DBASim

import (
	"strconv"
	"sync"
)

type Olt struct {
	maxBandwidth    int
	totalBandwidth  int
	numberOfOnt     int
	mapBuffers      map[int]int
	MapMultiplexing map[int][2]int
	mapContracts    map[int]int
	mapBytesSentR   map[int]int
	offset          int
	bytesSent 		int
	reserve			int
	ronda 			int
	mon             *Monitor
}

func (olt *Olt) NewOlt(mon *Monitor, n int) {
	olt.mon = mon
	olt.offset = 0
	olt.reserve = 0
	olt.numberOfOnt = n
	olt.mapBuffers = make(map[int]int)
	olt.mapContracts = make(map[int]int)
	olt.mapBytesSentR = make(map[int]int)
	olt.MapMultiplexing = make(map[int][2]int)
	olt.bytesSent = 0
}

//DBA implementation
func (olt *Olt) CalculateMap() {
	for k, v := range olt.mapBuffers {
		maxBytes := (olt.mapContracts[k]*125)/8
		if olt.bytesSent + v > 19440 {
			unlucky := [2]int{-1, -1}
			olt.MapMultiplexing[k] = unlucky
			continue
		}
		if v <= maxBytes {
			timeNeeded := float32(v* 8) / 1244.16
			olt.reserve += maxBytes - v
			aux := [2]int{olt.offset, int(timeNeeded)}
			olt.MapMultiplexing[k] = aux
			olt.offset += int(timeNeeded)
			olt.bytesSent += v
		}else {
			if v - maxBytes <= olt.reserve {
				timeNeeded := float32(v* 8) / 1244.66
				olt.reserve -= (v - maxBytes)
				aux := [2]int{olt.offset, int(timeNeeded)}
				olt.MapMultiplexing[k] = aux
				olt.offset += int(timeNeeded) + 5
				olt.bytesSent += v
			}else {
				unlucky := [2]int{-1, -1}
				olt.MapMultiplexing[k] = unlucky
			}
		}
	}
	olt.mapBytesSentR[olt.mon.ronda] = olt.bytesSent
	olt.reserve = 0
	olt.offset = 0
	olt.bytesSent = 0
}
//Get buffers from ONTs
func (olt *Olt) GetMap() {
	for v := range olt.mon.bufferCh {

		olt.mapBuffers[v[0]] = v[1]
	}
}
//Get contracts from ONT
func (olt *Olt) GetContracts() {
	for v := range olt.mon.contractCh {
		olt.mapContracts[v[0]] = v[1]
	}
}

//Sends map to ONTs
func (olt *Olt) SendMap() {
	i := 0

	for k, v := range olt.MapMultiplexing {
		i++
		olt.mon.mapCh[k] = make(chan [2]int, 1)
		olt.mon.mapCh[k] <- v
		close(olt.mon.mapCh[k])

	}

	olt.mon.mapsCondition.L.Lock()
	olt.mon.mapDone[olt.mon.ronda] = true
	olt.mon.mapsCondition.Broadcast()
	olt.mon.mapsCondition.L.Unlock()
}

//Sync methods
func (olt *Olt) WaitingContracts() {
	olt.mon.contractsCondition.L.Lock()
	for *olt.mon.contractsDone < olt.numberOfOnt {
		olt.mon.contractsCondition.Wait()
	}
	olt.mon.contractsDone =new(int)
	olt.mon.contractsCondition.L.Unlock()
}

func (olt *Olt) WaitingBuffer() {
	olt.mon.buffersCondition.L.Lock()
	for *olt.mon.buffersDone < olt.numberOfOnt {
		olt.mon.buffersCondition.Wait()

	}
	*olt.mon.buffersDone = 0
	olt.mon.buffersCondition.L.Unlock()
}
func (olt *Olt) Start() {
	olt.mon.startCondition.L.Lock()
	olt.mon.start[olt.mon.ronda] = true
	olt.mon.startCondition.Broadcast()
	olt.mon.startCondition.L.Unlock()
}

func (olt *Olt) Restart() {
	olt.mon.restartCondition.L.Lock()
	for *olt.mon.dataSent < olt.mon.numberOfOnt {
		olt.mon.restartCondition.Wait()
	}
	*olt.mon.dataSent = 0
	olt.mon.bufferCh = make(chan [2]int, 100)
	olt.mon.contractCh = make(chan [2]int, 100)
	olt.mon.restartCondition.L.Unlock()
}
func (olt *Olt) GetBytesSent(ronda int) string {
	a := float32(olt.mapBytesSentR[ronda])/float32(19440)
	return strconv.FormatFloat(float64(a), 'f', 4, 64)
}

//Main go routine for the OLT
func (olt *Olt) Routine(wg *sync.WaitGroup, rondas int) {
	olt.WaitingContracts()
	olt.GetContracts()


	for olt.mon.ronda < rondas {
		olt.Start()
		olt.WaitingBuffer()
		olt.GetMap()
		olt.CalculateMap()
		olt.SendMap()
		olt.Restart()
		olt.mon.ronda++
	}
	wg.Done()
}



