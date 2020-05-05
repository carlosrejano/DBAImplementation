package DBASim

import (
	"math/rand"
	"strconv"
	"sync"
	"time"

	//"time"
)

type Ont struct {
	bufferMax 		int
	ronda			int
	portID      	int
	bufferLength	int
	contract     	int
	interval     	[2]int
	Results 		map[int][2]int
	bufferR 		map[int]int
	mon          	*Monitor
}
//Creation ONT
func (o *Ont) NewOnt(contract int, bufferMax int,  mon *Monitor) {
	o.mon = mon
	o.ronda = 0
	o.contract = contract
	o.portID = *mon.portCounter
	*mon.portCounter++
	o.interval[0] = -2
	o.interval[1] = -2
	o.bufferLength = 0
	o.bufferMax = bufferMax
	o.Results = make(map[int][2]int)
	o.bufferR = make(map[int]int)
}

//Report to OLT the contract
func (o *Ont) ReportContract() {
	aux := [2]int{o.portID, o.contract}
	o.mon.contractCh <- aux
	o.mon.contractsCondition.L.Lock()
	*o.mon.contractsDone++
	if *o.mon.contractsDone == o.mon.numberOfOnt {
		close(o.mon.contractCh)
	}
	o.mon.contractsCondition.Signal()
	o.mon.contractsCondition.L.Unlock()

}
//Generate a random buffer
func (o *Ont) GenerateBuffer() {
	rand.Seed(time.Now().UnixNano()*int64(o.portID))
	o.bufferLength = rand.Intn(o.bufferMax) + 150
	o.bufferR[o.ronda] = o.bufferLength
}
//Report Buffer to OLT
func (o *Ont) ReportBuffer() {
	aux := [2]int{o.portID, o.bufferLength}
	o.mon.bufferCh <- aux
	o.mon.buffersCondition.L.Lock()
	*o.mon.buffersDone++
	if *o.mon.buffersDone == o.mon.numberOfOnt {
		close(o.mon.bufferCh)
	}
	o.mon.buffersCondition.Signal()
	o.mon.buffersCondition.L.Unlock()
}
//Simulates the data sent
func (o *Ont) SendData() {
	o.Results[o.ronda] = o.interval
	o.mon.restartCondition.L.Lock()
	*o.mon.dataSent++
	o.mon.restartCondition.Signal()
	o.mon.restartCondition.L.Unlock()
}
//Gets de interval from the Olt
func (o *Ont) GetInterval() {
	for k, v := range o.mon.mapCh {
		if k == o.portID {
			s := <-v
			o.interval = s
		}
	}
}
//Method that waits for the map to get interval
func (o *Ont) WaitingMap() {
	o.mon.mapsCondition.L.Lock()
	for o.mon.mapDone[o.ronda] == false {

		o.mon.mapsCondition.Wait()
	}
	o.mon.mapsCondition.L.Unlock()

}
//Sync go routines at the start
func (o *Ont) StartOk() {
	o.mon.startCondition.L.Lock()
	for o.mon.start[o.ronda] == false{
		o.mon.startCondition.Wait()
	}
	o.mon.startCondition.L.Unlock()
}

//All the above are methods to display the results correctly
func (o *Ont) GetPortID() string {
	return strconv.Itoa(o.portID)
}
func (o *Ont) GetBuffer(ronda int) string {
	if o.bufferR[ronda] <= 999 {
		return strconv.Itoa(o.bufferR[ronda])
	} else {
		return strconv.Itoa(o.bufferR[ronda])
	}
}
func (o *Ont) GetOffset(ronda int) string {
	return strconv.Itoa(o.Results[ronda][0])
}
func (o *Ont) GetIntervalR(ronda int) string {
	return strconv.Itoa(o.Results[ronda][1])
}
func (o *Ont) GetAbleToSend(ronda int) string {
	if o.Results[ronda][1]==-1 {
		return "NO"
	} else {
		return "YES"
	}
}

//Main routine of the ONT
func (o *Ont) Routine(wg *sync.WaitGroup, rondas int) {

	o.ReportContract()

	for o.ronda < rondas{
		o.StartOk()
		o.GenerateBuffer()
		o.ReportBuffer()
		o.WaitingMap()
		o.GetInterval()
		o.SendData()
		o.ronda++
	}
	wg.Done()
}


