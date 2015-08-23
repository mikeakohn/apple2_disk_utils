package apple2_disk

import "fmt"
import "os"

// http://fileformats.archiveteam.org/wiki/Apple_DOS_file_system

type Apple2Disk struct {
  data []byte
}

func GetOffset(track int, sector int) int {
  return (sector * 256) + (track * (256 * 16))
}

func GetInt16(data []byte, offset int) int {
  return int(data[offset]) | (int(data[offset + 1]) << 8)
}

func (apple2_disk *Apple2Disk) Load(filename string) bool {
  file, err := os.Open(filename)

  if err != nil {
    panic(err)
  }

  defer file.Close()

  stat, err := file.Stat()
  if err != nil {
    panic(err)
  }

  apple2_disk.data = make([]byte, int(stat.Size()))

  file.Read(apple2_disk.data)

  return true
}

func (apple2_disk *Apple2Disk)PrintDiskInfo() {
  // Track 17, Sector 0 should be the disk info area.  At least on 5 1/4"
  // disks.
  offset := GetOffset(17, 0)

  fmt.Println("================ Disk Info ===================")
  fmt.Printf("       Catalog Track: %d\n", apple2_disk.data[offset + 0x01])
  fmt.Printf("      Catalog Sector: %d\n", apple2_disk.data[offset + 0x02])
  fmt.Printf("      Release Number: %d\n", apple2_disk.data[offset + 0x03])
  fmt.Printf("       Volume Number: %d\n", apple2_disk.data[offset + 0x06])
  fmt.Printf("   Max Track/Sectors: %d\n", apple2_disk.data[offset + 0x27])
  fmt.Printf("     Last Used Track: %d\n", apple2_disk.data[offset + 0x30])
  fmt.Printf("Track Alloc Dirction: %d\n", apple2_disk.data[offset + 0x31])
  fmt.Printf("     Tracks Per Disk: %d\n", apple2_disk.data[offset + 0x34])
  fmt.Printf("    Sectors Per Disk: %d\n", apple2_disk.data[offset + 0x35])
  fmt.Printf("    Bytes Per Sector: %d\n", GetInt16(apple2_disk.data, offset + 0x36))
}


