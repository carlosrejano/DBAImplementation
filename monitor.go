package DBASim
import (
	"sync"
)
//Struct that synchronise go routines lets go!
type Monitor struct {

	numberOfOnt     	int
	ronda 				int
	contractsDone 		*int
	buffersDone 		*int
	portCounter			*int
	mapDone				map[int]bool
	dataSent 			*int
	contractCh         	chan [2]int
	bufferCh           	chan [2]int
	mapCh              	map[int] chan[2]int
	start				map[int]bool
	m                  	sync.Mutex
	contractsCondition 	*sync.Cond
	buffersCondition   	*sync.Cond
	mapsCondition      	*sync.Cond
	startCondition 		*sync.Cond
	restartCondition 		*sync.Cond
}

func (mon *Monitor) New(n int) {
	mon.numberOfOnt = n
	mon.ronda = 0
	mon.contractCh = make(chan [2]int, n)
	mon.bufferCh = make(chan [2]int, n)
	mon.mapCh = make(map[int](chan [2]int))
	mon.start = make(map[int]bool)
	mon.m = sync.Mutex{}
	mon.contractsCondition = sync.NewCond(&mon.m)
	mon.buffersCondition = sync.NewCond(&mon.m)
	mon.mapsCondition = sync.NewCond(&mon.m)
	mon.startCondition = sync.NewCond(&mon.m)
	mon.restartCondition = sync.NewCond(&mon.m)
	mon.buffersDone = new(int)
	mon.contractsDone = new(int)
	mon.portCounter = new(int)
	mon.mapDone = make(map[int]bool)
	mon.dataSent= new(int)
}
