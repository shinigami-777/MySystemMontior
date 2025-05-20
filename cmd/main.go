package main

import (
	"MySystemMontior/cmd/hardware"
	"fmt"
	"time"
)

func main() {
	fmt.Println("Starting the Golang System Monitor....")

	go func() {
		for {
			systemSection, err := hardware.GetSystemSection()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(systemSection)
			}

			diskSection, err := hardware.GetDiskSection()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(diskSection)
			}

			cpuSection, err := hardware.GetCPUSection()
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(cpuSection)
			}

			fmt.Println("---------------------------------------------------")

			//keep updating every 2 seconds
			time.Sleep(3 * time.Second)

		}
	}()

	// sleep the main thread for 5 Minutes
	time.Sleep(5 * time.Minute)
}
