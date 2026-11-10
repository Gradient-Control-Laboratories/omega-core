"""``gsl``: the bundled program, run with this command's arguments as given."""

from __future__ import annotations

import os
import subprocess
import sys

from omega_gsl import executable


def main() -> None:
    program = str(executable())
    if sys.platform == "win32":
        sys.exit(subprocess.call([program, *sys.argv[1:]]))
    os.execv(program, [program, *sys.argv[1:]])


if __name__ == "__main__":
    main()
