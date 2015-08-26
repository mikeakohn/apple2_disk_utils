package apple2_disk

import "fmt"
import "strings"
import "os"

// http://fileformats.archiveteam.org/wiki/Apple_DOS_file_system

type Apple2Disk struct {
  data []byte
  offset_to_disk_info int
  catalog_track int
  catalog_sector int
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

  // Track 17, Sector 0 should be the disk info area.  At least on 5 1/4"
  // disks.
  apple2_disk.offset_to_disk_info = GetOffset(17, 0)

  apple2_disk.catalog_track = int(apple2_disk.data[apple2_disk.offset_to_disk_info + 1])
  apple2_disk.catalog_sector = int(apple2_disk.data[apple2_disk.offset_to_disk_info + 2])

  return true
}

func (apple2_disk *Apple2Disk) PrintDiskInfo() {

  offset := apple2_disk.offset_to_disk_info

  total_tracks := int(apple2_disk.data[offset + 0x34])

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
  fmt.Printf("  Free Sector Bitmap:\n")

  for i := 0x38; i <= 0xff; i += 4 {
    if  (i - 0x38) / 4 >= total_tracks { break }
    fmt.Printf("         Track %d: %02x %02x %02x %02x\n",
      (i - 0x36) / 4,
      apple2_disk.data[offset + i + 0],
      apple2_disk.data[offset + i + 1],
      apple2_disk.data[offset + i + 2],
      apple2_disk.data[offset + i + 3])
  }
}

