package main

import "fmt"
import "os"
import "apple2_disk"

func main() {
  fmt.Println("read_disk_ii: Apple II disk reader - Copyright 2015 Michael Kohn")

  if len(os.Args) < 2 {
    fmt.Println("Usage: " + os.Args[0] + " <binfile> <optional:file to dump> <optional:outputfile>")
    os.Exit(0)
  }

  apple2_disk := new(apple2_disk.Apple2Disk)

  if apple2_disk.Load(os.Args[1]) == false {
    os.Exit(1)
  }

  apple2_disk.PrintDiskInfo()
  apple2_disk.PrintCatalog()

  if len(os.Args) == 4 {
    apple2_disk.DumpFile(os.Args[2], os.Args[3])
  }
}

