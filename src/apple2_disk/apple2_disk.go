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

func (apple2_disk *Apple2Disk) PrintDiskInfo() {
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

func (apple2_disk *Apple2Disk) PrintCatalog() {
  offset := GetOffset(17, 0)

  track := int(apple2_disk.data[offset + 1])
  sector := int(apple2_disk.data[offset + 2])

  fmt.Println("================ Catalog ===================")

  for true {
    fmt.Printf("Track: %d  Sector: %d\n", track, sector)

    offset := GetOffset(track, sector)

    // For each possible catalog entry.
    for i := 0x0b; i < 256; i += 0x23 {
      fmt.Printf("%02x: ", i)

      // Print file name.
      for n := 0x03; n <= 0x20; n++ {
        ch := int(apple2_disk.data[offset + i + n])
        ch &= 0x7f
        if ch >= 32 && ch < 127 {
          fmt.Printf("%c", ch)
        } else {
          fmt.Print(" ")
        }
      }

      // File type and flags.
      file_type := int(apple2_disk.data[offset + i + 2])
      var type_string string
      locked := false

      type_string = " "

      if (file_type & 0x80) != 0 { file_type &= 0x7f; locked = true }

      if file_type == 0x00 {
        type_string += "TXT "
      } else {
        if (file_type & 0x01) != 0 { type_string += "IBAS " }
        if (file_type & 0x02) != 0 { type_string += "ABAS " }
        if (file_type & 0x04) != 0 { type_string += "BIN " }
        if (file_type & 0x08) != 0 { type_string += "SEQ " }
        if (file_type & 0x10) != 0 { type_string += "REL " }
        if (file_type & 0x20) != 0 { type_string += "A " }
        if (file_type & 0x40) != 0 { type_string += "B " }
      }

      if locked { type_string += "LKD " }

      // Track and sector where the file descriptor is.
      fmt.Printf("%02x/%02x", apple2_disk.data[offset + i + 0],
                              apple2_disk.data[offset + i + 1])
      fmt.Print(type_string)

      if apple2_disk.data[offset + i + 0] == 0xff { fmt.Print(" DEL") }

      // Count of sectors used for this file.
      sectors := GetInt16(apple2_disk.data, offset + i + 0x21)

      fmt.Printf("sectors: %d", sectors)

      fmt.Println()
    }

    // Next sector in the linked list of Catalog entries.
    track = int(apple2_disk.data[offset + 1])
    sector = int(apple2_disk.data[offset + 2])

    if track == 0 { break }
  }
}



