# SiePatcher
Patcher and flasher for *nix systems, like V_Klay for Windows.
## Supported phones
 * R65 (SGOLD)
 * X75 (SGOLD2)
 * Azq2's emulator: https://github.com/Azq2/pmb887x-emu
 * Fullflash dumps

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

```
cmd/chaosloader/chaosloader -serial /dev/cu.usbserial-110 -loader /path/to/pmb887x-dev/boot/chaos_x85.bin -apply_patch -patch_file ~/Downloads/SL75v52_Work_without_SIM_card.vkp
```

### Revert patch
See a previous example, but specify `-revert_patch` instead of `-apply_patch`

### Test if a patch can be applied cleanly
See a previous example, specify `-dry_run` in addition to `-apply_patch` or `-revert_patch`.

### Working with emulator instead of a real phone
The same commands above will work with emulator if you supply a command-line flag `-emulator`. SiePatcher will wait for the emulator to start and connect to `/tmp/siemens.sock`

### Working with the fullflash file instead of a real phone
The same commands above will work with the fullflash file if you supply a command-line flag `-use_fullflash_not_phone`.
You must specify a path to the fullflash dump using `-use_fullflash_file_path /path/to/file.bin`.
