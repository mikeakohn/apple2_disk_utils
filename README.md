# apple2_disk_utils
Utilities for reading from Apple II disks.

August 23, 2015:
I was playing around with an Apple IIgs emulator today with the hopes of writing some code
with naken_asm (and eventually maybe Java Grinder) for Apple IIgs.  Using this project and a
small piece of code I didn't check in yet, I was able to update the HELLO program on some
Apple IIe .dsk image I downloaded and got my 20 byte program running.

Currently the only features supported are dumping disk info, the catalog, and dumping a file
off the Apple II disk to the local file system (or stdout for text files).

My hope is to eventually be able to format a blank disk with this and write my own software
on the disk.

I used Google Go (GoLang) for this project just for the heck of it.  Trying to learn something
new.
