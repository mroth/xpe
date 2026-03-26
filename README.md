Package xpe contains an experimental Go module for determining the number of
performance and efficiency cores ("P-Cores" and "E-Cores") on the runtime CPU
architecture.

CAVEATS: Currently supports Darwin (macOS) on Apple Silicon, Windows on 12th Gen
("Alder Lake") and later Intel processors, and Windows on ARM64 (for example,
Snapdragon-based systems). Linux and other platforms should be possible to support,
but I don't currently have the hardware to test this. Collaborators desired!
