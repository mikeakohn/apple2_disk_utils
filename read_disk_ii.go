package main

import "fmt"
import "os"
import "apple2_disk"

func main() {
  fmt.Println("read_disk_ii: Apple II disk reader - Copyright 2015 Michael Kohn")

  if len(os.Args) < 2 {
    fmt.Println("Usage: " + os.Args[0] + " <binfile> <optional:file to dump>")
    os.Exit(0)
  }

  apple2_disk := new(apple2_disk.Apple2Disk)

  if apple2_disk.Load(os.Args[1]) == false {
    os.Exit(1)
  }

  apple2_disk.PrintDiskInfo()
  apple2_disk.PrintCatalog()

  if len(os.Args) == 3 {
    track, sector, is_binary := apple2_disk.FindFile(os.Args[2])

    if track == 0 && sector == 0 {
      fmt.Println("Error: File not found '" + os.Args[2] + "'")
    } else {
      //apple2_disk.DumpSector(track, sector)
      apple2_disk.PrintFileSectorList(track, sector)

      if is_binary {
        apple2_disk.DumpBinaryFile(track, sector)
      } else {
        apple2_disk.PrintTextFile(track, sector)
      }
    }
  }
}

