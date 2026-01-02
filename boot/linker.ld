/* boot/linker.ld
 * Linker script for the kernel image.
 *
 * Flow:
 * - Set the ELF entry point to `_start` (the symbol exported by boot/boot.s).
 * - Place the binary so GRUB loads it at physical address 1 MiB.
 * - Emit the Multiboot header first and contiguous so GRUB can find it.
 * - Lay out text/rodata/data/bss in standard order; keep bss zero-initialized.
 */

ENTRY(_start)

SECTIONS
{
  /* Physical load address for the kernel image (common for Multiboot kernels) */
  . = 1M;

  /* Multiboot header must be near the start so GRUB can locate it */
  /* Executable code */

  .text BLOCK(4K) : ALIGN(4K)
    {
        *(.multiboot)
        *(.text)
    }

  /* Read-only data (const tables, strings, etc.) */
  .rodata BLOCK(4K) : ALIGN(4K)
  {
    *(.rodata*)
  }

  /* Writable data */
  .data BLOCK(4K) : ALIGN(4K)
    {
        *(.data)
    }

  /* Zero-initialized data and common symbols; stays zeroed at load time */
    .bss BLOCK(4K) : ALIGN(4K)
    {
        *(COMMON)
        *(.bss)
        *(.bootstrap_stack)
    }

  __kernel_end = .;
}
