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

func (apple2_disk *Apple2Disk) DumpSector(track int, sector int) {
  text := make([]byte, 16)

  offset := GetOffset(track, sector)

  fmt.Printf("===== Track: %d   Sector: %d   Offset: %d =====\n", track, sector, offset)

 for i := 0; i < 256; i++ {
    if (i % 16) == 0 {
      fmt.Printf("%02x: ", i)
    }

    fmt.Printf(" %02x", apple2_disk.data[offset + i])

    ch := int(apple2_disk.data[offset + i])
    ch &= 0x7f

    if ch >= 32 && ch < 127 {
      text[i % 16] = byte(ch)
    } else {
      text[i % 16] = '.'
    }

    if (i % 16) == 15 {
      fmt.Print("  ")
      for n := 0; n < 16; n++ {
        fmt.Printf("%c", text[n])
      }
      fmt.Println()
    }
  }

  fmt.Println()
}

func (apple2_disk *Apple2Disk) FindFile(filename string) (int, int, bool) {
  if len(filename) > 30 { return 0, 0, false }

  apple_name := make([]byte, 30)

  for i := 0; i < len(apple_name); i++ { apple_name[i] = 32 | 0x80 }
  for i := 0; i < len(filename); i++ {
    apple_name[i] = filename[i] | 0x80
  }

  offset := GetOffset(17, 0)

  track := int(apple2_disk.data[offset + 1])
  sector := int(apple2_disk.data[offset + 2])

  for true {
    //fmt.Printf("Track: %d  Sector: %d\n", track, sector)

    offset := GetOffset(track, sector)

    for i := 0x0b; i < 256; i += 0x23 {
      var n int

      for n = 0; n < len(apple_name); n++ {
        //fmt.Printf("%d: %d %d\n", i, apple2_disk.data[offset + i + n + 3], apple_name[n])
        if apple2_disk.data[offset + i + n + 3] != apple_name[n] { break }
      }

      if n == len(apple_name) {
        is_binary := true
        if apple2_disk.data[offset + i + 2] & 0x7f == 0 { is_binary = false }
        return int(apple2_disk.data[offset + i + 0]), int(apple2_disk.data[offset + i + 1]), is_binary
      }
    }
    track = int(apple2_disk.data[offset + 1])
    sector = int(apple2_disk.data[offset + 2])

    if track == 0 { break }
  }

  return 0, 0, false
}

func (apple2_disk *Apple2Disk) PrintFileSectorList(track int, sector int) {

  for true {
    offset := GetOffset(track, sector)

    fmt.Printf("======= File Track/Sector List %d/%d  ========\n", track, sector)

    fmt.Printf("           Track Next: %d\n", apple2_disk.data[offset + 0x01])
    fmt.Printf("          Sector Next: %d\n", apple2_disk.data[offset + 0x02])
    fmt.Printf("Sector Offset In File: %d\n", GetInt16(apple2_disk.data, offset + 0x02))

    for i := 0x0c; i <= 0xff; i += 2 {
      if apple2_disk.data[offset + i] == 0 &&
         apple2_disk.data[offset + i + 1] == 0 { break }

      fmt.Printf("                 Data: track=%d sector=%d\n",
        apple2_disk.data[offset + i], apple2_disk.data[offset + i + 1])
    }

    if apple2_disk.data[offset + 0x01] == 0x00 { break }

    track = int(apple2_disk.data[offset + 0x01])
    sector = int(apple2_disk.data[offset + 0x02])
  }
}

func (apple2_disk *Apple2Disk) DumpBinaryFile(track int, sector int) {
  binary_data := make([]byte, 0)

  for true {
    offset := GetOffset(track, sector)

    for i := 0x0c; i <= 0xff; i += 2 {
      if apple2_disk.data[offset + i] == 0 &&
         apple2_disk.data[offset + i + 1] == 0 { break }

      file_track := int(apple2_disk.data[offset + i])
      file_sector := int(apple2_disk.data[offset + i + 1])

      //dump_sector(apple2_disk.data, file_track, file_sector)
      bin_offset := GetOffset(file_track, file_sector)

      binary_data = append(binary_data, apple2_disk.data[bin_offset:bin_offset + 256]...)
    }

    track = int(apple2_disk.data[offset + 0x01])
    sector = int(apple2_disk.data[offset + 0x02])

    if track == 0 && sector == 0 { break }
  }

  file_out, err := os.Create("out.bin")

  if err != nil {
    panic(err)
  }

  load_offset := GetInt16(binary_data, 0)
  length := GetInt16(binary_data, 2)

  fmt.Printf("      Load Address: 0x%04x\n", load_offset)
  fmt.Printf("            Length: 0x%04x\n", length)

  file_out.Write(binary_data[4:length + 4])
  file_out.Close()
}

func (apple2_disk *Apple2Disk) PrintTextFile(track int, sector int) {
  for true {
    offset := GetOffset(track, sector)

    //fmt.Printf("Sector Offset In File: %d\n",
    //  int(apple2_disk.data[offset + 0x02]) | (int(apple2_disk.data[offset + 0x02] << 8)))

    for i := 0x0c; i <= 0xff; i += 2 {
      if apple2_disk.data[offset + i] == 0 && apple2_disk.data[offset + i + 1] == 0 { break }

      file_track := int(apple2_disk.data[offset + i])
      file_sector := int(apple2_disk.data[offset + i + 1])

      //dump_sector(apple2_disk.data, file_track, file_sector)
      text_offset := GetOffset(file_track, file_sector)
      for n := 0; n < 256; n++ {
        ch := apple2_disk.data[text_offset + n]

        if (ch == 0) { break; }
        ch &= 0x7f

        if (ch >= 32 && ch <= 127 || ch == '\r' || ch == '\n' || ch == '\t') {
          //fmt.Printf("[%d]%c", n, ch);
          fmt.Printf("%c", ch)
          if ch == '\r' { fmt.Printf("\n") }
        } else {
          fmt.Print("[%02x]", ch)
        }
      }
    }

    track = int(apple2_disk.data[offset + 0x01])
    sector = int(apple2_disk.data[offset + 0x02])

    if track == 0 && sector == 0 { break }
  }

  fmt.Println()
}



