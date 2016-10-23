# apple2_disk_utils
Utilities for reading from Apple II disks.

August 23, 2015:
I was playing around with an Apple IIgs emulator today with the hopes of writing some code
with naken_asm (and eventually maybe Java Grinder) for Apple IIgs.  Using this project and a
small piece of code I didn't check in yet, I was able to update the HELLO program on some
Apple IIe .dsk image I downloaded and got my 20 byte program running.

I used Google Go (GoLang) for this project just for the heck of it.  Trying to learn something
new.

The programs included here are:

read_dos33_disk.go
------------------

Read an Apple II disk file and dump disk info, the free sector bitmap,
and the "catalog" of files on the disk.

Example:

go run read_dos33_disk.go apple_iigs_java_demo.dsk

read_dos33_sector.go
--------------------

Read an Apple II disk and dump the given 255 byte track / sector.

Example:

go run read_dos33_sector.go apple_iigs_java_demo.dsk 0 12

init_dos33_disk.go
------------------

Create a new formatted Apple II disk image and write the specified
program to autorun at boot time.

Example:

go run init_dos33_disk.go apple_iigs_java_demo.dsk dos33.img apple_iigs_java_demo.bin

Note: The dos33.img is the first 8448 of any bootable disk image.  I didn't include one
in this repository since I didn't know if it was copyrighted.  It's possible that the
image doesn't need to be this big, but for whatever reason when I wrote this program I
used 8448 bytes of a working disk.

