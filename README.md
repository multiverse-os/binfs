[<img src="https://avatars2.githubusercontent.com/u/24763891?s=400&u=c1150e7da5667f47159d433d8e49dad99a364f5f&v=4"  width="256px" height="256px" align="right" alt="Multiverse OS Logo">](https://github.com/multiverse-os)

## Multiverse OS: `binfs` Runtime Executable Virtual Filesystem
**URL** [multiverse-os.org](https://multiverse-os.org)

The second project to be developed from the [`singularity` experimental
library](https://github.com/multiverse-os/singularity) enabling runtime storage
of files in the current runtime executable through live self udpates. The
filesystem uses two magic sequences: `BHS` for the header section and `BFS` for
the file section. The headers are made up of a 16 byte stringname, 8 byte
offset, 8 byte size, and 32 byte checksum for a total of 64 bits. Allowing the
number of files to be easily calculated from the size of the header section
between the delimeters `BHS` and `BFS`. 

Currently the Virtual Filesystem is very limited, its essentially a map and the
files are accessed via filenames. But future updates will focus on extending the
usability of the filesystem, and offering the ability to cache it to restricted
tempFS drive mounted after launch, keeping everything in memory. This will be
paired with the [`memexec` library](https://github.com/multiverse-os/memexec) to
embed binaries on-the-fly, update them without requiring binary destribution, or
function as a update system for other Multiverse OS applications. 


