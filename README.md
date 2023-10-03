# SiePatcher
Patcher and flasher for *nix systems, like V_Klay for Windows.
## Supported phones
 * R65 (SGOLD)
 * X75 (SGOLD2)
 * Azq2's emulator: https://github.com/Azq2/pmb887x-emu

## Features
 * Booting phone in Service Mode
 * Reading flash
 * Writing flash
 * Applying patches

## Dependencies
 * Go 1.20 or newer. Although it may work with more or less recent Go version.
 * Azq's `pmb887x-emu` for boot/chaos_x85.bin.

## Installation
 ```
 for APP in chaosloader servicemode; do (cd cmd/$APP && go build); done
 ```

 ## Examples
 ### Boot in Service Mode

 ```
 cmd/servicemode/servicemode -serial /dev/cu.usbserial-110
 ```

 ### Read 1024 bytes of flash from address 0xA1000000

```
cmd/chaosloader/chaosloader -serial /dev/cu.usbserial-110 -loader /path/to/pmb887x-dev/boot/chaos_x85.bin -read_flash -base_addr 0xA1000000 -length 1024 -flash_file /tmp/flash.bin
```

### Write flash
Writing flash is only supported when aligned on erase block boundary and exacly erase block boundary in size.

```
cmd/chaosloader/chaosloader -serial /dev/cu.usbserial-110 -loader /path/to/pmb887x-dev/boot/chaos_x85.bin -write_flash -base_addr 0xA1000000 -length 1024 -flash_file /tmp/flash.bin
```

### Apply patch
Note that reverting patches is currently NOT supported ;-)

```
cmd/chaosloader/chaosloader -serial /dev/cu.usbserial-110 -loader /path/to/pmb887x-dev/boot/chaos_x85.bin -apply_patch -patch_file ~/Downloads/SL75v52_Work_without_SIM_card.vkp
```

### Working with emulator instead of a real phone
The same commands above will work with emulator if you supply a command-line flash `-emulator`. SiePatcher will wait for the emulator to start and connect to `/tmp/siemens.sock`