func (apple2_disk *Apple2Disk) PrintCatalog() {
  track := apple2_disk.catalog_track
  sector := apple2_disk.catalog_sector

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

  track := apple2_disk.catalog_track
  sector := apple2_disk.catalog_sector

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

func (apple2_disk *Apple2Disk) DumpBinaryFile(output_name string, track int, sector int) {
  binary_data := make([]byte, 0)

  for true {
    offset := GetOffset(track, sector)

    for i := 0x0c; i <= 0xff; i += 2 {
      if apple2_disk.data[offset + i] == 0 &&
         apple2_disk.data[offset + i + 1] == 0 { break }

      file_track := int(apple2_disk.data[offset + i])
      file_sector := int(apple2_disk.data[offset + i + 1])

      //apple2_disk.DumpSector(file_track, file_sector)
      bin_offset := GetOffset(file_track, file_sector)

      binary_data = append(binary_data, apple2_disk.data[bin_offset:bin_offset + 256]...)
    }

    track = int(apple2_disk.data[offset + 0x01])
    sector = int(apple2_disk.data[offset + 0x02])

    if track == 0 && sector == 0 { break }
  }

  file_out, err := os.Create(output_name)

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

func (apple2_disk *Apple2Disk) PrintTextFile(output_name string, track int, sector int) {

  file_out, err := os.Create(output_name)

  if err != nil {
    panic(err)
  }

  for true {
    offset := GetOffset(track, sector)

    for i := 0x0c; i <= 0xff; i += 2 {
      if apple2_disk.data[offset + i] == 0 && apple2_disk.data[offset + i + 1] == 0 { break }

      file_track := int(apple2_disk.data[offset + i])
      file_sector := int(apple2_disk.data[offset + i + 1])

      //apple2_disk.DumpSector(file_track, file_sector)
      text_offset := GetOffset(file_track, file_sector)
      for n := 0; n < 256; n++ {
        ch := apple2_disk.data[text_offset + n]

        if (ch == 0) { break; }
        ch &= 0x7f

        if (ch >= 32 && ch <= 127 || ch == '\r' || ch == '\n' || ch == '\t') {
          file_out.WriteString(fmt.Sprintf("%c", ch))
          if ch == '\r' { file_out.WriteString("\n") }
        } else {
          file_out.WriteString(fmt.Sprintf("[%02x]", ch))
        }
      }
    }

    track = int(apple2_disk.data[offset + 0x01])
    sector = int(apple2_disk.data[offset + 0x02])

    if track == 0 && sector == 0 { break }
  }

  file_out.WriteString("\n")
}

func (apple2_disk *Apple2Disk) DumpFile(filename string, output_name string) {
  track, sector, is_binary := apple2_disk.FindFile(filename)

  if track == 0 && sector == 0 {
    fmt.Println("Error: File not found '" + filename + "'")
  } else {
    //apple2_disk.DumpSector(track, sector)
    apple2_disk.PrintFileSectorList(track, sector)

    if is_binary {
      apple2_disk.DumpBinaryFile(output_name, track, sector)
    } else {
      apple2_disk.PrintTextFile(output_name, track, sector)
    }
  }
}

func (apple2_disk *Apple2Disk) MarkSectorUsed(track int, sector int) {
  offset := apple2_disk.offset_to_disk_info

  byte_pair_offset := offset + 0x38 + (track * 4)
  current := (int(apple2_disk.data[byte_pair_offset]) << 8) |
              int(apple2_disk.data[byte_pair_offset + 1])

  current &= 0xffff ^ (1 << uint32(sector))

  apple2_disk.data[byte_pair_offset + 0] = byte(current >> 8)
  apple2_disk.data[byte_pair_offset + 1] = byte(current & 0xff)
}

func (apple2_disk *Apple2Disk) MarkSectorFree(track int, sector int) {
  offset := GetOffset(17, 0)

  byte_pair_offset := offset + 0x38 + (track * 4)
  current := (int(apple2_disk.data[byte_pair_offset]) << 8) |
              int(apple2_disk.data[byte_pair_offset + 1])

  current |= 1 << uint32(sector)

  apple2_disk.data[byte_pair_offset + 0] = byte(current >> 8)
  apple2_disk.data[byte_pair_offset + 1] = byte(current & 0xff)
}

func (apple2_disk *Apple2Disk) IsSectorFree(track int, sector int) bool {
  offset := apple2_disk.offset_to_disk_info

  byte_pair_offset := offset + 0x38 + (track * 4)
  current := (int(apple2_disk.data[byte_pair_offset]) << 8) |
              int(apple2_disk.data[byte_pair_offset + 1])

  if (current & (1 << uint32(sector))) == 0 {
    return false
  } else {
    return true
  }
}

func (apple2_disk *Apple2Disk) Init() {
  apple2_disk.data = make([]byte, 143360)

  disk_info := [...]byte{
    0,       // unused
    17, 15,  // track/sector
    3,       // release number
    0, 0,    // unused
    254,     // volume number
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
    0, 0, 0, 0, 0, 0, 0, 0, 0, 0 ,0, 0,
    122,     // max number of track/sector pairs
    0, 0, 0, 0, 0, 0, 0, 0,
    18,      // last track sectors were allocated
    1,       // direction of track allocation
    0, 0,
    35, 16,  // max tracks, max sectors per track
  }

  offset := apple2_disk.offset_to_disk_info

  for i := 0; i < len(disk_info); i++ {
    apple2_disk.data[offset + i] = disk_info[i]
  }

  for i := 0x38; i < 0xc4; i += 4 {
    apple2_disk.data[offset + i + 0] = 0xff
    apple2_disk.data[offset + i + 1] = 0xff
  }

  apple2_disk.MarkSectorUsed(17, 0)

  // Link catalog
  for sector := 15; sector >= 1; sector-- {
    offset := GetOffset(17, sector)

    if sector != 1 {
      apple2_disk.data[offset + 1 ] = 17
      apple2_disk.data[offset + 2 ] = byte(sector) - 1
    }

    apple2_disk.MarkSectorUsed(17, sector)
  }
}

func (apple2_disk *Apple2Disk) AddDos(filename string) {
  file, err := os.Open(filename)

  if err != nil {
    panic(err)
  }

  defer file.Close()

  stat, err := file.Stat()
  if err != nil {
    panic(err)
  }

  if (stat.Size() & 0xff) != 0 {
    fmt.Println("Error: dos33.img is not a multiple of 256 bytes")
    return
  }

  dos33 := make([]byte, stat.Size())
  file.Read(dos33)

  for i := 0; i < len(dos33); i++ {
    apple2_disk.data[i] = dos33[i]
  }

  track := 0
  sector := 0

  for i := 0; i < len(dos33) / 256; i++ {
    apple2_disk.MarkSectorUsed(track, sector)
    sector++
    if sector == 16 { sector = 0; track++ }
  }
}

func (apple2_disk *Apple2Disk) AllocSector() (int, int) {
  offset := apple2_disk.offset_to_disk_info

  track := int(apple2_disk.data[offset + 0x30])
  sector := 0

  for count := 0; count < 35 * 16; count++ {
    if apple2_disk.IsSectorFree(track, sector) {
      apple2_disk.data[offset + 0x30] = byte(track)
      apple2_disk.MarkSectorUsed(track, sector)
      return track, sector
    }

    sector++
    if sector == 16 {
      sector = 0
      track++
      if track == 35 {
        track = 0
      }
    }
  }

  return -1, -1
}

func (apple2_disk *Apple2Disk) AddFile(filename string, apple_name string, address int) {

  if len(apple_name) > 30 {
    fmt.Println("Error: Filename can't be more then 30 chars.\n")
    return
  }

  track := apple2_disk.catalog_track
  sector := apple2_disk.catalog_sector

  // Search catalog for an empty space
  for true {
    offset := GetOffset(track, sector)

    // For each possible catalog entry.
    for i := 0x0b; i < 256; i += 0x23 {
      // If Catalog entry is empty.
      if apple2_disk.data[offset + i + 0] == 0xff ||
         apple2_disk.data[offset + i + 0] == 0x00 &&
         apple2_disk.data[offset + i + 1] == 0x00 {
        if address != 0 { apple2_disk.data[offset + i + 2] = 0x04 }
        for n := 0; n < 30; n++ { apple2_disk.data[offset + i + 3 + n ] = 0xa0 }
        apple_name = strings.ToUpper(apple_name)
        for n := 0; n < len(apple_name); n++ {
          apple2_disk.data[offset + i + 3 + n ] = apple_name[n] | 0x80
        }

        file, err := os.Open(filename)

        if err != nil {
          panic(err)
        }

        defer file.Close()

        stat, err := file.Stat()
        if err != nil {
          panic(err)
        }

        binfile := make([]byte, stat.Size())
        file.Read(binfile)

        if address != 0 {
          binsize := stat.Size()
          binfile = append([]byte{
            byte(address & 0xff), byte(address >> 8),
            byte(binsize & 0xff), byte(binsize >> 8), }, binfile...)
        }

        sectors := (len(binfile) + 255) / 256

        apple2_disk.data[offset + i + 0x21 ] = byte(sectors & 0xff)
        apple2_disk.data[offset + i + 0x22 ] = byte(sectors >> 8)

        // Allocate sector for Track/Sector List
        track, sector = apple2_disk.AllocSector()

        // Save track and sector pointer in catalog sector
        apple2_disk.data[offset + i + 0x00 ] = byte(track)
        apple2_disk.data[offset + i + 0x01 ] = byte(sector)

        offset = GetOffset(track, sector)
        entry := 0x0c

        for i := 0; i < len(binfile); i += 256 {
          track, sector = apple2_disk.AllocSector()
          file_offset := GetOffset(track, sector)

          apple2_disk.data[offset + entry + 0] = byte(track)
          apple2_disk.data[offset + entry + 1] = byte(sector)

          length := len(binfile) - i
          if length > 256 { length = 256 }

          for n := 0; n < length; n++ {
            apple2_disk.data[file_offset + n] = binfile[i + n]
          }
        }

        //fmt.Printf("Alloc() track=%d sector=%d\n", track, sector

        return
      }
    }

    // Next sector in the linked list of Catalog entries.
    track = int(apple2_disk.data[offset + 1])
    sector = int(apple2_disk.data[offset + 2])

    if track == 0 { break }
  }

  fmt.Println("Disk is full\n")
}

func (apple2_disk *Apple2Disk) Save(filename string) {
  file_out, err := os.Create(filename)

  if err != nil {
    panic(err)
  }

  file_out.Write(apple2_disk.data)
  file_out.Close()
}



