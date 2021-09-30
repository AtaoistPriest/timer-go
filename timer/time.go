package timer

import (
	"golang.org/x/sys/unix"
	"log"
	"sync"
	"time"
)
const(
	Type1 = 1
	Type2 = 2
	Type3	= 3
)

type Task struct {
	Id int64		    /* id */
	Type uint8		  /* task type */
	SockFd int		  /* socket */
	DstIp [4]byte   /* dst ip and port */
	DstPort int
	Data []byte		  /* data to be sent */
	Count int		    /* retransmit counts */
	Time time.Time  /* the time when the task started executing */
	ReTime int		  /* retransmit time period（unit： s）*/
	ReLimit int 	  /* retransmit count limit */
	IsRepeat bool   /* is it a task that needs to be repeated?*/
}

type Node struct {
	TaskInfo Task
	Back     *Node
}

var Mutex sync.Mutex
// time accuracy threshold（unit：ns）
var threshold = 20

var ch = make(chan int)
//timer
var tickTask = time.NewTicker(1)
var tickBreak = time.NewTicker(time.Hour * 24 * 365)

func Execution(header *Node)  {
	for{
		nodeCur, ok := Next(header)
		if !ok{
			continue
		}
		tNow := time.Now()
		//execution interval of the next task
		duration := time.Duration(nodeCur.TaskInfo.Time.UnixNano() - tNow.UnixNano())
		if duration >= time.Duration(threshold){
			tickTask.Reset(duration)
			select {
				case <- tickTask.C:
					err := taskUnit(header, nodeCur)
					if err != nil {
						log.Print(err)
					}
				//break block state
				case <- tickBreak.C:
					tickBreak.Reset(time.Hour * 24 * 365)
					continue
			}
		}else if -time.Duration(threshold) < duration && duration < time.Duration(threshold){
			// the task is almost timed out, or just time out(within the threshold)
			err := taskUnit(header, nodeCur)
			if err != nil {
				log.Print(err)
			}
		}else{
			// the task has timed out and is ready to execute the next task
			nodeCur.TaskInfo.Time = time.Now().Add(time.Second * time.Duration(nodeCur.TaskInfo.ReTime))
			err := taskUpdate(header, nodeCur)
			if err != nil {
				log.Print( err)
			}
		}
	}
}
/*
	task execute unit
*/
func taskUnit(header *Node, node *Node) error {
	var err error
	if node.TaskInfo.Count > node.TaskInfo.ReLimit && node.TaskInfo.IsRepeat{
		Remove(header, node)
		log.Print("Over")
		return nil
	}
	switch node.TaskInfo.Type {
		case Type1:
			log.Print("[TimerThread][Execution] : execute task1")
		case Type2:
			log.Print("[TimerThread][Execution] : execute task2")
		case Type3:
			log.Print("[TimerThread][Execution] : this is a sending task")
			err = sendData(node)
			if err != nil{
				return err
			}
	}
	if node.TaskInfo.IsRepeat{
		// for tasks that need to be repeated, update time and times after execution
		node.TaskInfo.Count++
		node.TaskInfo.Time = time.Now().Add(time.Second * time.Duration(node.TaskInfo.ReTime))
		err = taskUpdate(header, node)
	}else{
		// one time task, deleted immediately after execution
		Remove(header, node)
	}
	return err
}

func sendData(node *Node) error{
	toAddress := &unix.SockaddrInet4{
		Port: node.TaskInfo.DstPort,
		Addr: node.TaskInfo.DstIp,
	}
	err := unix.Sendto(node.TaskInfo.SockFd, node.TaskInfo.Data, 0, toAddress)
	return err
}

/*
*******************************************************
        operations on timer linked list
*******************************************************
*/

/*
	new linked list
*/
func New() *Node {
	header := Node{
		Back: nil,
	}
	return &header
}

/*
	add new timer task to the timer linked list, and order by time
*/
func Add(head *Node, node *Node) bool{
	timeNow := time.Now()
	if node.TaskInfo.Time.Before(timeNow){
		return false
	}
	tickBreak.Reset(time.Nanosecond * 1)
	Mutex.Lock()
	tmpPre := head
	tmpCur := tmpPre.Back
	for tmpCur != nil{
		if node.TaskInfo.Time.Before(tmpCur.TaskInfo.Time){
			break
		}
		tmpPre = tmpCur
		tmpCur = tmpPre.Back
	}
	tmpPre.Back = node
	node.Back = tmpCur
	Mutex.Unlock()
	return true
}

func Next(header *Node) (*Node, bool) {

	Mutex.Lock()
	res := header.Back
	Mutex.Unlock()
	if res == nil{
		return nil, false
	}
	return res, true
}

func Remove(header *Node,node *Node){

	Mutex.Lock()
	tmpPre := header
	tmpCur := tmpPre.Back
	for tmpCur != nil{
		if node.TaskInfo.Id == tmpCur.TaskInfo.Id{
			break
		}
		tmpPre = tmpCur
		tmpCur = tmpPre.Back
	}
	if tmpCur == nil{
		return
	}
	if header.Back == tmpCur{
		tickBreak.Reset(time.Nanosecond * 1)
	}
	tmpPre.Back = tmpCur.Back
	Mutex.Unlock()
}

func taskUpdate(header *Node, node *Node) error{
	Mutex.Lock()
	header.Back = node.Back
	Mutex.Unlock()
	Add(header, node)
	return nil
}
