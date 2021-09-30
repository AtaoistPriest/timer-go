package main

import (
	"timer"
	"time"
)

func main(){
	header := timer.New()
	go timer.Execution(header)
	node1 := timer.Node{
		TaskInfo: timer.Task{
			Id: time.Now().UnixNano(),
			Type: timer.Type1,
			Count: 0,
			Time: time.Now().Add(time.Duration(time.Second * 2)),
			ReTime: 2,
			IsRepeat: false,
		},
		Back: nil,
	}
	timer.Add(header, &node)

	node2 := timer.Node{
		TaskInfo: timer.Task{
			Id: time.Now().UnixNano(),
			Type: timer.Type1,
			Count: 0,
			Time: time.Now().Add(time.Duration(time.Second * 2)),
			ReTime: 2,
                                                ReLimit: 3,
			IsRepeat: true,
		},
		Back: nil,
	}
	timer.Add(header, &node)
	time.Sleep(time.Second * 1000)
}
